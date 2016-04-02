package parser

import (
    "log"
    "strings"
    "strconv"
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

func (p *parser) matchToken(ahead int, t lexer.TokenType, contents ...string) bool {
    tok := p.peek(ahead)
    if tok == nil || tok.Type != t {
        return false
    }

    for i := 0; i != len(contents); i++ {
        if contents[i] == "" || contents[i] == tok.Content {
            return true
        }
    }

    return false
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

func (p *parser) parseDecl() (node ParseNode) {
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

func (p *parser) parseVarDecl() (res *VarDeclNode) {
    t := p.parseTypeReference()

    name := NewIdentifier(p.expect(lexer.TOKEN_IDENTIFIER, ""))

    var value ParseNode
    if p.matchToken(0, lexer.TOKEN_OPERATOR, "=") {
        p.consume()

        value = p.parseExpr()
    }

    res = &VarDeclNode{
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
    return
}

func (p *parser) parseTypeReference() (node ParseNode) {
    if p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
        node = p.parseFunctionType()
    } else if  p.matchToken(0, lexer.TOKEN_SEPARATOR, "[") {
        node = p.parseArrayType()
    } else {
        node = p.parseNamedType()
    }

    return
}

func (p *parser) parseArrayType() (res *ArrayTypeNode) {
    if !p.matchToken(0, lexer.TOKEN_SEPARATOR, "[") {
        return nil
    }
    start := p.consume()
    p.expect(lexer.TOKEN_SEPARATOR, "]")

    t := p.parseTypeReference()

    res = &ArrayTypeNode{MemberType: t}
    res.SetLoc(lexer.Span{start.Location.Start, t.Loc().End})
    return
}

func (p *parser) parseFunctionType() (res *FunctionTypeNode) {
    start := p.consume()
    p.expect(lexer.TOKEN_SEPARATOR, "(")
    params := p.parseTypes()
    p.expect(lexer.TOKEN_SEPARATOR, ")")

    p.expect(lexer.TOKEN_OPERATOR, ">")

    p.expect(lexer.TOKEN_SEPARATOR, "(")
    returns := p.parseTypes()
    end := p.expect(lexer.TOKEN_SEPARATOR, ")")

    res = &FunctionTypeNode{Parameters: params, Return: returns}
    res.SetLoc(lexer.Span{start.Location.Start, end.Location.End})
    return
}

func (p *parser) parseTypes() (types []ParseNode) {
    for {
        if p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
            break
        }

        types = append(types, p.parseTypeReference())

        if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ",") {
            break
        }

        p.consume()
    }
    p.expect(lexer.TOKEN_SEPARATOR, ")")

    return
}

func (p *parser) parseNamedType() (res *NamedTypeNode) {
    name := NewIdentifier(p.expect(lexer.TOKEN_IDENTIFIER, ""))

    res = &NamedTypeNode{Name: name}
    res.SetLoc(name.Loc)
    return
}

func (p *parser) parseExpr() (res ParseNode) {
    res = p.parsePostfixExpr()
    if res == nil {
        return
    }

    return
}

func (p *parser) parsePostfixExpr() (res ParseNode) {
    res = p.parsePrimaryExpr()
    if res == nil {
        return
    }

    for {
        if p.matchToken(0, lexer.TOKEN_SEPARATOR, ".") {
            res = p.parseObjectAccess(res)
        } else if p.matchToken(0, lexer.TOKEN_SEPARATOR, "[") {
            res = p.parseArrayAccess(res)
        } else if p.matchToken(0, lexer.TOKEN_SEPARATOR, "(") {
            res = p.parseCallExpr(res)
        } else {
            break
        }
    }

    return
}

func (p *parser) parseObjectAccess(obj ParseNode) (res *ObjectAccessNode) {
    p.consume()
    member := p.expect(lexer.TOKEN_IDENTIFIER, "")

    res = &ObjectAccessNode{Object: obj, Member: NewIdentifier(member)}
    res.SetLoc(lexer.Span{obj.Loc().Start, member.Location.End})
    return
}

func (p *parser) parseArrayAccess(arr ParseNode) (res *ArrayAccessNode) {
    p.consume()
    index := p.parseExpr()
    end := p.expect(lexer.TOKEN_SEPARATOR, "]")

    res = &ArrayAccessNode{Array: arr, Index: index}
    res.SetLoc(lexer.Span{arr.Loc().Start, end.Location.End})
    return
}

func (p *parser) parseCallExpr(fn ParseNode) (res *CallExprNode) {
    p.consume()

    args := p.parseArguments()
    end := p.expect(lexer.TOKEN_SEPARATOR, ")")

    res = &CallExprNode{Function: fn, Arguments: args}
    res.SetLoc(lexer.Span{fn.Loc().Start, end.Location.End})
    return
}

func (p *parser) parseArguments() (args []ParseNode) {
    for {
        if p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
            break
        }

        args = append(args, p.parseExpr())

        if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ",") {
            break
        }
        p.consume()
    }

    return
}

