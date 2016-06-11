package lexer

import (
    "log"
    "strings"
)

type lexer struct {
    *File
    start  Position

    Tokens []*Token
}

func Lex(f *File) []*Token {
    lexer := &lexer{
        f,
        Position{0,1,1},
        []*Token{},
    }

    for {
        if lexer.peek(0) == 0 {
            break;
        }
        lexer.lex()
    }

    return lexer.Tokens
}

func (l *lexer) lex() {
    l.resetToken()

    if l.peek(0) == '/' && (l.peek(1) == '/' || l.peek(1) == '*') {
        l.ignoreComment()
    } else if l.peek(0) == ' ' {
        l.ignoreWhitespace()
    } else if l.peek(0) == '\n' {
        l.consume()
    } else if l.peek(0) == '_' || IsLetter(l.peek(0)) {
        l.lexIdentifier()
    } else if l.peek(0) == '"' {
        l.lexString()
    } else if IsDecimal(l.peek(0)) {
        l.lexNumber()
    } else if l.peek(0) == '\'' {
        l.lexCharacter()
    } else if IsOperator(l.peek(0)) {
        l.lexOperator()
    } else if IsSeparator(l.peek(0)) {
        l.lexSeparator()
    }
}

func (l *lexer) resetToken() {
    l.start = l.curr
}

func (l *lexer) pushToken(t TokenType) {
    l.Tokens = append(l.Tokens, &Token{
        t,
        string(l.contents[l.start.Raw:l.curr.Raw]),
        Span{l.start, l.curr},
    })
}

func (l *lexer) ignoreComment() {
    l.consume()
    l.expect('/', '*')
    if l.peek(0) == '/' {
        l.consume()
        for l.peek(0) != '\n' {
            l.consume()
        }
        l.consume()
    } else if l.peek(0) == '*' {
        l.consume()
        for l.peek(0) != '*' && l.peek(0) != '/' {
            l.consume()
        }
        l.consume()
        l.consume()
    }
}

func (l *lexer) ignoreWhitespace() {
    l.consume()
    for l.peek(0) == ' ' {
        l.consume()
    }
}

func (l *lexer) lexIdentifier() {
    l.consume()
    for IsLetter(l.peek(0)) || IsDecimal(l.peek(0)) || l.peek(0) == '_' {
        l.consume()
    }

    l.pushToken(TOKEN_IDENTIFIER)
}

func (l *lexer) lexString() {
    l.consume()
    l.resetToken()
    for {
        if l.peek(0) == '\\' {
            l.consume()
            l.lexEscape('"')
        } else if l.peek(0) == '"' {
            l.pushToken(TOKEN_STRING)
            l.consume()
            break
        } else if l.peek(0) == 0 || l.peek(0) == '\n' {
            log.Fatal("Unterminated string literal")
        } else {
            l.consume()
        }
    }
}

func (l *lexer) lexEscape(r rune) {
    switch l.peek(0) {
    case '\\', 'n', r:
        l.consume()
    }
}

func (l *lexer) lexNumber() {
    l.consume()
    for IsDecimal(l.peek(0)) || l.peek(0) == '.' {
        l.consume()
    }

    l.pushToken(TOKEN_NUMBER)
}

func (l *lexer) lexCharacter() {
    l.consume()
    l.resetToken()
    for {
        if l.peek(0) == '\\' {
            l.consume()
            l.lexEscape('\'')
        } else if l.peek(0) == '\'' {
            l.pushToken(TOKEN_CHARACTER)
            l.consume()
            break
        } else if l.peek(0) == 0 || l.peek(0) == '\n' {
            log.Fatal("Unterminated character literal")
        } else {
            l.consume()
        }
    }
}

func (l *lexer) lexOperator() {
    if strings.ContainsRune("=!><", l.peek(0)) {
        l.consume()
        if l.peek(0) == '=' {
            l.consume()
        }
    } else {
        l.consume()
        if IsOperator(l.peek(0)) && l.peek(0) != '=' {
            l.consume()
        }
    }

    l.pushToken(TOKEN_OPERATOR)
}

func (l *lexer) lexSeparator() {
    l.consume()
    l.pushToken(TOKEN_SEPARATOR)
}

func IsLetter(r rune) bool {
    return (r >= 'a' && r <= 'z') || (r >= 'A' && r <='Z')
}

func IsDecimal(r rune) bool {
    return r >= '0' && r <= '9'
}

func IsOperator(r rune) bool {
    return strings.ContainsRune("+-*/=><!|&%", r)
}

func IsSeparator(r rune) bool {
    return strings.ContainsRune(" :;,.(){}[]", r)
}

func (l *lexer) expect(runes ...rune) {
    for _, r := range runes {
        if l.peek(0) == r {
            return
        }
    }

    log.Fatal("Unexpected token:", string(l.peek(0)))
}

func (l *lexer) printTokens() {
    log.Print("Tokens: [")
    for _, tok := range l.Tokens {
        log.Println(TOKEN_NAMES[tok.Type], tok.Content)
    }
    log.Print("]")
}
