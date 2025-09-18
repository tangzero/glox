package glox

import (
	"fmt"
)

type Interpreter struct {
	env Env
}

func NewInterpreter() *Interpreter {
	return &Interpreter{env: NewEnvironment(nil)}
}

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
	default:
		return nil, Error(expr.Operator.Line, "unknown binary operator.")
	}
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
	default:
		return nil, Error(expr.Operator.Line, "unknown unary operator.")
	}
}

func (i *Interpreter) VisitVariableExpr(expr *VariableExpr[any]) (any, error) {
	return i.env.Get(expr.Name)
}

func (i *Interpreter) VisitAssignExpr(expr *AssignExpr[any]) (any, error) {
	value, err := i.Evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	return value, i.env.Assign(expr.Name, value)
}

func (i *Interpreter) VisitLogicalExpr(expr *LogicalExpr[any]) (any, error) {
	left, err := i.Evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.Type == Or {
		if isTruthy(left) {
			return left, nil
		}
	} else {
		if !isTruthy(left) {
			return left, nil
		}
	}

	return i.Evaluate(expr.Right)
}

func (i *Interpreter) Evaluate(expr Expr[any]) (any, error) {
	return expr.Accept(i)
}

func (i *Interpreter) VisitExpressionStmt(stmt *ExpressionStmt[any]) error {
	_, err := i.Evaluate(stmt.Expr)
	return err
}

func (i *Interpreter) VisitPrintStmt(stmt *PrintStmt[any]) error {
	value, err := i.Evaluate(stmt.Expr)
	if err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}

func (i *Interpreter) VisitVarDeclStmt(stmt *VarDeclStmt[any]) error {
	if stmt.Initializer == nil {
		i.env.Define(stmt.Name, nil)
		return nil
	}

	value, err := i.Evaluate(stmt.Initializer)
	if err != nil {
		return err
	}
	i.env.Define(stmt.Name, value)
	return nil
}

func (i *Interpreter) VisitBlockStmt(stmt *BlockStmt[any]) error {
	return i.ExecuteBlock(stmt.Statements, NewEnvironment(i.env))
}

func (i *Interpreter) VisitIfStmt(stmt *IfStmt[any]) error {
	condition, err := i.Evaluate(stmt.Condition)
	if err != nil {
		return err
	}
	if isTruthy(condition) {
		return i.Execute(stmt.ThenBranch)
	}
	if stmt.ElseBranch != nil {
		return i.Execute(stmt.ElseBranch)
	}
	return nil
}

func (i *Interpreter) VisitWhileStmt(stmt *WhileStmt[any]) error {
	for {
		condition, err := i.Evaluate(stmt.Condition)
		if err != nil {
			return err
		}
		if !isTruthy(condition) {
			break
		}
		if err := i.Execute(stmt.Body); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) VisitForStmt(stmt *ForStmt[any]) error {
	// using a new environment to avoid variable leakage
	previously := i.env
	i.env = NewEnvironment(i.env)
	defer func() { i.env = previously }()

	if stmt.Initializer != nil {
		if err := i.Execute(stmt.Initializer); err != nil {
			return err
		}
	}
	for {
		if stmt.Condition != nil {
			condition, err := i.Evaluate(stmt.Condition)
			if err != nil {
				return err
			}
			if !isTruthy(condition) {
				break
			}
		}
		if err := i.Execute(stmt.Body); err != nil {
			return err
		}
		if stmt.Increment != nil {
			if _, err := i.Evaluate(stmt.Increment); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *Interpreter) ExecuteBlock(statements []Stmt[any], env Env) error {
	previous := i.env
	i.env = env
	defer func() { i.env = previous }()
	for _, stmt := range statements {
		if err := i.Execute(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) Execute(stmt Stmt[any]) error {
	return stmt.Accept(i)
}

func (i *Interpreter) Interpret(program Program[any]) error {
	for _, stmt := range program {
		if err := i.Execute(stmt); err != nil {
			return err
		}
	}
	return nil
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
