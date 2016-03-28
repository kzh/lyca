package parser

import (
    "log"
    "github.com/furryfaust/lyca/src/lexer"
)

type parser struct {
    Tree  *ParseTree
    tokens []*lexer.Token

    curr   int
}

func Parse(tokens []*lexer.Token) *ParseTree {
    p := &parser{
        Tree: &ParseTree{},
        tokens: tokens,
        curr: 0,
    }

    for p.peek(0) != nil {
        p.parse()
    }

    return p.Tree
}

func (p *parser) peek(ahead int) *lexer.Token {
    if p.curr + ahead >= len(p.tokens) {
        return nil
    }

    return p.tokens[p.curr + ahead]
}

func (p *parser) consume() *lexer.Token {
    tok := p.peek(0)
    p.curr++
    return tok
}

func (p *parser) matchToken(ahead int, t lexer.TokenType, content string) bool {
    tok := p.peek(ahead)
    return tok != nil && tok.Type == t && (content == tok.Content || content == "")
}

func (p *parser) matchTokens(tokens ...interface{}) bool {
    for i := 0; i != len(tokens) / 2; i++ {
        if !p.matchToken(i, tokens[i * 2].(lexer.TokenType), tokens[i * 2 +1].(string)) {
            return false
        }
    }

    return true
}

func (p *parser) expect(t lexer.TokenType, content string) *lexer.Token {
    if !p.matchToken(0, t, content) {
        log.Fatal("Unexpected token", p.peek(0).Content)
    }

    return p.consume()
}

func (p *parser) parse() {
    if node := p.parseDecl(); node != nil {
        p.Tree.AddNode(node)
    }
}

func (p *parser) parseDecl() ParseNode {
    var node ParseNode
    if tmplNode := p.parseTemplateDecl(); tmplNode != nil {
        node =  tmplNode
    } else if funcNode := p.parseFuncDecl(); funcNode != nil {
        node = funcNode
    } else if varNode := p.parseVarDecl(); varNode != nil {
        node = varNode
    }

    return node
}

func (p *parser) parseTemplateDecl() ParseNode {
    return nil
}

func (p *parser) parseFuncDecl() ParseNode {
    return nil
}

func (p *parser) parseVarDecl() *VarDeclNode {
    t := p.parseTypeReference()

    name := NewIdentifier(p.expect(lexer.TOKEN_IDENTIFIER, ""))

    var value ParseNode
    if p.matchToken(0, lexer.TOKEN_OPERATOR, "=") {
        p.consume()

    }

    res := &VarDeclNode{
        Name: name,
        Type: t,
    }

    var end lexer.Position
    if value != nil {
        res.Value = value
        end = value.Loc().End
    } else {
        end = name.Loc.End
    }

    p.expect(lexer.TOKEN_SEPARATOR, ";")
    res.SetLoc(lexer.Span{t.Loc().Start, end})
    return res
}

func (p *parser) parseTypeReference() *TypeReferenceNode {
    var node ParseNode
    if p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
        node = p.parseFunctionType()
    } else if false /* p.matchToken(0, lexer.TOKEN_SEPARATOR, "[") */ {
        //node = p.parseArrayType()
    } else {
        node = p.parseNamedType()
    }

    res := &TypeReferenceNode{Type: node}
    res.SetLoc(lexer.Span{node.Loc().Start, node.Loc().End})
    return res
}

func (p *parser) parseFunctionType() *FunctionTypeNode {
    start := p.consume()
    p.expect(lexer.TOKEN_SEPARATOR, "(")
    params := p.parseTypes()
    p.expect(lexer.TOKEN_SEPARATOR, ")")

    p.expect(lexer.TOKEN_OPERATOR, ">")

    p.expect(lexer.TOKEN_SEPARATOR, "(")
    returns := p.parseTypes()
    end := p.expect(lexer.TOKEN_SEPARATOR, ")")

    res := &FunctionTypeNode{Parameters: params, Return: returns}
    res.SetLoc(lexer.Span{start.Location.Start, end.Location.End})
    return res
}

func (p *parser) parseTypes() []*TypeReferenceNode {
    var types []*TypeReferenceNode

    for {
        if p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
            break
        }

        types = append(types, p.parseTypeReference())

        if p.matchToken(0, lexer.TOKEN_SEPARATOR, ",") {
            p.consume()
        } else if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
            //Error
        }
    }

    return types
}

func (p *parser) parseNamedType() *NamedTypeNode {
    name := NewIdentifier(p.expect(lexer.TOKEN_IDENTIFIER, ""))

    res := &NamedTypeNode{Name: name}
    res.SetLoc(name.Loc)
    return res
}