func (p *parser) parsePrimaryExpr() (res ParseNode) {
    if makeExpr := p.parseMakeExpr(); makeExpr != nil {
        res = makeExpr
    } else if litExpr := p.parseLitExpr(); litExpr != nil {
        res = litExpr
    } else if unaryExpr := p.parseUnaryExpr(); unaryExpr != nil {
        res = unaryExpr
    } else if varAcc := p.parseVarAccess(); varAcc != nil {
        res = varAcc
    }

    return
}

func (p *parser) parseMakeExpr() (res *MakeExprNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, "make") {
        return nil
    }
    start     := p.consume()
    template  := NewIdentifier(p.expect(lexer.TOKEN_IDENTIFIER, ""))
    p.expect(lexer.TOKEN_SEPARATOR, "(")
    construct := p.parseArguments()
    end := p.expect(lexer.TOKEN_SEPARATOR, ")")

    res = &MakeExprNode{Template: template, Arguments: construct}
    res.SetLoc(lexer.Span{start.Location.Start, end.Location.Start})
    return
}

func (p *parser) parseLitExpr() (res ParseNode) {
    if boolLit := p.parseBoolLit(); boolLit != nil {
        res = boolLit
    } else if numLit := p.parseNumLit(); numLit != nil {
        res = numLit
    } else if strLit := p.parseStringLit(); strLit != nil {
        res = strLit
    } else if charLit := p.parseCharLit(); charLit != nil {
        res = charLit
    }

    return
}

func (p *parser) parseUnaryExpr() (res *UnaryExprNode) {
    if !p.matchToken(0, lexer.TOKEN_OPERATOR, "!", "-") {
        return
    }
    operator := p.consume()
    value := p.parsePostfixExpr()

    res = &UnaryExprNode{Value: value, Operator: operator.Content}
    res.SetLoc(lexer.Span{operator.Location.Start, value.Loc().End})
    return
}

func (p *parser) parseVarAccess() (res *VarAccessNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, "") {
        return
    }
    token := p.consume()

    res = &VarAccessNode{Name: NewIdentifier(token)}
    res.SetLoc(token.Location)
    return
}

func (p *parser) parseBoolLit() (res *BoolLitNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, "true", "false") {
        return
    }
    token := p.consume()

    value := token.Content == "true"
    res = &BoolLitNode{Value: value}
    res.SetLoc(token.Location)
    return
}

func (p *parser) parseNumLit() (res *NumLitNode) {
    if !p.matchToken(0, lexer.TOKEN_NUMBER, "") {
        return
    }
    token := p.consume()

    res = &NumLitNode{}
    res.SetLoc(token.Location)

    count := strings.Count(token.Content, ".")
    if count == 0 {
        val, err := strconv.Atoi(token.Content)
        if err == nil {
            res.IntValue = val
        }
    } else if count == 1 {
        val, err := strconv.ParseFloat(token.Content, 64)
        if err == nil {
            res.FloatValue = val
            res.IsFloat = true
        }
    }

    return
}

func (p *parser) parseStringLit() (res *StringLitNode) {
    if !p.matchToken(0, lexer.TOKEN_STRING, "") {
        return
    }
    token := p.consume()

    res = &StringLitNode{Value: token.Content}
    res.SetLoc(token.Location)
    return
}

func (p *parser) parseCharLit() (res *CharLitNode) {
    if !p.matchToken(0, lexer.TOKEN_CHARACTER, "") {
        return
    }
    token := p.consume()

    res = &CharLitNode{Value: []rune(token.Content)[0]}
    res.SetLoc(token.Location)
    return
}
