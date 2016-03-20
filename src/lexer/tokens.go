package lexer

type TokenType int

const (
    TOKEN_IDENTIFIER TokenType = iota

    TOKEN_NUMBER
    TOKEN_STRING
    TOKEN_CHARACTER

    TOKEN_OPERATOR
    TOKEN_SEPARATOR
)

type Token struct {
    Type     TokenType
    Content  string
    Location Span
}

var TOKEN_NAMES = []string{
    "IDENTIFIER",
    "NUMBER",
    "STRING",
    "CHARACTER",
    "OPERATOR",
    "SEPARATOR",
}
