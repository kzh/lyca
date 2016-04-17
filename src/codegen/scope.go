package codegen

import (
    "llvm.org/llvm/bindings/go/llvm"
)

type Variable struct {
    Type llvm.Type
    Value llvm.Value
}

type Scope struct {
    Outer *Scope
    Children []*Scope

    variables map[string]Variable
}

func (s *Scope) GetValue(name string) llvm.Value {
    if res, ok := s.variables[name]; ok {
        return res.Value
    }

    if s.Outer != nil {
        return s.Outer.GetValue(name)
    }

    return llvm.Value{}
}

func (s *Scope) GetType(name string) llvm.Type {
    if res, ok := s.variables[name]; ok {
        return res.Type
    }

    if s.Outer != nil {
        return s.Outer.GetType(name)
    }

    return llvm.VoidType()
}

func (s *Scope) Declared(name string) bool {
    _, ok := s.variables[name]
    return ok
}

func (s *Scope) AddVariable(t llvm.Type, name string, val llvm.Value) {
    s.variables[name] = Variable{t, val}
}

func (s *Scope) AddScope() *Scope {
    scope := &Scope{variables: map[string]Variable{}}
    scope.Outer = s
    s.Children = append(s.Children, scope)

    return scope
}
