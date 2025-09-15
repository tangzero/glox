package glox

import (
	"fmt"
)

var _ Visitor[any] = (*Interpreter)(nil)

type Interpreter struct{}

func (i *Interpreter) VisitBinaryExpr(expr *BinaryExpr[any]) (any, error) {
	left, err := i.Evaluate(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := i.Evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Greater:
		if err := checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) > right.(float64), nil
	case GreaterEqual:
		if err := checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) >= right.(float64), nil
	case Less:
		if err := checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) < right.(float64), nil
	case LessEqual:
		if err := checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) <= right.(float64), nil
	case EqualEqual:
		return isEqual(left, right), nil
	case BangEqual:
		return !isEqual(left, right), nil
	case Minus:
		if err := checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) - right.(float64), nil
	case Slash:
		if err := checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) / right.(float64), nil
	case Star:
		if err := checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}
		return left.(float64) * right.(float64), nil
	case Plus:
		if isFloat(left) && isFloat(right) {
			return left.(float64) + right.(float64), nil
		}
		if isString(left) && isString(right) {
			return left.(string) + right.(string), nil
		}
		if isString(left) || isString(right) {
			return fmt.Sprintf("%v%v", left, right), nil
		}
		return nil, Error(expr.Operator.Line, fmt.Sprintf("'+' operation not supported for %T and %T.", left, right))
	}
	panic("unreachable")
}

func (i *Interpreter) VisitGroupingExpr(expr *GroupingExpr[any]) (any, error) {
	return i.Evaluate(expr.Expression)
}

func (i *Interpreter) VisitLiteralExpr(expr *LiteralExpr[any]) (any, error) {
	return expr.Value, nil
}

func (i *Interpreter) VisitUnaryExpr(expr *UnaryExpr[any]) (any, error) {
	right, err := i.Evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Minus:
		if err := checkNumberOperand(expr.Operator, right); err != nil {
			return nil, err
		}
		return -(right.(float64)), nil
	case Bang:
		return !isTruthy(right), nil
	}

	panic("unreachable")
}

func (i *Interpreter) Interpret(expr Expr[any]) (any, error) {
	return i.Evaluate(expr)
}

func (i *Interpreter) Evaluate(expr Expr[any]) (any, error) {
	return expr.Accept(i)
}

func isTruthy(value any) bool {
	if value == nil {
		return false
	}
	if isBool(value) {
		return value.(bool)
	}
	return true
}

func isEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a == b
}

func isBool(value any) bool {
	_, ok := value.(bool)
	return ok
}

func isFloat(value any) bool {
	_, ok := value.(float64)
	return ok
}

func isString(value any) bool {
	_, ok := value.(string)
	return ok
}

func checkNumberOperand(operator Token, operand any) error {
	if isFloat(operand) {
		return nil
	}
	return Error(operator.Line, fmt.Sprintf("operand must be a number, got %T.", operand))
}

func checkNumberOperands(operator Token, left, right any) error {
	if isFloat(left) && isFloat(right) {
		return nil
	}
	return Error(operator.Line, fmt.Sprintf("operands must be numbers, got %T and %T.", left, right))
}
