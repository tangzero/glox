package glox

type Token struct{}

type Scanner struct{}

func NewScanner(source string) *Scanner {
	return &Scanner{}
}

func (s *Scanner) ScanTokens() []Token {
	return []Token{}
}
