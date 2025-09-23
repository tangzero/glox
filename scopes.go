package glox

type Scope map[string]bool

func NewScope() Scope {
	return make(map[string]bool)
}

func (s Scope) Declare(name string) {
	s[name] = false
}

func (s Scope) Define(name string) {
	s[name] = true
}

func (s Scope) Declared(name string) bool {
	_, ok := s[name]
	return ok
}

func (s Scope) Defined(name string) bool {
	defined, ok := s[name]
	return ok && defined
}

type Scopes = Stack[Scope]

type Stack[T any] []T

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(item T) {
	*s = append(*s, item)
}

func (s *Stack[T]) Pop() T {
	defer func() { *s = (*s)[:len(*s)-1] }()
	return s.Peek()
}

func (s *Stack[T]) Peek() T {
	return (*s)[len(*s)-1]
}

func (s *Stack[T]) At(index int) T {
	return (*s)[index]
}

func (s *Stack[T]) Len() int {
	return len(*s)
}

func (s *Stack[T]) Empty() bool {
	return s.Len() == 0
}
