package codegen

import (
//    "log"

    "llvm.org/llvm/bindings/go/llvm"
    "github.com/furryfaust/lyca/src/parser"
)

func (c *Codegen) injectStdLib() {
    c.stdString();
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

    toCFuncType := llvm.FunctionType(llvm.PointerType(PRIMITIVE_TYPES["char"], 0), []llvm.Type{
        llvm.PointerType(tmpl.Type, 0),
    }, false)
    toCFunc := llvm.AddFunction(c.module, "-string-toCStr", toCFuncType)
    toCFunc.Param(0).SetName("this")
    block := llvm.AddBasicBlock(c.module.NamedFunction("-string-toCStr"), "entry")
    c.functions["-string-toCStr"] = block
    c.currFunc = "-string-toCStr"
    c.builder.SetInsertPoint(block, block.LastInstruction())
    ret := c.builder.CreateStructGEP(c.getCurrParam("this"), 0, "")
    ret = c.builder.CreateLoad(ret, "")
    c.builder.CreateRet(ret)

    printFuncType := llvm.FunctionType(PRIMITIVE_TYPES["int"], []llvm.Type{
        llvm.PointerType(PRIMITIVE_TYPES["char"], 0),
    }, true)
    printFunc := llvm.AddFunction(c.module, "printf", printFuncType)
    printFunc.SetLinkage(llvm.ExternalLinkage)
    printFunc.SetFunctionCallConv(llvm.CCallConv)
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
    c.builder.CreateStore(llvm.ConstInt(PRIMITIVE_TYPES["int"], uint64(len(vals) * 2), false), c.builder.CreateStructGEP(str, 2, ""))

    return str
}
