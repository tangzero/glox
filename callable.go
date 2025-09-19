package glox

type Call = func(interpreter *Interpreter, arguments []any) (any, error)

type Callable struct {
	Name  string
	Arity int
	Call  Call
}

func NewCallable(name string, arity int, call Call) Callable {
	return Callable{
		Name:  name,
		Arity: arity,
		Call:  call,
	}
}

func (c Callable) String() string {
	return c.Name
}
