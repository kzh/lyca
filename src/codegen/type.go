package codegen

import (
    "llvm.org/llvm/bindings/go/llvm"
    "github.com/furryfaust/lyca/src/parser"
)

var PRIMITIVE_TYPES map[string]llvm.Type = map[string]llvm.Type {
    "int": llvm.Int32Type(), "char": llvm.Int8Type(), "float": llvm.FloatType(), "boolean": llvm.Int1Type(),
}

func (c *Codegen) getLLVMFuncType(ret parser.Node, params []*parser.VarDeclNode, obj llvm.Type) llvm.Type {
    p := make([]llvm.Type, 0)
    if obj != llvm.VoidType() {
        p = append(p, obj)
    }

    for _, v := range params {
        p = append(p, c.getLLVMType(v.Type))
    }

    return llvm.FunctionType(c.getLLVMType(ret), p, false)
}

func (c *Codegen) getLLVMType(node parser.Node) llvm.Type {
    switch t := node.(type) {
    /*
    case *FuncTypeNode:
    case *ArrayTypeNode:
    */
    case *parser.NamedTypeNode:
        if prim, ok := PRIMITIVE_TYPES[t.Name.Value]; ok {
            return prim
        } else if t, ok := c.templates[t.Name.Value]; ok {
            return llvm.PointerType(t.Type, 0)
        }
    case *parser.BinaryExprNode:
        return c.getLLVMType(t.Left)
    case *parser.CharLitNode:
        return PRIMITIVE_TYPES["char"]
    case *parser.BoolLitNode:
        return PRIMITIVE_TYPES["boolean"]
    case *parser.NumLitNode:
        if t.IsFloat {
            return PRIMITIVE_TYPES["float"]
        } else {
            return PRIMITIVE_TYPES["int"]
        }
    case *parser.VarAccessNode:
        return c.scope.GetType(t.Name.Value)
    }

    return llvm.VoidType()
}

func (c *Codegen) getLLVMDefaultValue(node parser.Node) llvm.Value {
    llvmType := c.getLLVMType(node)
    switch t := node.(type) {
    /*
    case *FuncTypeNode:
    case *ArrayTypeNode:
    */
    case *parser.NamedTypeNode:
        switch t.Name.Value {
        case "int", "char", "bool":
            return llvm.ConstInt(llvmType, 0, false)
        case "float":
            return llvm.ConstFloat(llvmType, 0)
        default:
            tmpl, ok := c.templates[t.Name.Value]
            if !ok {
                //Error undefined template
            }

            return llvm.ConstNamedStruct(tmpl.Type, []llvm.Value{})
        }
    }

    return llvm.Value{}
}
