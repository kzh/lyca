package codegen

import (
    "log"

    "llvm.org/llvm/bindings/go/llvm"
    "github.com/furryfaust/lyca/src/parser"
)

var PRIMITIVE_TYPES map[string]llvm.Type = map[string]llvm.Type {
    "int": llvm.Int32Type(), "char": llvm.Int1Type(), "float": llvm.FloatType(),
}

type Template struct {
    Type llvm.Type
    Variables map[string]int
    Values []llvm.Value
}

type Codegen struct {
    tree *parser.AST
    scope *Scope

    module llvm.Module
    builder llvm.Builder

    templates map[string]*Template
    functions map[string]llvm.BasicBlock
}

func Generate(tree *parser.AST) {
    gen := &Codegen{
        tree: tree,
        scope: &Scope{},

        module: llvm.NewModule("main"),
        builder: llvm.NewBuilder(),

        templates: map[string]*Template{},
        functions: map[string]llvm.BasicBlock{},
    }
    gen.declareTopLevelNodes()

    if ok := llvm.VerifyModule(gen.module, llvm.ReturnStatusAction); ok != nil {
        log.Println(ok.Error())
    }
    gen.module.Dump()
}

func (c *Codegen) enterScope() {
    s := c.scope.AddScope()
    c.scope = s
}

func (c *Codegen) exitScope() {
    c.scope = c.scope.Outer
}

func (c *Codegen) declareTopLevelNodes() {
    for _, node := range c.tree.Nodes {
        switch n := node.(type) {
        case *parser.FuncDeclNode:
            c.declareFunc(n, llvm.VoidType())
        case *parser.TemplateNode:
            c.declareTemplate(n)
        }
    }
}

func (c *Codegen) declareFunc(n *parser.FuncDeclNode, obj llvm.Type) {
    sig := n.Function.Signature
    name := sig.Name.Value
    f := c.getLLVMFuncType(sig.Return, sig.Parameters, obj)
    llvm.AddFunction(c.module, name, f)
    block := llvm.AddBasicBlock(c.module.NamedFunction(name),"entry")

    c.functions[sig.Name.Value] = block
}

func (c *Codegen) declareTemplate(n *parser.TemplateNode) {
    c.templates[n.Name.Value] = &Template{Variables: map[string]int{}}

    var vars []llvm.Type
    for i, v := range n.Variables {
        vars = append(vars, c.getLLVMType(v.Type))
        c.templates[n.Name.Value].Variables[v.Name.Value] = i
    }

    tmpl := llvm.StructType(vars, false)
    c.templates[n.Name.Value].Type = tmpl

    if n.Constructor != nil {
        f := &parser.FuncDeclNode{
            Function: &parser.FuncNode{
                Signature: &parser.FuncSignatureNode{
                    Name: parser.Identifier{Value: "-" + n.Name.Value},
                    Parameters: n.Constructor.Parameters,
                },
            },
        }
        c.declareFunc(f, tmpl)
    }

    for _, meth := range n.Methods {
        name := "-" + n.Name.Value + "-" + meth.Function.Signature.Name.Value
        meth.Function.Signature.Name = parser.Identifier{Value: name}
        c.declareFunc(meth, tmpl)
    }
}

func (c *Codegen) generateTopLevelNodes() {
    /*
    for _, node := range c.tree.Nodes {
        switch n := node.(type) {
        case *parser.TemplateNode:
        case *parser.FuncDeclNode:
        case *parser.VarDeclNode:
        }
    }
    */
}

func (c *Codegen) generateTemplateDecl(node *parser.TemplateNode) {

}

func (c *Codegen) generateVarDecl(node *parser.VarDeclNode, top bool) {
    /*
    t := c.getType(node.Type)
    name := node.Name.Value
    */
}

/*
func (c *Codegen) generateExpression(node parser.Node) llvm.Value {
}
*/

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
            return t.Type
        }
    }

    return llvm.VoidType()
}
