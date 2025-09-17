package glox

import "fmt"

type Environment map[string]any

func (e Environment) Define(name Token, value any) {
	e[name.Lexeme] = value
}

func (e Environment) Assign(name Token, value any) error {
	if _, ok := e[name.Lexeme]; ok {
		e[name.Lexeme] = value
		return nil
	}
	return Error(name.Line, fmt.Sprintf("undefined variable '%s", name.Lexeme))
}

func (e Environment) Get(name Token) (any, error) {
	if value, ok := e[name.Lexeme]; ok {
		return value, nil
	}
	return nil, Error(name.Line, fmt.Sprintf("undefined variable '%s", name.Lexeme))
}
