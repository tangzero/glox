package glox

type Callable interface {
	Arity() int
	Call(interpreter *Interpreter, arguments []any) (any, error)
}

type Function struct {
	closure Env
	stmt    *FunctionStmt[any]
}

func NewFunction(closure Env, stmt *FunctionStmt[any]) Callable {
	return &Function{closure, stmt}
}

func (f *Function) Arity() int {
	return len(f.stmt.Params)
}

func (f *Function) String() string {
	return "<fn " + f.stmt.Name.Lexeme + ">"
}

func (f *Function) Call(interpreter *Interpreter, arguments []any) (any, error) {
	env := NewEnvironment(f.closure)
	for i, param := range f.stmt.Params {
		env.Define(param.Lexeme, arguments[i])
	}

	if err := interpreter.ExecuteBlock(f.stmt.Body, env); err != nil {
		if returnValue, ok := err.(*ReturnValue); ok {
			return returnValue.Value, nil
		}
		return nil, err
	}
	return nil, nil
}

type Lambda struct {
	closure Env
	expr    *LambdaExpr[any]
}

func NewLambda(closure Env, expr *LambdaExpr[any]) Callable {
	return &Lambda{closure, expr}
}

func (l *Lambda) Arity() int {
	return len(l.expr.Params)
}

func (l *Lambda) String() string {
	return "<lambda>"
}

func (l *Lambda) Call(interpreter *Interpreter, arguments []any) (any, error) {
	env := NewEnvironment(l.closure)
	for i, param := range l.expr.Params {
		env.Define(param.Lexeme, arguments[i])
	}

	if err := interpreter.ExecuteBlock(l.expr.Body, env); err != nil {
		if returnValue, ok := err.(*ReturnValue); ok {
			return returnValue.Value, nil
		}
		return nil, err
	}
	return nil, nil
}

type NativeFunctionHandler func(*Interpreter, []any) (any, error)

type NativeFunction struct {
	name    string
	arity   int
	handler NativeFunctionHandler
}

func NewNativeFunction(name string, arity int, handler NativeFunctionHandler) Callable {
	return &NativeFunction{name, arity, handler}
}

func (n *NativeFunction) Arity() int {
	return n.arity
}

func (n *NativeFunction) String() string {
	return "<native fn " + n.name + ">"
}

func (n *NativeFunction) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return n.handler(interpreter, arguments)
}
