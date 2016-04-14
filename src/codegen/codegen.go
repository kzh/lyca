package codegen

import (
    "log"

    "llvm.org/llvm/bindings/go/llvm"
    "github.com/furryfaust/lyca/src/parser"
)

var PRIMITE_TYPES map[string]llvm.Type = map[string]llvm.Type {
}

type Codegen struct {
    tree *parser.AST
    scope *Scope

    module llvm.Module
    builder llvm.Builder

    templates map[string]llvm.Type
    functions map[string]llvm.BasicBlock
}

func Generate(tree *parser.AST) {
    gen := &Codegen{
        tree: tree,
        scope: &Scope{},

        module: llvm.NewModule("main"),
        builder: llvm.NewBuilder(),

        templates: map[string]llvm.Type{},
        functions: map[string]llvm.BasicBlock{},
    }
    gen.declareGlobalNodes()
    gen.generateTopLevelNodes()

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

func (c *Codegen) declareGlobalNodes() {
    for _, node := range c.tree.Nodes {
        switch n := node.(type) {
        case *parser.FuncDeclNode:
            sig := n.Function.(*parser.FuncNode).Signature.(*parser.FuncSignatureNode)
            name := sig.Name.Value
            ret := c.getType(sig.Return)
            params := make([]llvm.Type, 0)
            for _, v := range sig.Parameters {
                t := v.(*parser.VarDeclNode).Type
                params = append(params, c.getType(t))
            }
            f := llvm.FunctionType(ret, params, false)
            llvm.AddFunction(c.module, name, f)
            block := llvm.AddBasicBlock(c.module.NamedFunction(name),"entry")

            c.functions[sig.Name.Value] = block
        }
    }
}

func (c *Codegen) generateTopLevelNodes() {
    for _, node := range c.tree.Nodes {
        switch n := node.(type) {
        case *parser.TemplateNode:
            c.generateTemplate(n)
        case *parser.FuncDeclNode:
            c.generateFuncDecl(n)
        case *parser.VarDeclNode:
            c.generateVarDecl(n, true)
        }
    }
}

func (c *Codegen) generateFuncDecl(node *parser.FuncDeclNode) {

}

func (c *Codegen) generateTemplate(node *parser.TemplateNode) {

}

func (c *Codegen) generateVarDecl(node *parser.VarDeclNode, top bool) {
    /*
    t := c.getType(node.Type)
    name := node.Name.Value
    */
}

func (c *Codegen) getType(node parser.Node) llvm.Type {
    switch t := node.(type) {
    /*
    case *FuncTypeNode:
    case *ArrayTypeNode:
    */
    case *parser.NamedTypeNode:
        switch (t.Name.Value) {
        case "int":
            return llvm.Int32Type()
        case "char":
            return llvm.Int1Type()
        case "float":
            return llvm.FloatType()
        default:
            templ, ok := c.templates[t.Name.Value]
            if !ok {
                //Unknown data type error
            }
            return templ
        }
    }

    return llvm.VoidType()
}
