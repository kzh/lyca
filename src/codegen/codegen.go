package codegen

import (
    "log"

    "llvm.org/llvm/bindings/go/llvm"
    "github.com/furryfaust/lyca/src/parser"
)

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
        scope: &Scope{variables: map[string]llvm.Value{}},

        module: llvm.NewModule("main"),
        builder: llvm.NewBuilder(),

        templates: map[string]*Template{},
        functions: map[string]llvm.BasicBlock{},
    }
    gen.declareTopLevelNodes()
    gen.generateTopLevelNodes()

    if ok := llvm.VerifyModule(gen.module, llvm.ReturnStatusAction); ok != nil {
        log.Println(ok.Error())
    }
    gen.module.Dump()

    engine, err := llvm.NewExecutionEngine(gen.module)
    if err != nil {
        log.Println(err.Error())
    }

    funcResult := engine.RunFunction(gen.module.NamedFunction("main"), []llvm.GenericValue{})
    log.Println("Output:", funcResult.Int(false))
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
    c.templates[n.Name.Value] = &Template{
        Type: llvm.GlobalContext().StructCreateNamed(n.Name.Value),
        Variables: map[string]int{},
    }

    var vars []llvm.Type
    for i, v := range n.Variables {
        vars = append(vars, c.getLLVMType(v.Type))
        c.templates[n.Name.Value].Variables[v.Name.Value] = i
    }

    c.templates[n.Name.Value].Type.StructSetBody(vars, false)
    pointer := llvm.PointerType(c.templates[n.Name.Value].Type, 0)

    if n.Constructor != nil {
        f := &parser.FuncDeclNode{
            Function: &parser.FuncNode{
                Signature: &parser.FuncSignatureNode{
                    Name: parser.Identifier{Value: "-" + n.Name.Value},
                    Parameters: n.Constructor.Parameters,
                },
            },
        }
        c.declareFunc(f, pointer)
    }

    for _, meth := range n.Methods {
        name := "-" + n.Name.Value + "-" + meth.Function.Signature.Name.Value
        meth.Function.Signature.Name = parser.Identifier{Value: name}
        c.declareFunc(meth, pointer)
    }
}

func (c *Codegen) generateTopLevelNodes() {
    for _, node := range c.tree.Nodes {
        switch n := node.(type) {
        case *parser.TemplateNode:
        case *parser.FuncDeclNode:
            c.generateFuncDecl(n)
        case *parser.VarDeclNode:
//            c.generateVarDecl(n, true)
        }
    }
}

func (c *Codegen) generateFuncDecl(node *parser.FuncDeclNode) {
    c.enterScope()
    block := c.functions[node.Function.Signature.Name.Value]
    c.builder.SetInsertPoint(block, block.FirstInstruction())

    var ret bool
    for _, n := range node.Function.Body.Nodes {
        switch t := n.(type) {
            case *parser.VarDeclNode:
                c.generateVarDecl(t, false)
            case *parser.ReturnStmtNode:
                ret = true
                c.generateReturn(t)
        }
    }

    if !ret {
        c.builder.CreateRetVoid()
    }
    c.exitScope()
}

func (c *Codegen) generateTemplateDecl(node *parser.TemplateNode) {

}

func (c *Codegen) generateReturn(node *parser.ReturnStmtNode) {
    ret := c.generateExpression(node.Value)
    c.builder.CreateRet(ret)
}

func (c *Codegen) generateVarDecl(node *parser.VarDeclNode, top bool) {
    t := c.getLLVMType(node.Type)
    name := node.Name.Value
    if c.scope.Declared(name) {
        // Error name has already been declared
    }
    alloc := c.builder.CreateAlloca(t, name)
    c.scope.AddValue(name, alloc)

    var val llvm.Value
    if node.Value == nil {
        val = c.getLLVMDefaultValue(node.Type)
    } else {
        val = c.generateExpression(node.Value)
    }

    c.builder.CreateStore(val, alloc)
}

func (c *Codegen) generateExpression(node parser.Node) llvm.Value {
    switch n := node.(type) {
    case *parser.BinaryExprNode:
        return c.generateBinaryExpression(n)
    case *parser.NumLitNode:
        if n.IsFloat {
            return llvm.ConstFloat(PRIMITIVE_TYPES["float"], n.FloatValue)
        } else {
            return llvm.ConstInt(PRIMITIVE_TYPES["int"], uint64(n.IntValue), false)
        }
    case *parser.VarAccessNode:
        v := n.Name.Value
        val := c.builder.CreateLoad(c.scope.GetValue(v), "")
        return val
    }

    return llvm.Value{}
}

func (c *Codegen) generateBinaryExpression(node *parser.BinaryExprNode) llvm.Value {
    left := c.generateExpression(node.Left)
    right := c.generateExpression(node.Right)

    t := c.getLLVMType(node)
    switch node.Operator.Value {
    case "+":
        if t == PRIMITIVE_TYPES["float"] {
            return c.builder.CreateFAdd(left, right, "")
        } else if t == PRIMITIVE_TYPES["int"] {
            return c.builder.CreateAdd(left, right, "")
        }
    case "-":
        if t == PRIMITIVE_TYPES["float"] {
            return c.builder.CreateFSub(left, right, "")
        } else if t == PRIMITIVE_TYPES["int"] {
            return c.builder.CreateSub(left, right, "")
        }
    case "*":
        if t == PRIMITIVE_TYPES["float"] {
            return c.builder.CreateFMul(left, right, "")
        } else if t == PRIMITIVE_TYPES["int"] {
            return c.builder.CreateMul(left, right, "")
        }
    case "/":
        if t == PRIMITIVE_TYPES["float"] {
            return c.builder.CreateFDiv(left, right, "")
        } else if t == PRIMITIVE_TYPES["int"] {
            return c.builder.CreateSDiv(left, right, "")
        }
    }

    return llvm.Value{}
}
