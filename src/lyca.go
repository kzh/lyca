package main

import (
    "os"
    "log"
    "strings"
    "os/exec"

    "github.com/furryfaust/lyca/src/lexer"
    "github.com/furryfaust/lyca/src/parser"
    "github.com/furryfaust/lyca/src/codegen"
)

func main() {
    path := os.Args[1];
    strip := strings.Split(path, ".")[0]

    f, err := os.Open(path);
    if err != nil {
        log.Fatal(err)
    }

    file := lexer.LycaFile(f)
    f.Close()

    toks := lexer.Lex(file)
    tree := parser.Parse(toks)
    tree.Print()

    gen := codegen.Construct(tree)
    ir  := gen.Generate()
    log.Println("\n" + ir)

    f, err = os.Create(strip + ".ll")
    if err != nil {
        log.Fatal(err)
    }

    f.WriteString(ir)

    toObj := exec.Command("llc", "-filetype=obj", strip + ".ll")
    err = toObj.Run()
    if err != nil {
        log.Fatal(err)
    }

    toBin := exec.Command("gcc", strip + ".o", "-o", strip)
    err = toBin.Run()
    if err != nil {
        log.Fatal(err)
    }

    os.Remove(strip + ".ll")
    os.Remove(strip + ".o")
}
