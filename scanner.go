package glox

import (
	"fmt"

	"github.com/samber/lo"
)

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
			return s.Tokens, fmt.Errorf("failed to scan the next token: %w", err)
		}
	}
	s.Tokens = append(s.Tokens, Token{Type: EOF, Lexeme: "", Literal: nil})
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
	default:
		return fmt.Errorf("unexpected character: %q", c)
	}
	return nil
}

func (s *Scanner) Advance() byte {
	i := s.Current
	s.Current++
	return s.Source[i]
}

func (s *Scanner) AddToken(t TokenType) {
	s.AddTokenLiteral(t, nil)
}

func (s *Scanner) AddTokenLiteral(t TokenType, literal any) {
	text := s.Source[s.Start:s.Current]
	s.Tokens = append(s.Tokens, Token{Type: t, Lexeme: text, Literal: literal, Line: s.Line})
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
