package codegen

import (
    "llvm.org/llvm/bindings/go/llvm"
)

type Scope struct {
    Outer *Scope
    Children []*Scope

    variables map[string]llvm.Value
}

func (s *Scope) GetValue(name string) llvm.Value {
    if res, ok := s.variables[name]; ok {
        return res
    }

    if s.Outer != nil {
        return s.Outer.GetValue(name)
    }

    return llvm.Value{}
}

func (s *Scope) AddValue(name string, val llvm.Value) {
    s.variables[name] = val
}

func (s *Scope) AddScope() *Scope {
    scope := &Scope{}
    scope.Outer = s
    s.Children = append(s.Children, scope)

    return scope
}
