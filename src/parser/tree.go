package parser

import (
    "log"
    "strconv"
    "github.com/furryfaust/lyca/src/lexer"
)

type ParseNode interface {
    Loc() lexer.Span
    SetLoc(lexer.Span)
}

type baseNode struct {
    location lexer.Span
}

func (b *baseNode) Loc() lexer.Span {
    return b.location
}

func (b *baseNode) SetLoc(location lexer.Span) {
    b.location = location;
}

type ParseTree struct {
    baseNode
    Nodes []ParseNode
}

func (p *ParseTree) AddNode(node ParseNode) {
    p.Nodes = append(p.Nodes, node)
}

type Identifier struct {
    Loc lexer.Span
    Value string
}

func NewIdentifier(token *lexer.Token) Identifier {
    return Identifier{Loc: token.Location, Value: token.Content}
}

type ArrayTypeNode struct {
    baseNode
    MemberType ParseNode
}

type FunctionTypeNode struct {
    baseNode
    Parameters []ParseNode
    Return ParseNode
}

type NamedTypeNode struct {
    baseNode
    Name Identifier
}

type VarDeclNode struct {
    baseNode
    Name Identifier
    Type ParseNode
    Value ParseNode
}

type MakeExprNode struct {
    baseNode
    Template Identifier
    Arguments []ParseNode
}

type BoolLitNode struct {
    baseNode
    Value bool
}

type NumLitNode struct {
    baseNode
    IntValue int
    FloatValue float64
    IsFloat bool
}

type CharLitNode struct {
    baseNode
    Value rune
}

type StringLitNode struct {
    baseNode
    Value string
}

type UnaryExprNode struct {
    baseNode
    Operator string
    Value ParseNode
}

type BinaryExprNode struct {
    baseNode
    Left ParseNode
    Right ParseNode
    Operator Identifier
}

type VarAccessNode struct {
    baseNode
    Name Identifier
}

type ObjectAccessNode struct {
    baseNode
    Object ParseNode
    Member Identifier
}

type ArrayAccessNode struct {
    baseNode
    Array ParseNode
    Index ParseNode
}

type CallExprNode struct {
    baseNode
    Function ParseNode
    Arguments []ParseNode
}

type FuncExprNode struct {
    baseNode
    Function ParseNode
}

type TemplateNode struct {
    baseNode
}

type FuncNode struct {
    baseNode
    Anon bool
    Signature ParseNode
    Body ParseNode
}

type FuncSignatureNode struct {
    baseNode
    Name Identifier
    Parameters []ParseNode
    Return ParseNode
}

type FuncDeclNode struct {
    baseNode
    Function ParseNode
    Body ParseNode
}

type BlockNode struct {
    baseNode
    Nodes []ParseNode
}

type ReturnStmtNode struct {
    baseNode
    Value ParseNode
}

type CallStmtNode struct {
    baseNode
    Call ParseNode
}

func (p *ParseTree) Print() {
    for _, node := range p.Nodes {
        p.printNode(node, 0)
    }
}

