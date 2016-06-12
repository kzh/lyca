package parser

import (
    "log"
    "strings"
    "strconv"
    "github.com/furryfaust/lyca/src/lexer"
)

var OPERATOR_PRECEDENCE map[string]int = map[string]int{
    "==": 1, "!=": 1,
    ">":  2, "<":  2, ">=": 2, "<=": 2,
    "+":  3, "-":  3,
    "*":  4, "/":  4, "%": 4,
}

type parser struct {
    Tree  *AST
    tokens []*lexer.Token

    curr   int
}

func Parse(tokens []*lexer.Token) *AST {
    p := &parser{
        Tree: &AST{},
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
        log.Fatal(p.peek(0).Location.Start.Line, ":", p.peek(0).Location.Start.Offset, " Unexpected token ", p.peek(0).Content, " Expected ", content)
    }

    return p.consume()
}

func (p *parser) parse() {
    if node := p.parseDecl(); node != nil {
        p.Tree.AddNode(node)
    }
}

func (p *parser) parseDecl() (node Node) {
    if tmplNode := p.parseTemplateDecl(); tmplNode != nil {
        node =  tmplNode
    } else if funcNode := p.parseFuncDecl(); funcNode != nil {
        node = funcNode
    } else if varNode := p.parseVarDecl(); varNode != nil {
        node = varNode
        p.expect(lexer.TOKEN_SEPARATOR, ";")
    }

    return node
}

func (p *parser) parseBlock() (res *BlockNode) {
    if !p.matchToken(0, lexer.TOKEN_SEPARATOR, "{") {
        return
    }

    start := p.expect(lexer.TOKEN_SEPARATOR, "{")
    var nodes []Node
    for {
        node := p.parseNode()
        if node == nil {
            break
        }
        nodes = append(nodes, node)
    }
    end := p.expect(lexer.TOKEN_SEPARATOR, "}")

    res = &BlockNode{Nodes: nodes}
    res.SetLoc(lexer.Span{start.Location.Start, end.Location.End})
    return
}

func (p *parser) parseNode() (res Node) {
    stmt, term := p.parseStmt();
    if stmt != nil {
        res = stmt
    } else if varDecl := p.parseVarDecl(); varDecl != nil {
        res = varDecl
    }

    if res != nil && term {
        p.expect(lexer.TOKEN_SEPARATOR, ";")
    }
    return
}

func (p *parser) parseTemplateDecl() (res *TemplateNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_TMPL) {
        return
    }
    start := p.consume()

    res = &TemplateNode{}
    res.Name = NewIdentifier(p.expect(lexer.TOKEN_IDENTIFIER, ""))
    p.expect(lexer.TOKEN_SEPARATOR, "{")
    for {
        if p.matchToken(0, lexer.TOKEN_SEPARATOR, "}") {
            break
        }

        if construct := p.parseConstructor(); construct != nil {
            res.Constructor = construct
        } else if method := p.parseFuncDecl(); method != nil {
            res.Methods = append(res.Methods, method)
        } else if variable := p.parseVarDecl(); variable != nil {
            res.Variables = append(res.Variables, variable)
            p.expect(lexer.TOKEN_SEPARATOR, ";")
        }
    }
    end := p.expect(lexer.TOKEN_SEPARATOR, "}")
    res.SetLoc(lexer.Span{start.Location.Start, end.Location.End})
    return
}

func (p *parser) parseConstructor() (res *ConstructorNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_CONSTRUCTOR) {
        return
    }
    start := p.consume()
    p.expect(lexer.TOKEN_OPERATOR, "<")
    p.expect(lexer.TOKEN_SEPARATOR, "(")
    var params []*VarDeclNode
    for {
        if p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
            break
        }

        decl := p.parseVarDecl()
        params = append(params, decl)

        if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ",") {
            break
        }
        p.consume()
    }
    p.expect(lexer.TOKEN_SEPARATOR, ")")
    body := p.parseBlock()

    res = &ConstructorNode{Parameters: params, Body: body}
    res.SetLoc(lexer.Span{start.Location.Start, body.Loc().End})
    return
}

