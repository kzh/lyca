package codegen

import (
//    "log"

    "llvm.org/llvm/bindings/go/llvm"
    "github.com/furryfaust/lyca/src/parser"
)

var mangleFuncs map[string]int = map[string]int {
    "malloc": 0,
}

func (c *Codegen) mangle(name string) string {
    if _, ok := mangleFuncs[name]; ok {
        name = "--" + name
    }

    return name
}

func (c *Codegen) injectStdLib() {
    c.declareMemcpy();

    c.stdString();
}

func (c *Codegen) declareMemcpy() {
    t := llvm.FunctionType(llvm.VoidType(), []llvm.Type{
        llvm.PointerType(PRIMITIVE_TYPES["char"], 0),
        llvm.PointerType(PRIMITIVE_TYPES["char"], 0),
        PRIMITIVE_TYPES["int"],
        PRIMITIVE_TYPES["int"],
        PRIMITIVE_TYPES["boolean"],
    }, false)
    llvm.AddFunction(c.module, "llvm.memcpy.p0i8.p0i8.i32", t)
}

func (c *Codegen) stdString() {
    tmpl := &Template{
        Type: llvm.GlobalContext().StructCreateNamed("string"),
        Variables: map[string]int{},
    }
    c.templates["string"] = tmpl

    vars := []llvm.Type{
        llvm.PointerType(PRIMITIVE_TYPES["char"], 0),
        PRIMITIVE_TYPES["int"],
        PRIMITIVE_TYPES["int"],
    }
    tmpl.Type.StructSetBody(vars, false)

    lenFuncType := llvm.FunctionType(PRIMITIVE_TYPES["int"], []llvm.Type{llvm.PointerType(tmpl.Type, 0)}, false)
    lenFunc := llvm.AddFunction(c.module, "-string-len", lenFuncType)
    lenFunc.Param(0).SetName("this")
    block := llvm.AddBasicBlock(c.module.NamedFunction("-string-len"), "entry")
    c.functions["-string-len"] = block
    c.currFunc = "-string-len"
    c.builder.SetInsertPoint(block, block.LastInstruction())
    ret := c.builder.CreateStructGEP(c.getCurrParam("this"), 1, "")
    ret = c.builder.CreateLoad(ret, "")
    ret = c.builder.CreateSub(ret, llvm.ConstInt(PRIMITIVE_TYPES["int"], 1, false), "")
    c.builder.CreateRet(ret)

    printFuncType := llvm.FunctionType(PRIMITIVE_TYPES["int"], []llvm.Type{
        llvm.PointerType(PRIMITIVE_TYPES["char"], 0),
    }, true)
    llvm.AddFunction(c.module, "printf", printFuncType)
}

func (c *Codegen) generateStringLiteral(n *parser.StringLitNode) llvm.Value {
    vals := []llvm.Value{}

    for i := 0; i != len(n.Value); i++ {
        char := &parser.CharLitNode{Value: rune(n.Value[i])}
        vals = append(vals, c.generateExpression(char))
    }
    vals = append(vals, c.generateExpression(&parser.CharLitNode{Value: 0}))
    arr := llvm.ConstArray(PRIMITIVE_TYPES["char"], vals)
    chars := c.builder.CreateMalloc(arr.Type(), "")
    c.builder.CreateStore(arr, chars)

    str := c.builder.CreateMalloc(c.templates["string"].Type, "")
    chars = c.builder.CreateBitCast(chars, llvm.PointerType(PRIMITIVE_TYPES["char"], 0), "")

    c.builder.CreateStore(chars, c.builder.CreateStructGEP(str, 0, ""))
    c.builder.CreateStore(llvm.ConstInt(PRIMITIVE_TYPES["int"], uint64(len(vals)), false), c.builder.CreateStructGEP(str, 1, ""))
    c.builder.CreateStore(llvm.ConstInt(PRIMITIVE_TYPES["int"], uint64(len(vals)), false), c.builder.CreateStructGEP(str, 2, ""))

    return str
}

func (c *Codegen) generateStringConcat(str1, str2 llvm.Value) llvm.Value {
//    one := llvm.ConstInt(PRIMITIVE_TYPES["int"], 1, false)

    len1      := c.builder.CreateCall(c.module.NamedFunction("-string-len"), []llvm.Value{str1}, "")
    len2      := c.builder.CreateLoad(c.builder.CreateStructGEP(str2, 1, ""), "")
    len_sum   := c.builder.CreateAdd(len1, len2, "")

    chars := c.builder.CreateCall(c.module.NamedFunction("malloc"), []llvm.Value{len_sum}, "")
    c.builder.CreateCall(c.module.NamedFunction("llvm.memcpy.p0i8.p0i8.i32"), []llvm.Value{
        chars, c.unbox(str1), len1,
        llvm.ConstInt(PRIMITIVE_TYPES["int"], 0, false),
        llvm.ConstInt(PRIMITIVE_TYPES["boolean"], 0, false),
    }, "")
    c.builder.CreateCall(c.module.NamedFunction("llvm.memcpy.p0i8.p0i8.i32"), []llvm.Value{
        c.builder.CreateGEP(chars, []llvm.Value{len1}, ""), c.unbox(str2), len2,
        llvm.ConstInt(PRIMITIVE_TYPES["int"], 0, false),
        llvm.ConstInt(PRIMITIVE_TYPES["boolean"], 0, false),
    }, "")

    str := c.builder.CreateMalloc(c.templates["string"].Type, "")
    c.builder.CreateStore(chars, c.builder.CreateStructGEP(str, 0, ""))
    c.builder.CreateStore(len_sum, c.builder.CreateStructGEP(str, 1, ""))
    c.builder.CreateStore(len_sum, c.builder.CreateStructGEP(str, 2, ""))

    return str
}
