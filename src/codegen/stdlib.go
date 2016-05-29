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
