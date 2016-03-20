package lexer

import (
    "os"
    "log"
    "io/ioutil"
)

type File struct {
    contents []rune
    curr     Position
}

func LycaFile(file *os.File) *File {
    contents, err := ioutil.ReadAll(file)
    if err != nil {
        log.Fatal("There was a problem reading the file " + file.Name());
    }

    return &File{[]rune(string(contents)), Position{0, 1, 1}}
}

func (f *File) peek(ahead int) rune {
    if f.curr.Raw + ahead >= len(f.contents) {
        return 0
    }

    return f.contents[f.curr.Raw + ahead]
}

func (f *File) consume() {
    log.Println("Consumed:", string(f.peek(0)))

    f.curr.Raw++
    f.curr.Offset++

    if f.peek(0) == '\n' {
        f.curr.Line++
        f.curr.Offset = 1
    }

    log.Println("Finished consuming")
}

type Position struct {
    Raw, Line, Offset int
}

type Span struct {
    Start, End Position
}
