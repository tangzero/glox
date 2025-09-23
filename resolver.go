package glox

import (
	"errors"
	"fmt"
)

var _ Visitor[any] = (*Resolver)(nil)

type Resolver struct {
	Interpreter *Interpreter
	Scopes      *Scopes
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		Interpreter: interpreter,
		Scopes:      new(Scopes),
	}
}

func (r *Resolver) BeginScope() {
	r.Scopes.Push(make(map[string]bool))
}

func (r *Resolver) EndScope() {
	r.Scopes.Pop()
}

func (r *Resolver) VisitBinaryExpr(expr *BinaryExpr[any]) (any, error) {
	if err := r.ResolveExpr(expr.Left); err != nil {
		return nil, err
	}
	return nil, r.ResolveExpr(expr.Right)
}

func (r *Resolver) VisitGroupingExpr(expr *GroupingExpr[any]) (any, error) {
	return nil, r.ResolveExpr(expr.Expression)
}

func (r *Resolver) VisitLiteralExpr(expr *LiteralExpr[any]) (any, error) {
	return nil, nil // do nothing
}

func (r *Resolver) VisitUnaryExpr(expr *UnaryExpr[any]) (any, error) {
	return nil, r.ResolveExpr(expr.Right)
}

func (r *Resolver) VisitVariableExpr(expr *VariableExpr[any]) (any, error) {
	if !r.Scopes.Empty() {
		if s := r.Scopes.Peek(); s.Declared(expr.Name.Lexeme) && !s.Defined(expr.Name.Lexeme) {
			return nil, errors.New("can't read local variable in its own initializer")
		}
	}
	return nil, r.ResolveLocal(expr, expr.Name)
}

func (r *Resolver) VisitAssignExpr(expr *AssignExpr[any]) (any, error) {
	if err := r.ResolveExpr(expr.Value); err != nil {
		return nil, err
	}
	return nil, r.ResolveLocal(expr, expr.Name)
}

func (r *Resolver) VisitLogicalExpr(expr *LogicalExpr[any]) (any, error) {
	if err := r.ResolveExpr(expr.Left); err != nil {
		return nil, err
	}
	return nil, r.ResolveExpr(expr.Right)
}

func (r *Resolver) VisitCallExpr(expr *CallExpr[any]) (any, error) {
	if err := r.ResolveExpr(expr.Callee); err != nil {
		return nil, err
	}
	for _, arg := range expr.Arguments {
		if err := r.ResolveExpr(arg); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (r *Resolver) VisitLambdaExpr(expr *LambdaExpr[any]) (any, error) {
	return nil, r.ResolveFunction(&FunctionStmt[any]{
		Params: expr.Params,
		Body:   expr.Body,
	})
}

func (r *Resolver) VisitExpressionStmt(stmt *ExpressionStmt[any]) error {
	return r.ResolveExpr(stmt.Expr)
}

func (r *Resolver) VisitPrintStmt(stmt *PrintStmt[any]) error {
	return r.ResolveExpr(stmt.Expr)
}

func (r *Resolver) VisitVarDeclStmt(stmt *VarDeclStmt[any]) error {
	if err := r.Declare(stmt.Name.Lexeme); err != nil {
		return err
	}
	if stmt.Initializer != nil {
		if err := r.ResolveExpr(stmt.Initializer); err != nil {
			return err
		}
	}
	r.Define(stmt.Name.Lexeme)
	return nil
}

func (r *Resolver) VisitBlockStmt(stmt *BlockStmt[any]) error {
	r.BeginScope()
	defer r.EndScope()
	return r.Resolve(stmt.Statements)
}

func (r *Resolver) VisitIfStmt(stmt *IfStmt[any]) error {
	if err := r.ResolveExpr(stmt.Condition); err != nil {
		return err
	}
	if err := r.ResolveStmt(stmt.ThenBranch); err != nil {
		return err
	}
	if stmt.ElseBranch != nil {
		if err := r.ResolveStmt(stmt.ElseBranch); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *WhileStmt[any]) error {
	if err := r.ResolveExpr(stmt.Condition); err != nil {
		return err
	}
	return r.ResolveStmt(stmt.Body)
}

func (r *Resolver) VisitBreakStmt(*BreakStmt[any]) error {
	return nil
}

func (r *Resolver) VisitContinueStmt(*ContinueStmt[any]) error {
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *ReturnStmt[any]) (err error) {
	return r.ResolveExpr(stmt.Value)
}

func (r *Resolver) VisitFunctionStmt(stmt *FunctionStmt[any]) error {
	if err := r.Declare(stmt.Name.Lexeme); err != nil {
		return err
	}
	r.Define(stmt.Name.Lexeme)
	return r.ResolveFunction(stmt)
}

func (r *Resolver) Resolve(stmts []Stmt[any]) error {
	for _, stmt := range stmts {
		if err := r.ResolveStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) ResolveStmt(stmt Stmt[any]) error {
	return stmt.Accept(r)
}

func (r *Resolver) ResolveExpr(expr Expr[any]) error {
	_, err := expr.Accept(r)
	return err
}

func (r *Resolver) ResolveFunction(stmt *FunctionStmt[any]) error {
	r.BeginScope()
	defer r.EndScope()
	for _, param := range stmt.Params {
		if err := r.Declare(param.Lexeme); err != nil {
			return err
		}
		r.Define(param.Lexeme)
	}
	return r.Resolve(stmt.Body)
}

func (r *Resolver) ResolveLocal(expr Expr[any], name Token) error {
	for i := r.Scopes.Len() - 1; i >= 0; i-- {
		if r.Scopes.At(i).Defined(name.Lexeme) {
			return r.Interpreter.Resolve(expr, r.Scopes.Len()-1-i)
		}
	}
	return nil
}

func (r *Resolver) Declare(name string) error {
	if r.Scopes.Empty() {
		return nil
	}
	scope := r.Scopes.Peek()
	if scope.Declared(name) {
		return fmt.Errorf("variable with this name already declared in this scope: %s", name)
	}
	scope.Declare(name)
	return nil
}

func (r *Resolver) Define(name string) {
	if r.Scopes.Empty() {
		return
	}
	r.Scopes.Peek().Define(name)
}
