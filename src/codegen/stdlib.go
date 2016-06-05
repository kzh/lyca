package codegen

import (
//    "log"
//    "fmt"

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
        PRIMITIVE_TYPES["char"],
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