func (p *ParseTree) printNode(node ParseNode, pad int) {
    if node == nil {
        return
    }

    switch node := node.(type) {
    case *VarDeclNode:
        padPrint("[Var Decl Node]", pad)
        padPrint("Name: " + node.Name.Value, pad + 1)
        padPrint("Type: ", pad + 1)
        p.printNode(node.Type, pad + 2)
        if node.Value != nil {
            padPrint("Value: ", pad + 1)
            p.printNode(node.Value, pad + 2)
        }
    case *FunctionTypeNode:
        padPrint("[Function Type Node]", pad)
        padPrint("Parameters: ", pad + 1)
        for _, param := range node.Parameters {
            p.printNode(param, pad + 2)
        }
        padPrint("Return: ", pad + 1)
        p.printNode(node.Return, pad + 2)
    case *NamedTypeNode:
        padPrint("[Named Type Node]", pad)
        padPrint("Type: " + node.Name.Value, pad + 1)
    case *ArrayTypeNode:
        padPrint("[Array Type Node]", pad)
        padPrint("Member Type: ", pad + 1)
        p.printNode(node.MemberType, pad + 2)
    case *CharLitNode:
        padPrint("[Char Lit Node]", pad)
        padPrint("Value: " + string(node.Value), pad + 1)
    case *BoolLitNode:
        padPrint("[Boolean Lit Node]", pad)
        padPrint("Value: " + strconv.FormatBool(node.Value), pad + 1)
    case *StringLitNode:
        padPrint("[String Lit Node]", pad)
        padPrint("Value: " + node.Value, pad + 1)
    case *NumLitNode:
        padPrint("[Num Lit Node]", pad)
        if node.IsFloat {
            padPrint("Value: " + strconv.FormatFloat(node.FloatValue, 'f', -1, 64), pad + 1)
        } else {
            padPrint("Value: " + strconv.Itoa(node.IntValue), pad + 1)
        }
    case *UnaryExprNode:
        padPrint("[Unary Expr Node]", pad)
        padPrint("Operator: " + node.Operator, pad + 1)
        padPrint("Value: ", pad + 1)
        p.printNode(node.Value, pad + 2)
    case *VarAccessNode:
        padPrint("[Var Access Node]", pad)
        padPrint("Name: " + node.Name.Value, pad + 1)
    case *ObjectAccessNode:
        padPrint("[Object Access Node]", pad)
        padPrint("Object: ", pad + 1)
        p.printNode(node.Object, pad + 2)
        padPrint("Member: " + node.Member.Value, pad + 1)
    case *ArrayAccessNode:
        padPrint("[Array Access Node]", pad)
        padPrint("Array: ", pad + 1)
        p.printNode(node.Array, pad + 2)
        padPrint("Index: ", pad + 1)
        p.printNode(node.Index, pad + 2)
    case *CallExprNode:
        padPrint("[Call Expr Node]", pad)
        padPrint("Function: ", pad + 1)
        p.printNode(node.Function, pad + 2)
        padPrint("Arguments: ", pad + 1)
        for _, arg := range node.Arguments {
            p.printNode(arg, pad + 2)
        }
    case *MakeExprNode:
        padPrint("[Make Expr Node]", pad)
        padPrint("Template: " + node.Template.Value, pad + 1)
        for _, arg := range node.Arguments {
            p.printNode(arg, pad + 2)
        }
    case *BinaryExprNode:
        padPrint("[Binary Expr Node]", pad)
        padPrint("Operator: " + node.Operator.Value, pad + 1)
        padPrint("Left: ", pad + 1)
        p.printNode(node.Left, pad + 2)
        padPrint("Right: ", pad + 1)
        p.printNode(node.Right, pad + 2)
    case *FuncDeclNode:
        padPrint("[Func Decl Node]", pad)
        padPrint("Function: ", pad + 1)
        p.printNode(node.Function, pad + 2)
    case *FuncNode:
        padPrint("[Func Node]", pad)
        padPrint("Signature: ", pad + 1)
        p.printNode(node.Signature, pad + 1)
        padPrint("Body: ", pad + 1)
        p.printNode(node.Body, pad + 1)
    case *FuncSignatureNode:
        padPrint("[Func Signature Node]", pad)
        padPrint("Name: " + node.Name.Value, pad + 1)
        padPrint("Parameters: ", pad + 1)
        for _, param := range node.Parameters {
            p.printNode(param, pad + 2)
        }
        padPrint("Return: ", pad + 1)
        p.printNode(node.Return, pad + 2)
    case *BlockNode:
        padPrint("[Block Node]", pad)
        padPrint("Nodes: ", pad + 1)
        for _, node := range node.Nodes {
            p.printNode(node, pad + 2)
        }
    case *CallStmtNode:
        padPrint("[Call Stmt Node]", pad)
        padPrint("Expr: ", pad + 1)
        p.printNode(node.Call, pad + 2)
    case *ReturnStmtNode:
        padPrint("[Return Stmt Node]", pad)
        padPrint("Return: ", pad + 1)
        p.printNode(node.Value, pad + 2)
    }
}

func padPrint(s string, pad int) {
    padding := ""
    for ; pad != 0; pad-- {
        padding += "    ";
    }

    log.Println(padding + s)
}
