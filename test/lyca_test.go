package main

import (
    "os"
    "log"
    "testing"
    "os/exec"
    "github.com/furryfaust/lyca/src/lexer"
    "github.com/furryfaust/lyca/src/parser"
    "github.com/furryfaust/lyca/src/codegen"
)

func Test(t *testing.T) {
    f, err := os.Open("src/test.lyca");
    if err != nil {
        log.Fatal(err)
    }

    file := lexer.LycaFile(f)
    f.Close()

    toks := lexer.Lex(file)
    tree := parser.Parse(toks)
    tree.Print()

    gen := codegen.Generate(tree)
    ir  := gen.Generate()

    f, err = os.Create("src/test.ll")
    if err != nil {
        log.Fatal(err)
    }

    f.WriteString(ir)

    toObj := exec.Command("llc", "-filetype=obj", "src/test.ll")
    err = toObj.Run()
    if err != nil {
        log.Fatal(err)
    }

    toBin := exec.Command("gcc", "src/test.o", "-o", "src/test")
    err = toBin.Run()
    if err != nil {
        log.Fatal(err)
    }

    os.Remove("src/test.ll")
    os.Remove("src/test.o")
}
