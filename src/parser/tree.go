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
    Return []ParseNode
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

type VarAccessNode struct {
    baseNode
}

func (p *ParseTree) Print() {
    for _, node := range p.Nodes {
        p.printNode(node, 0)
    }
}

func (p *ParseTree) printNode(node ParseNode, pad int) {
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
        padPrint("Returns: ", pad + 1)
        for _, ret := range node.Return {
            p.printNode(ret, pad + 2)
        }
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
    }
}

func padPrint(s string, pad int) {
    padding := ""
    for ; pad != 0; pad-- {
        padding += "    ";
    }

    log.Println(padding + s)
}
