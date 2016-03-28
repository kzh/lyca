package parser

import (
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

type FunctionTypeNode struct {
    baseNode
    Parameters []*TypeReferenceNode
    Return []*TypeReferenceNode
}

type NamedTypeNode struct {
    baseNode
    Name Identifier
}

type TypeReferenceNode struct {
    baseNode
    Type ParseNode
}

type VarDeclNode struct {
    baseNode
    Name Identifier
    Type *TypeReferenceNode
    Value ParseNode
}