func (p *parser) parseFuncDecl() (res *FuncDeclNode) {
    function := p.parseFunc(false)
    if function == nil {
        return
    }

    res = &FuncDeclNode{Function: function}
    res.SetLoc(function.Loc())
    return
}

func (p *parser) parseFunc(anon bool) (res *FuncNode) {
    sig := p.parseFuncSignature(anon)
    if sig == nil {
        return
    }
    body := p.parseBlock()

    res = &FuncNode{Signature: sig, Body: body}
    res.SetLoc(lexer.Span{sig.Loc().Start, body.Loc().End})
    return
}

func (p *parser) parseFuncSignature(anon bool) (res *FuncSignatureNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
        return
    }
    rollback := p.curr
    start := p.consume()
    var end *lexer.Token

    p.expect(lexer.TOKEN_SEPARATOR, "(")
    var params []*VarDeclNode
    for {
        if p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
            break
        }

        decl := p.parseVarDecl()
        if decl == nil {
            goto rollback
        }
        params = append(params, decl)

        if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ",") {
            break
        }
        p.consume()
    }
    p.expect(lexer.TOKEN_SEPARATOR, ")")
    p.expect(lexer.TOKEN_OPERATOR, ">")

    res = &FuncSignatureNode{Parameters: params}
    if !anon {
        res.Name = NewIdentifier(p.expect(lexer.TOKEN_IDENTIFIER, ""))
        p.expect(lexer.TOKEN_OPERATOR, ">")
    }

    p.expect(lexer.TOKEN_SEPARATOR, "(")
    if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
        res.Return = p.parseTypeReference()
    }
    end = p.expect(lexer.TOKEN_SEPARATOR, ")")

    res.SetLoc(lexer.Span{start.Location.Start, end.Location.End})
    return

rollback:
    p.curr = rollback
    return
}

func (p *parser) parseStmt() (res Node, term bool) {
    term = true
    if ifStmt := p.parseIfStmt(); ifStmt != nil {
        res  = ifStmt
        term = false
    } else if returnStmt := p.parseReturnStmt(); returnStmt != nil {
        res = returnStmt
    } else if callStmt := p.parseCallStmt(); callStmt != nil {
        res = callStmt
    } else if assignStmt := p.parseAssignStmt(); assignStmt != nil {
        res = assignStmt
    }
    /* else if binopAssign := p.parseBinopAssignStmt(); binopAssign != nil {
        res = assignStmt
    } */

    return
}

func (p *parser) parseIfStmt() (res *IfStmtNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IF) {
        return
    }

    token := p.consume()
    p.expect(lexer.TOKEN_SEPARATOR, "(")
    cond := p.parseExpr()
    p.expect(lexer.TOKEN_SEPARATOR, ")")
    body := p.parseBlock()

    res = &IfStmtNode{
        Condition: cond,
        Body: body,
    }

    loc := lexer.Span{token.Location.Start, body.Loc().End}
    if p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_ELSE) {
        p.consume()
        if p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_IF) {
            res.Else = p.parseIfStmt()
        } else if p.matchToken(0, lexer.TOKEN_SEPARATOR, "{") {
            res.Else = p.parseBlock()
        }
        loc.End = res.Else.Loc().End
    }

    res.SetLoc(loc)
    return
}

func (p *parser) parseCallStmt() (res *CallStmtNode) {
    rollback := p.curr

    callExpr, ok := p.parseExpr().(*CallExprNode)
    if !ok {
        p.curr = rollback
        return
    }

    res = &CallStmtNode{Call: callExpr}
    res.SetLoc(callExpr.Loc())
    return
}

func (p *parser) parseReturnStmt() (res *ReturnStmtNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_RETURN) {
        return
    }
    token := p.consume()

    var (
        expr Node
        end  lexer.Position = token.Location.End
    )

    if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ";") {
        expr = p.parseExpr()
        end  = expr.Loc().End
    }

    res = &ReturnStmtNode{Value: expr}
    res.SetLoc(lexer.Span{token.Location.Start, end})
    return
}

