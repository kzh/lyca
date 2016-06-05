package codegen

import (
    "log"

    "llvm.org/llvm/bindings/go/llvm"
    "github.com/furryfaust/lyca/src/parser"
)

type Template struct {
    Type llvm.Type
    Variables map[string]int
    Values []*parser.VarDeclNode
    HasConstructor bool
}

type Codegen struct {
    tree *parser.AST
    scope *Scope

    module llvm.Module
    builder llvm.Builder

    templates map[string]*Template
    functions map[string]llvm.BasicBlock

    currFunc string
}

func Construct(tree *parser.AST) *Codegen {
    return &Codegen{
        tree: tree,
        scope: &Scope{variables: map[string]llvm.Value{}},

        module: llvm.NewModule("main"),
        builder: llvm.NewBuilder(),

        templates: map[string]*Template{},
        functions: map[string]llvm.BasicBlock{},
    }
}

func (c *Codegen) Generate() string {
    c.injectStdLib()
    c.declareTopLevelNodes()
    c.generateTopLevelNodes()

    if ok := llvm.VerifyModule(c.module, llvm.ReturnStatusAction); ok != nil {
        log.Println(ok.Error())
    }

    return c.module.String()
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
        case *parser.TemplateNode:
            c.presetTemplate(n)
        }
    }

    for _, node := range c.tree.Nodes {
        switch n := node.(type) {
        case *parser.FuncDeclNode:
            c.declareFunc(n.Function, llvm.VoidType())
        case *parser.TemplateNode:
            c.declareTemplate(n)
        }
    }
}

func (c *Codegen) presetTemplate(n *parser.TemplateNode) {
    c.templates[n.Name.Value] = &Template{
        Type: llvm.GlobalContext().StructCreateNamed(n.Name.Value),
        Variables: map[string]int{},
    }
    c.templates[n.Name.Value].Values = n.Variables
}

func (c *Codegen) declareFunc(n *parser.FuncNode, obj llvm.Type) {
    sig := n.Signature
    name := c.mangle(sig.Name.Value)
    f := c.getLLVMFuncType(sig.Return, sig.Parameters, obj)
    llvmf := llvm.AddFunction(c.module, name, f)

    offset := 0
    if obj != llvm.VoidType() {
        llvmf.Param(0).SetName("this")
        offset = 1
    }

    for i, name := range sig.Parameters {
        llvmf.Param(i + offset).SetName(name.Name.Value)
    }

    block := llvm.AddBasicBlock(c.module.NamedFunction(name),"entry")

    c.functions[name] = block
}

func (c *Codegen) declareTemplate(n *parser.TemplateNode) {
    name := n.Name.Value
    var vars []llvm.Type
    for i, v := range n.Variables {
        vars = append(vars, c.getLLVMType(v.Type))
        c.templates[name].Variables[v.Name.Value] = i
    }

    c.templates[name].Type.StructSetBody(vars, false)
    pointer := llvm.PointerType(c.templates[name].Type, 0)

    if n.Constructor != nil {
       f := &parser.FuncNode{
            Signature: &parser.FuncSignatureNode{
                Name: parser.Identifier{Value: "-" + n.Name.Value},
                Parameters: n.Constructor.Parameters,
            },
        }
        c.declareFunc(f, pointer)
        c.templates[name].HasConstructor = true
    }

    for _, meth := range n.Methods {
        name := "-" + n.Name.Value + "-" + meth.Function.Signature.Name.Value
        meth.Function.Signature.Name = parser.Identifier{Value: name}
        c.declareFunc(meth.Function, pointer)
    }
}

func (c *Codegen) getFunction(node parser.Node) (llvm.Value, []llvm.Value) {
    switch t := node.(type) {
    case *parser.VarAccessNode:
        name := c.mangle(t.Name.Value)
        return c.module.NamedFunction(name), []llvm.Value{}
    case *parser.ObjectAccessNode:
        tmpl := c.getStructFromPointer(c.getLLVMType(t.Object))
        obj := c.generateAccess(t.Object, true)

        return c.module.NamedFunction("-" + tmpl + "-" + t.Member.Value), []llvm.Value{obj}
    }

    return null, []llvm.Value{}
}

func (c *Codegen) getCurrParam(name string) llvm.Value {
    currFunc := c.module.NamedFunction(c.currFunc)
    for _, param := range currFunc.Params() {
        if param.Name() == name {
            return param
        }
    }

    return null
}

func (c *Codegen) generateTopLevelNodes() {
    for _, node := range c.tree.Nodes {
        switch n := node.(type) {
        case *parser.TemplateNode:
            c.generateTemplateDecl(n)
        case *parser.FuncDeclNode:
            c.generateFunc(n.Function)
        case *parser.VarDeclNode:
            c.generateVarDecl(n, true)
        }
    }
}

