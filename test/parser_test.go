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

    log.Println(len(tree.Nodes))

    // Yes this is temporary for debugging lol too sleepy to set up something better for digging through parse tree

    // Function name
    log.Println(tree.Nodes[0].(*parser.VarDeclNode).Name.Value)
    // First parameter of function
    log.Println(tree.Nodes[0].(*parser.VarDeclNode).Type.Type.(*parser.FunctionTypeNode).Parameters[0].Type.(*parser.NamedTypeNode).Name.Value)
    // Second parameter of function
    log.Println(tree.Nodes[0].(*parser.VarDeclNode).Type.Type.(*parser.FunctionTypeNode).Parameters[1].Type.(*parser.NamedTypeNode).Name.Value)

    // First return type of function
    log.Println(tree.Nodes[0].(*parser.VarDeclNode).Type.Type.(*parser.FunctionTypeNode).Return[0].Type.(*parser.NamedTypeNode).Name.Value)
    // Second return type of function
    log.Println(tree.Nodes[0].(*parser.VarDeclNode).Type.Type.(*parser.FunctionTypeNode).Return[1].Type.(*parser.NamedTypeNode).Name.Value)
}