func (p *parser) parseAssignStmt() (res *AssignStmtNode) {
    rollback := p.curr

    target := p.parseExpr()
    if target == nil || !p.matchToken(0, lexer.TOKEN_OPERATOR, "=") {
        p.curr = rollback
        return
    }

    p.consume()
    value := p.parseExpr()

    res = &AssignStmtNode{Target: target, Value: value}
    res.SetLoc(lexer.Span{target.Loc().Start, value.Loc().End})
    return
}

func (p *parser) parseVarDecl() (res *VarDeclNode) {
    t := p.parseTypeReference()
    if t == nil || !p.matchToken(0, lexer.TOKEN_IDENTIFIER, "") {
        return
    }
    name := NewIdentifier(p.consume())

    var value Node
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

    res.SetLoc(lexer.Span{t.Loc().Start, end})
    return
}

func (p *parser) parseTypeReference() (node Node) {
    if p.matchToken(0, lexer.TOKEN_IDENTIFIER, KEYWORD_FUNC) {
        node = p.parseFuncType()
    } else if p.matchToken(0, lexer.TOKEN_SEPARATOR, "[") {
        node = p.parseArrayType()
    } else {
        node = p.parseNamedType()
    }

    return
}

func (p *parser) parseArrayType() (res *ArrayTypeNode) {
    if !p.matchToken(0, lexer.TOKEN_SEPARATOR, "[") {
        return
    }
    start := p.consume()
    p.expect(lexer.TOKEN_SEPARATOR, "]")

    t := p.parseTypeReference()

    res = &ArrayTypeNode{MemberType: t}
    res.SetLoc(lexer.Span{start.Location.Start, t.Loc().End})
    return
}

func (p *parser) parseFuncType() (res *FuncTypeNode) {
    start := p.consume()
    p.expect(lexer.TOKEN_SEPARATOR, "(")
    params := p.parseTypes()
    p.expect(lexer.TOKEN_SEPARATOR, ")")

    p.expect(lexer.TOKEN_OPERATOR, ">")

    p.expect(lexer.TOKEN_SEPARATOR, "(")
    var ret Node
    if !p.matchToken(0, lexer.TOKEN_SEPARATOR, ")") {
        ret = p.parseTypeReference()
    }
    end := p.expect(lexer.TOKEN_SEPARATOR, ")")

    res = &FuncTypeNode{Parameters: params, Return: ret}
    res.SetLoc(lexer.Span{start.Location.Start, end.Location.End})
    return
}

func (p *parser) parseTypes() (types []Node) {
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

    return
}

func (p *parser) parseNamedType() (res *NamedTypeNode) {
    if !p.matchToken(0, lexer.TOKEN_IDENTIFIER, "") {
        return
    }
    name := NewIdentifier(p.consume())

    res = &NamedTypeNode{Name: name}
    res.SetLoc(name.Loc)
    return
}

func (p *parser) parseExpr() (res Node) {
    if res = p.parsePostfixExpr(); res == nil {
        return
    }

    if bin := p.parseBinaryExpr(res, 0); bin != nil {
        res = bin
    }

    return
}

func (p *parser) parseBinaryExpr(expr Node, min int) Node {
    if !p.matchToken(0, lexer.TOKEN_OPERATOR, "") {
        return nil
    }
    rollback := p.curr

    loc := expr.Loc()
    for {
        if !p.matchToken(0, lexer.TOKEN_OPERATOR, "") {
            break
        }

        precedence, ok := OPERATOR_PRECEDENCE[p.peek(0).Content]
        if !ok || precedence < min {
            break
        }

        operator := p.consume()

        right := p.parsePostfixExpr()
        if right == nil {
            goto rollback
        }

        for p.matchToken(0, lexer.TOKEN_OPERATOR, "") && OPERATOR_PRECEDENCE[p.peek(0).Content] > precedence {
            if right = p.parseBinaryExpr(right, OPERATOR_PRECEDENCE[p.peek(0).Content]); right == nil {
                goto rollback
            }
        }

        expr = &BinaryExprNode{
            Operator: NewIdentifier(operator),
            Left: expr,
            Right: right,
        }

        loc.End = right.Loc().End
    }

    expr.SetLoc(loc)
    return expr

rollback:
    p.curr = rollback
    return nil
}

