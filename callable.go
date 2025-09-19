package glox

type (
	Arity = func() int
	Call  = func(interpreter *Interpreter, arguments []any) (any, error)
)

type Callable struct {
	Arity Arity
	Call  Call
}

func NewCallable(arity Arity, call Call) Callable {
	return Callable{
		Arity: arity,
		Call:  call,
	}
}

func (Callable) String() string {
	return "<callable>"
}