func (c *Codegen) generateFunc(node *parser.FuncNode) {
    c.enterScope()
    c.currFunc = c.mangle(node.Signature.Name.Value)
    block := c.functions[c.currFunc]
    c.builder.SetInsertPoint(block, block.LastInstruction())

    var ret bool
    for _, n := range node.Body.Nodes {
        switch t := n.(type) {
            case *parser.VarDeclNode:
                c.generateVarDecl(t, false)
            case *parser.AssignStmtNode:
                c.generateAssign(t)
            case *parser.CallStmtNode:
                c.generateCall(t.Call, null)
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
    name := node.Name.Value

    if node.Constructor != nil {
       f := &parser.FuncNode{
            Signature: &parser.FuncSignatureNode{
                Name: parser.Identifier{Value: "-" + name},
            },
            Body: node.Constructor.Body,
        }
        c.generateFunc(f)
    }

    for _, meth := range node.Methods {
        c.generateFunc(meth.Function)
    }
}

func (c *Codegen) generateAssign(node *parser.AssignStmtNode) {
    access := c.generateAccess(node.Target, false)
    expr := c.convert(c.generateExpression(node.Value), access.Type().ElementType())

    c.builder.CreateStore(expr, access)
}

func (c *Codegen) generateCall(node *parser.CallExprNode, obj llvm.Value) llvm.Value {
    fn, args := c.getFunction(node.Function)

    if !obj.IsNil() {
        args = append([]llvm.Value{obj}, args...)
    }

    for i, arg := range node.Arguments {
        expr := c.generateExpression(arg)
        if fn.Type().ElementType().ParamTypesCount() > i {
            expr = c.convert(expr, fn.Type().ElementType().ParamTypes()[i])
        }

        //Unbox arguments for C functions
        if fn.BasicBlocksCount() == 0 {
            expr = c.unbox(expr)
        }

        args = append(args, expr)
    }

    return c.builder.CreateCall(fn, args, "")
}

func (c *Codegen) generateMake(node *parser.MakeExprNode) llvm.Value {
    t := c.templates[node.Template.Value]
    alloc := c.builder.CreateMalloc(t.Type, "")

    if t.HasConstructor {
        call := &parser.CallExprNode{
            Function: &parser.VarAccessNode{Name: parser.Identifier{Value: "-" + node.Template.Value}},
            Arguments: node.Arguments,
        }
        c.generateCall(call, alloc)
    }

    return alloc
}

func (c *Codegen) generateReturn(node *parser.ReturnStmtNode) {
    t := c.module.NamedFunction(c.currFunc).Type().ElementType().ReturnType()
    ret := c.convert(c.generateExpression(node.Value), t)

    c.builder.CreateRet(ret)
}

func (c *Codegen) generateVarDecl(node *parser.VarDeclNode, global bool) {
    t := c.getLLVMType(node.Type)
    name := node.Name.Value
    if c.scope.Declared(name) {
        // Error name has already been declared
    }

    var alloc, val llvm.Value
    if node.Value == nil {
        val = llvm.Undef(t)
    } else {
        val = c.convert(c.generateExpression(node.Value), t)
    }

    if !global {
        alloc = c.builder.CreateAlloca(t, name)
        c.builder.CreateStore(val, alloc)
    } else {
        alloc = llvm.AddGlobal(c.module, t, name)
        alloc.SetInitializer(val)
    }

    c.scope.AddVariable(name, alloc)
}

func (c *Codegen) generateAccess(node parser.Node, val bool) (v llvm.Value) {
    switch t := node.(type) {
    case *parser.VarAccessNode:
        name := t.Name.Value
        if param := c.getCurrParam(name); !param.IsNil() {
            return param
        } else {
            v = c.scope.GetValue(name);
        }
    case *parser.ObjectAccessNode:
        obj := c.generateAccess(t.Object, true)
        index := c.templates[obj.Type().ElementType().StructName()].Variables[t.Member.Value]
        v = c.builder.CreateStructGEP(obj, index, "")
    case *parser.StringLitNode:
        return c.generateStringLiteral(t)
    case *parser.CallExprNode:
        return c.generateCall(t, null)
    case *parser.MakeExprNode:
        return c.generateMake(t)
    }

    if val {
        v = c.builder.CreateLoad(v, "")
    }
    return
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
    case *parser.BoolLitNode:
        i := 0
        if n.Value {
            i = 1
        }

        return llvm.ConstInt(PRIMITIVE_TYPES["boolean"], uint64(i), false)
    case *parser.CharLitNode:
        return llvm.ConstInt(PRIMITIVE_TYPES["char"], uint64(n.Value), false)
    case *parser.VarAccessNode, *parser.ObjectAccessNode, *parser.ArrayAccessNode, *parser.CallExprNode, *parser.StringLitNode, *parser.MakeExprNode:
        return c.generateAccess(n, true)
    }

    return null
}

func (c *Codegen) generateBinaryExpression(node *parser.BinaryExprNode) llvm.Value {
    left := c.generateExpression(node.Left)
    right := c.generateExpression(node.Right)
    if left.Type() == PRIMITIVE_TYPES["float"] && right.Type() == PRIMITIVE_TYPES["int"] {
        right = c.convert(right, PRIMITIVE_TYPES["float"])
    } else if left.Type() == PRIMITIVE_TYPES["int"] && right.Type() == PRIMITIVE_TYPES["float"] {
        left = c.convert(left, PRIMITIVE_TYPES["float"])
    }

    t := left.Type()
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

    return null
}