func (p *parser) parsePostfixExpr() (res Node) {
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

func (p *parser) parseObjectAccess(obj Node) (res *ObjectAccessNode) {
    p.consume()
    member := p.expect(lexer.TOKEN_IDENTIFIER, "")

    res = &ObjectAccessNode{Object: obj, Member: NewIdentifier(member)}
    res.SetLoc(lexer.Span{obj.Loc().Start, member.Location.End})
    return
}

func (p *parser) parseArrayAccess(arr Node) (res *ArrayAccessNode) {
    p.consume()
    index := p.parseExpr()
    end := p.expect(lexer.TOKEN_SEPARATOR, "]")

    res = &ArrayAccessNode{Array: arr, Index: index}
    res.SetLoc(lexer.Span{arr.Loc().Start, end.Location.End})
    return
}

func (p *parser) parseCallExpr(fn Node) (res *CallExprNode) {
    p.consume()

    args := p.parseArguments()
    end := p.expect(lexer.TOKEN_SEPARATOR, ")")

    res = &CallExprNode{Function: fn, Arguments: args}
    res.SetLoc(lexer.Span{fn.Loc().Start, end.Location.End})
    return
}

func (p *parser) parseArguments() (args []Node) {
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

func (p *parser) parsePrimaryExpr() (res Node) {
    if p.matchToken(0, lexer.TOKEN_SEPARATOR, "(") {
        p.consume()
        res = p.parseExpr();
        p.expect(lexer.TOKEN_SEPARATOR, ")")
    } else if makeExpr := p.parseMakeExpr(); makeExpr != nil {
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
    p.expect(lexer.TOKEN_OPERATOR, "<")
    p.expect(lexer.TOKEN_SEPARATOR, "(")
    construct := p.parseArguments()
    end := p.expect(lexer.TOKEN_SEPARATOR, ")")

    res = &MakeExprNode{Template: template, Arguments: construct}
    res.SetLoc(lexer.Span{start.Location.Start, end.Location.Start})
    return
}

func (p *parser) parseLitExpr() (res Node) {
    if boolLit := p.parseBoolLit(); boolLit != nil {
        res = boolLit
    } else if numLit := p.parseNumLit(); numLit != nil {
        res = numLit
    } else if strLit := p.parseStringLit(); strLit != nil {
        res = strLit
    } else if charLit := p.parseCharLit(); charLit != nil {
        res = charLit
    } else if funcLit := p.parseFuncLit(); funcLit != nil {
        res = funcLit
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

func (p *parser) parseFuncLit() (res *FuncLitNode) {
    function := p.parseFunc(true)
    if function == nil {
        return
    }

    res = &FuncLitNode{Function: function}
    res.SetLoc(function.Loc())
    log.Println("Returning func lit")
    log.Println(p.curr)
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

var ESCAPE map[rune]rune = map[rune]rune{
    '\\': '\\',
    'n': '\n',
}

func (p *parser) unescape(str string) string {
    res := []rune{}
    for i := 0; i != len(str); i++ {
        if str[i] == '\\' {
            i++
            r, _ := ESCAPE[rune(str[i])]

            res = append(res, r)
            continue;
        }

        res = append(res, rune(str[i]))
    }

    return string(res)
}

func (p *parser) parseStringLit() (res *StringLitNode) {
    if !p.matchToken(0, lexer.TOKEN_STRING, "") {
        return
    }
    token := p.consume()

    res = &StringLitNode{Value: p.unescape(token.Content)}
    res.SetLoc(token.Location)
    return
}

func (p *parser) parseCharLit() (res *CharLitNode) {
    if !p.matchToken(0, lexer.TOKEN_CHARACTER, "") {
        return
    }
    token := p.consume()

    res = &CharLitNode{Value: []rune(p.unescape(token.Content))[0]}
    res.SetLoc(token.Location)
    return
}
