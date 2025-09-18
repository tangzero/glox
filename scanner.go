package glox

import (
	"strconv"
	"unicode"

	"github.com/samber/lo"
)

var Keywords = map[string]TokenType{
	"and":      And,
	"class":    Class,
	"else":     Else,
	"false":    False,
	"for":      For,
	"fun":      Fun,
	"if":       If,
	"nil":      Nil,
	"or":       Or,
	"print":    Print,
	"return":   Return,
	"super":    Super,
	"this":     This,
	"true":     True,
	"var":      Var,
	"while":    While,
	"break":    Break,
	"continue": Continue,
}

type Scanner struct {
	Source  string
	Tokens  []Token
	Start   int
	Current int
	Line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		Source: source,
		Tokens: []Token{},
		Line:   1, // code always starts at line 1
	}
}

func (s *Scanner) ScanTokens() ([]Token, error) {
	for !s.AtEnd() {
		s.Start = s.Current
		if err := s.ScanToken(); err != nil {
			return s.Tokens, err
		}
	}
	s.Tokens = append(s.Tokens, Token{
		Type: EOF,
		Line: s.Line,
	})
	return s.Tokens, nil
}

func (s *Scanner) AtEnd() bool {
	return s.Current >= len(s.Source)
}

func (s *Scanner) ScanToken() error {
	c := s.Advance()
	switch c {
	case '(':
		s.AddToken(LeftParen)
	case ')':
		s.AddToken(RightParen)
	case '{':
		s.AddToken(LeftBrace)
	case '}':
		s.AddToken(RightBrace)
	case ',':
		s.AddToken(Comma)
	case '.':
		s.AddToken(Dot)
	case '-':
		s.AddToken(Minus)
	case '+':
		s.AddToken(Plus)
	case ';':
		s.AddToken(Semicolon)
	case '*':
		s.AddToken(Star)
	case '!':
		s.AddToken(lo.Ternary(s.Match('='), BangEqual, Bang))
	case '=':
		s.AddToken(lo.Ternary(s.Match('='), EqualEqual, Equal))
	case '<':
		s.AddToken(lo.Ternary(s.Match('='), LessEqual, Less))
	case '>':
		s.AddToken(lo.Ternary(s.Match('='), GreaterEqual, Greater))
	case '/':
		if s.Match('/') {
			s.AdvanceUntil('\n') // a comment goes until the end of the line.
		} else {
			s.AddToken(Slash)
		}
	case ' ', '\r', '\t':
		// ignore whitespaces
	case '\n':
		s.Line++
	case '"':
		return s.ParseString()
	default:
		if IsDigit(c) {
			return s.ParseNumber()
		}
		if IsAlpha(c) {
			return s.ParseIdentifier()
		}
		return Error(s.Line, "unexpected character")
	}
	return nil
}

func (s *Scanner) Advance() byte {
	i := s.Current
	s.Current++
	return s.Source[i]
}

func (s *Scanner) AdvanceUntil(target byte) {
	for !s.AtEnd() && s.Peek() != target {
		if s.Peek() == '\n' {
			s.Line++
		}
		s.Advance()
	}
}

func (s *Scanner) Peek() byte {
	if s.AtEnd() {
		return 0
	}
	return s.Source[s.Current]
}

func (s *Scanner) PeekNext() byte {
	if s.Current+1 >= len(s.Source) {
		return 0
	}
	return s.Source[s.Current+1]
}

func (s *Scanner) Match(expected byte) bool {
	if s.AtEnd() {
		return false
	}
	if s.Source[s.Current] != expected {
		return false
	}
	s.Current++
	return true
}

func (s *Scanner) AddToken(t TokenType) {
	s.AddTokenLiteral(t, nil)
}

func (s *Scanner) AddTokenLiteral(t TokenType, literal any) {
	text := s.Source[s.Start:s.Current]
	s.Tokens = append(s.Tokens, Token{
		Type:    t,
		Lexeme:  text,
		Literal: literal,
		Line:    s.Line,
	})
}

func (s *Scanner) ParseString() error {
	s.AdvanceUntil('"')
	if s.AtEnd() {
		return Error(s.Line, "unterminated string")
	}
	s.Advance() // the closing "
	value := s.Source[s.Start+1 : s.Current-1]
	s.AddTokenLiteral(String, value)
	return nil
}

func (s *Scanner) ParseNumber() error {
	for unicode.IsDigit(rune(s.Peek())) {
		s.Advance()
	}
	if s.Peek() == '.' && IsDigit(s.PeekNext()) {
		s.Advance() // consume the "."
		for unicode.IsDigit(rune(s.Peek())) {
			s.Advance()
		}
	}
	number, err := strconv.ParseFloat(s.Source[s.Start:s.Current], 64)
	if err != nil {
		// if we reach here, something is really wrong with our parser.
		// just return an error.
		return Error(s.Line, "invalid number")
	}
	s.AddTokenLiteral(Number, number)
	return nil
}

func (s *Scanner) ParseIdentifier() error {
	for IsAlphaNumeric(s.Peek()) {
		s.Advance()
	}
	text := s.Source[s.Start:s.Current]
	if keyword, ok := Keywords[text]; ok {
		s.AddToken(keyword) // it's a reserved keyword
		return nil
	}
	s.AddToken(Identifier)
	return nil
}

func IsDigit(c byte) bool {
	return unicode.IsDigit(rune(c))
}

func IsAlpha(c byte) bool {
	return unicode.IsLetter(rune(c)) || c == '_'
}

func IsAlphaNumeric(c byte) bool {
	return IsAlpha(c) || IsDigit(c)
}
