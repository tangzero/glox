package glox

import "fmt"

type TokenType int

const (
	// Single-character tokens
	LeftParen TokenType = iota
	RightParen
	LeftBrace
	RightBrace
	Comma
	Dot
	Minus
	Plus
	Semicolon
	Slash
	Star

	// One or two character tokens
	Bang
	BangEqual
	Equal
	EqualEqual
	Greater
	GreaterEqual
	Less
	LessEqual

	// Literals
	Identifier
	String
	Number

	// Keywords
	And
	Class
	Else
	False
	Fun
	For
	If
	Nil
	Or
	Print
	Return
	Super
	This
	True
	Var
	While
	Break
	Continue

	EOF
)

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Line    int
}

func (t Token) String() string {
	return fmt.Sprintf("%v %v %v", t.Type, t.Lexeme, t.Literal)
}
