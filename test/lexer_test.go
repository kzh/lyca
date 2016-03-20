package main

import (
    "os"
    "log"
    "testing"
    "github.com/furryfaust/lyca/src/lexer"
)

func TestLexer(t *testing.T) {
    f, err := os.Open("test.lyca");
    if err != nil {
        log.Fatal(err)
    }

    lexer.Lex(lexer.LycaFile(f))
}
