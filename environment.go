package glox

import "fmt"

type Env interface {
	Define(name string, value any)
	Assign(name Token, value any) error
	AssignAt(distance int, name Token, value any) error
	Get(name Token) (any, error)
	GetAt(distance int, name string) (any, error)
}

type Environment struct {
	Values    map[string]any
	Enclosing Env
}

func NewEnvironment(enclosing Env) *Environment {
	return &Environment{
		Values:    make(map[string]any),
		Enclosing: enclosing,
	}
}

func (e Environment) Define(name string, value any) {
	e.Values[name] = value
}

func (e Environment) Assign(name Token, value any) error {
	if _, ok := e.Values[name.Lexeme]; ok {
		e.Values[name.Lexeme] = value
		return nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Assign(name, value)
	}
	return Error(name.Line, fmt.Sprintf("undefined variable '%s", name.Lexeme))
}

func (e Environment) AssignAt(distance int, name Token, value any) error {
	return e.Ancestor(distance).Assign(name, value)
}

func (e Environment) Get(name Token) (any, error) {
	if value, ok := e.Values[name.Lexeme]; ok {
		return value, nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}
	return nil, Error(name.Line, fmt.Sprintf("undefined variable '%s", name.Lexeme))
}

func (e Environment) GetAt(distance int, name string) (any, error) {
	return e.Ancestor(distance).Get(Token{Lexeme: name})
}

func (e Environment) Ancestor(distance int) Env {
	env := e
	for range distance {
		e, _ := env.Enclosing.(*Environment)
		env = *e
	}
	return env
}
