package glox

import "fmt"

type Visitor[R any] interface {
	VisitBinaryExpr(expr *BinaryExpr[R]) (R, error)
	VisitGroupingExpr(expr *GroupingExpr[R]) (R, error)
	VisitLiteralExpr(expr *LiteralExpr[R]) (R, error)
	VisitUnaryExpr(expr *UnaryExpr[R]) (R, error)
}

type Expr[R any] interface {
	Accept(visitor Visitor[R]) (R, error)
}

type BinaryExpr[R any] struct {
	Left     Expr[R]
	Operator Token
	Right    Expr[R]
}

func (b *BinaryExpr[R]) Accept(visitor Visitor[R]) (R, error) {
	return visitor.VisitBinaryExpr(b)
}

type GroupingExpr[R any] struct {
	Expression Expr[R]
}

func (g *GroupingExpr[R]) Accept(visitor Visitor[R]) (R, error) {
	return visitor.VisitGroupingExpr(g)
}

type LiteralExpr[R any] struct {
	Value any
}

func (l *LiteralExpr[R]) Accept(visitor Visitor[R]) (R, error) {
	return visitor.VisitLiteralExpr(l)
}

type UnaryExpr[R any] struct {
	Operator Token
	Right    Expr[R]
}

func (u *UnaryExpr[R]) Accept(visitor Visitor[R]) (R, error) {
	return visitor.VisitUnaryExpr(u)
}

var _ Visitor[string] = &ASTPrinter{}

type ASTPrinter struct{}

func (a *ASTPrinter) Print(expr Expr[string]) (string, error) {
	return expr.Accept(a)
}

func (a *ASTPrinter) VisitBinaryExpr(expr *BinaryExpr[string]) (string, error) {
	return a.Parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (a *ASTPrinter) VisitGroupingExpr(expr *GroupingExpr[string]) (string, error) {
	return a.Parenthesize("group", expr.Expression)
}

func (a *ASTPrinter) VisitLiteralExpr(expr *LiteralExpr[string]) (string, error) {
	if expr.Value == nil {
		return "nil", nil
	}
	return fmt.Sprintf("%v", expr.Value), nil
}

func (a *ASTPrinter) VisitUnaryExpr(expr *UnaryExpr[string]) (string, error) {
	return a.Parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (a *ASTPrinter) Parenthesize(name string, exprs ...Expr[string]) (string, error) {
	result := "(" + name
	for _, expr := range exprs {
		value, err := expr.Accept(a)
		if err != nil {
			return "", err
		}
		result += " " + value
	}
	result += ")"
	return result, nil
}
