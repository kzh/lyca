package main

import (
    "os"
    "log"
    "testing"
    "github.com/furryfaust/lyca/src/lexer"
    "github.com/furryfaust/lyca/src/parser"
)

func TestParser(t *testing.T) {
    f, err := os.Open("src/parser_test.lyca");
    if err != nil {
        log.Fatal(err)
    }

    toks := lexer.Lex(lexer.LycaFile(f))
    tree := parser.Parse(toks)
    tree.Print()
}
