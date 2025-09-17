package glox

import "fmt"

type Env interface {
	Define(name Token, value any)
	Assign(name Token, value any) error
	Get(name Token) (any, error)
}

type Environment struct {
	values    map[string]any
	enclosing Env
}

func NewEnvironment(enclosing Env) *Environment {
	return &Environment{
		values:    make(map[string]any),
		enclosing: enclosing,
	}
}

func (e Environment) Define(name Token, value any) {
	e.values[name.Lexeme] = value
}

func (e Environment) Assign(name Token, value any) error {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return nil
	}
	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}
	return Error(name.Line, fmt.Sprintf("undefined variable '%s", name.Lexeme))
}

func (e Environment) Get(name Token) (any, error) {
	if value, ok := e.values[name.Lexeme]; ok {
		return value, nil
	}
	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}
	return nil, Error(name.Line, fmt.Sprintf("undefined variable '%s", name.Lexeme))
}
