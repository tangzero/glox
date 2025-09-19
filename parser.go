package glox

import "slices"

type Parser[R any] struct {
	Tokens     []Token
	Current    int
	LoopScopes int
}

func NewParser[R any](tokens []Token) *Parser[R] {
	return &Parser[R]{Tokens: tokens}
}

func (p *Parser[R]) Parse() (Program[R], error) {
	return p.Program()
}

// Program -> Declaration* EOF ;
func (p *Parser[R]) Program() (Program[R], error) {
	var program Program[R]
	for !p.IsAtEnd() {
		stmt, err := p.Declaration()
		if err != nil {
			return nil, err
		}
		program = append(program, stmt)
	}
	return program, nil
}

// Declaration -> FunDeclaration | VarDeclaration | Statement ;
func (p *Parser[R]) Declaration() (Stmt[R], error) {
	if p.Match(Fun) {
		return p.FunDeclaration("function")
	}
	if p.Match(Var) {
		return p.VarDeclaration()
	}
	return p.Statement()
}

// FunDeclaration -> "fun" IDENTIFIER "(" Parameters? ")" Block ;
func (p *Parser[R]) FunDeclaration(kind string) (_ Stmt[R], err error) {
	if !p.Match(Identifier) {
		return nil, p.Error(p.Peek(), "expect "+kind+" name")
	}
	name := p.Previous()
	if !p.Match(LeftParen) {
		return nil, p.Error(p.Peek(), "expect '(' after function name")
	}
	var parameters []Token
	if !p.Check(RightParen) {
		parameters, err = p.Parameters()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(RightParen) {
		return nil, p.Error(p.Peek(), "expect ')' after parameters")
	}
	if !p.Match(LeftBrace) {
		return nil, p.Error(p.Peek(), "expect '{' before "+kind+" body")
	}
	body, err := p.Block()
	if err != nil {
		return nil, err
	}
	return &FunctionStmt[R]{
		Name:   name,
		Params: parameters,
		Body:   body.Statements,
	}, nil
}

// Parameters -> IDENTIFIER ( "," IDENTIFIER )* ;
func (p *Parser[R]) Parameters() ([]Token, error) {
	var parameters []Token
	for {
		if len(parameters) >= 255 {
			return nil, p.Error(p.Peek(), "can't have more than 255 parameters")
		}
		if !p.Match(Identifier) {
			return nil, p.Error(p.Peek(), "expect parameter name")
		}
		parameters = append(parameters, p.Previous())
		if !p.Match(Comma) {
			break
		}
	}
	return parameters, nil
}

// VarDeclaration -> "var" IDENTIFIER ( "=" Expression )? ";" ;
func (p *Parser[R]) VarDeclaration() (_ Stmt[R], err error) {
	if !p.Match(Identifier) {
		return nil, p.Error(p.Peek(), "expect variable name")
	}
	name := p.Previous()
	var initializer Expr[R]
	if p.Match(Equal) {
		initializer, err = p.Expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after variable declaration")
	}
	return &VarDeclStmt[R]{Name: name, Initializer: initializer}, nil
}

// Statement -> IfStatement
//
//			| WhileStatement
//			| ForStatement
//			| PrintStatement
//			| Block
//	    | BreakStatement
//	    | ContinueStatement
//			| ExpressionStatement ;
func (p *Parser[R]) Statement() (Stmt[R], error) {
	if p.Match(If) {
		return p.IfStatement()
	}
	if p.Match(While) {
		return p.WhileStatement()
	}
	if p.Match(For) {
		return p.ForStatement()
	}
	if p.Match(Print) {
		return p.PrintStatement()
	}
	if p.Match(LeftBrace) {
		return p.Block()
	}
	if p.Match(Break) {
		return p.BreakStatement()
	}
	if p.Match(Continue) {
		return p.ContinueStatement()
	}
	return p.ExpressionStatement()
}

// IfStatement -> "if" "(" Expression ")" Statement ( "else" Statement )? ;
func (p *Parser[R]) IfStatement() (Stmt[R], error) {
	if !p.Match(LeftParen) {
		return nil, p.Error(p.Peek(), "expect '(' after 'if'")
	}
	condition, err := p.Expression()
	if err != nil {
		return nil, err
	}
	if !p.Match(RightParen) {
		return nil, p.Error(p.Peek(), "expect ')' after if condition")
	}
	thenBranch, err := p.Statement()
	if err != nil {
		return nil, err
	}
	var elseBranch Stmt[R]
	if p.Match(Else) {
		elseBranch, err = p.Statement()
		if err != nil {
			return nil, err
		}
	}
	return &IfStmt[R]{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

// WhileStatement -> "while" "(" Expression ")" Statement ;
func (p *Parser[R]) WhileStatement() (Stmt[R], error) {
	if !p.Match(LeftParen) {
		return nil, p.Error(p.Peek(), "expect '(' after 'while'")
	}
	condition, err := p.Expression()
	if err != nil {
		return nil, err
	}
	if !p.Match(RightParen) {
		return nil, p.Error(p.Peek(), "expect ')' after condition")
	}

	p.LoopScopes++
	body, err := p.Statement()
	if err != nil {
		return nil, err
	}
	p.LoopScopes--

	return &WhileStmt[R]{
		Condition: condition,
		Body:      body,
	}, nil
}

// ForStatement -> "for" "(" ( VarDeclaration | ExpressionStatement | ";" ) Expression? ";" Expression? ")" Statement ;
func (p *Parser[R]) ForStatement() (_ Stmt[R], err error) {
	if !p.Match(LeftParen) {
		return nil, p.Error(p.Peek(), "expect '(' after 'for'")
	}
	var initializer Stmt[R]
	if p.Match(Semicolon) {
		initializer = nil // not needed, but explicit
	} else if p.Match(Var) {
		initializer, err = p.VarDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = p.ExpressionStatement()
		if err != nil {
			return nil, err
		}
	}
	var condition Expr[R] = &LiteralExpr[R]{Value: true} // default to true
	if !p.Check(Semicolon) {
		condition, err = p.Expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after loop condition")
	}
	var increment Expr[R]
	if !p.Check(RightParen) {
		increment, err = p.Expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(RightParen) {
		return nil, p.Error(p.Peek(), "expect ')' after for clauses")
	}

	p.LoopScopes++
	body, err := p.Statement()
	if err != nil {
		return nil, err
	}
	p.LoopScopes--

	// desugar for loop into while loop
	if increment != nil {
		body = &BlockStmt[R]{
			Statements: []Stmt[R]{
				body,
				&ExpressionStmt[R]{Expr: increment},
			},
		}
	}

	body = &WhileStmt[R]{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = &BlockStmt[R]{
			Statements: []Stmt[R]{
				initializer,
				body,
			},
		}
	}

	return body, nil
}

// PrintStatement -> "print" Expression ";" ;
func (p *Parser[R]) PrintStatement() (Stmt[R], error) {
	expr, err := p.Expression()
	if err != nil {
		return nil, err
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after value")
	}
	return &PrintStmt[R]{Expr: expr}, nil
}

// ExpressionStatement -> Expression ";" ;
func (p *Parser[R]) ExpressionStatement() (Stmt[R], error) {
	expr, err := p.Expression()
	if err != nil {
		return nil, err
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after expression")
	}
	return &ExpressionStmt[R]{Expr: expr}, nil
}

// Block -> "{" Declaration* "}" ;
func (p *Parser[R]) Block() (*BlockStmt[R], error) {
	var statements []Stmt[R]
	for !p.Check(RightBrace) && !p.IsAtEnd() {
		stmt, err := p.Declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}
	if !p.Match(RightBrace) {
		return nil, p.Error(p.Peek(), "expect '}' after block")
	}
	return &BlockStmt[R]{Statements: statements}, nil
}

// BreakStatement -> "break" ";" ;
func (p *Parser[R]) BreakStatement() (Stmt[R], error) {
	if p.LoopScopes == 0 {
		return nil, p.Error(p.Previous(), "unexpected 'break' outside a loop")
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after 'break'")
	}
	return &BreakStmt[R]{}, nil
}

// ContinueStatement -> "continue" ";" ;
func (p *Parser[R]) ContinueStatement() (Stmt[R], error) {
	if p.LoopScopes == 0 {
		return nil, p.Error(p.Previous(), "unexpected 'continue' outside a loop")
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after 'continue'")
	}
	return &ContinueStmt[R]{}, nil
}

// Expression -> Assignment ;
func (p *Parser[R]) Expression() (Expr[R], error) {
	return p.Assignment()
}

// Assignment -> IDENTIFIER "=" Assignment | Logical ;
func (p *Parser[R]) Assignment() (Expr[R], error) {
	expr, err := p.Logical()
	if err != nil {
		return nil, err
	}
	if p.Match(Equal) {
		equals := p.Previous()
		value, err := p.Assignment()
		if err != nil {
			return nil, err
		}
		if varExpr, ok := expr.(*VariableExpr[R]); ok {
			return &AssignExpr[R]{Name: varExpr.Name, Value: value}, nil
		}
		return nil, p.Error(equals, "invalid assignment target")
	}
	return expr, nil
}

// Logical -> Equality ( ( "or" | "and" ) Equality )* ;
func (p *Parser[R]) Logical() (zero Expr[R], _ error) {
	expr, err := p.Equality()
	if err != nil {
		return zero, err
	}
	for p.Match(Or, And) {
		operator := p.Previous()
		right, err := p.Equality()
		if err != nil {
			return zero, err
		}
		expr = &LogicalExpr[R]{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Equality -> Comparison ( ( "!=" | "==" ) Comparison )* ;
func (p *Parser[R]) Equality() (zero Expr[R], _ error) {
	expr, err := p.Comparison()
	if err != nil {
		return zero, err
	}
	for p.Match(BangEqual, EqualEqual) {
		operator := p.Previous()
		right, err := p.Comparison()
		if err != nil {
			return zero, err
		}
		expr = &BinaryExpr[R]{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Comparison -> Term ( ( ">" | ">=" | "<" | "<=" ) Term )* ;
func (p *Parser[R]) Comparison() (zero Expr[R], _ error) {
	expr, err := p.Term()
	if err != nil {
		return zero, err
	}
	for p.Match(Greater, GreaterEqual, Less, LessEqual) {
		operator := p.Previous()
		right, err := p.Term()
		if err != nil {
			return zero, err
		}
		expr = &BinaryExpr[R]{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Term -> Factor ( ( "-" | "+" ) Factor )* ;
func (p *Parser[R]) Term() (zero Expr[R], _ error) {
	expr, err := p.Factor()
	if err != nil {
		return zero, err
	}
	for p.Match(Minus, Plus) {
		operator := p.Previous()
		right, err := p.Factor()
		if err != nil {
			return zero, err
		}
		expr = &BinaryExpr[R]{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Factor -> Unary ( ( "/" | "*" ) Unary )* ;
func (p *Parser[R]) Factor() (zero Expr[R], _ error) {
	expr, err := p.Unary()
	if err != nil {
		return zero, err
	}
	for p.Match(Slash, Star) {
		operator := p.Previous()
		right, err := p.Unary()
		if err != nil {
			return zero, err
		}
		expr = &BinaryExpr[R]{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Unary -> ( "!" | "-" ) Unary | Call ;
func (p *Parser[R]) Unary() (zero Expr[R], _ error) {
	if p.Match(Bang, Minus) {
		operator := p.Previous()
		right, err := p.Unary()
		if err != nil {
			return zero, err
		}
		return &UnaryExpr[R]{
			Operator: operator,
			Right:    right,
		}, nil
	}
	return p.Call()
}

// Call -> Primary ( "(" Arguments? ")" )* ;
func (p *Parser[R]) Call() (zero Expr[R], _ error) {
	expr, err := p.Primary()
	if err != nil {
		return zero, err
	}
	for p.Match(LeftParen) {
		var arguments []Expr[R]
		if !p.Check(RightParen) {
			arguments, err = p.Arguments()
			if err != nil {
				return zero, err
			}
		}
		if !p.Match(RightParen) {
			return zero, p.Error(p.Peek(), "expect ')' after arguments")
		}
		paren := p.Previous()
		expr = &CallExpr[R]{
			Callee:    expr,
			Paren:     paren,
			Arguments: arguments,
		}
	}
	return expr, nil
}

// Arguments -> Expression ( "," Expression )* ;
func (p *Parser[R]) Arguments() ([]Expr[R], error) {
	var arguments []Expr[R]
	for {
		if len(arguments) >= 255 {
			return nil, p.Error(p.Peek(), "can't have more than 255 arguments")
		}
		arg, err := p.Expression()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, arg)
		if !p.Match(Comma) {
			break
		}
	}
	return arguments, nil
}

// Primary -> NUMBER | STRING | "true" | "false" | "nil" | "(" Expression ")" | IDENTIFIER ;
func (p *Parser[R]) Primary() (zero Expr[R], _ error) {
	if p.Match(False) {
		return &LiteralExpr[R]{Value: false}, nil
	}
	if p.Match(True) {
		return &LiteralExpr[R]{Value: true}, nil
	}
	if p.Match(Nil) {
		return &LiteralExpr[R]{Value: nil}, nil
	}
	if p.Match(Number, String) {
		return &LiteralExpr[R]{Value: p.Previous().Literal}, nil
	}
	if p.Match(Identifier) {
		return &VariableExpr[R]{Name: p.Previous()}, nil
	}
	if p.Match(LeftParen) {
		expr, err := p.Expression()
		if err != nil {
			return zero, err
		}
		if !p.Match(RightParen) {
			return zero, p.Error(p.Peek(), "expect ')' after expression")
		}
		return &GroupingExpr[R]{Expression: expr}, nil
	}
	return zero, p.Error(p.Peek(), "expect expression")
}

func (p *Parser[R]) Check(t TokenType) bool {
	if p.IsAtEnd() {
		return false
	}
	return p.Peek().Type == t
}

func (p *Parser[R]) Match(types ...TokenType) bool {
	return slices.ContainsFunc(types, func(t TokenType) bool {
		return p.Check(t) && func() bool { p.Advance(); return true }()
	})
}

func (p *Parser[R]) Advance() Token {
	if !p.IsAtEnd() {
		p.Current++
	}
	return p.Previous()
}

func (p *Parser[R]) IsAtEnd() bool {
	return p.Peek().Type == EOF
}

func (p *Parser[R]) Peek() Token {
	return p.Tokens[p.Current]
}

func (p *Parser[R]) Previous() Token {
	return p.Tokens[p.Current-1]
}

func (p *Parser[R]) Error(token Token, message string) error {
	if token.Type == EOF {
		return Report(token.Line, " at end", message)
	}
	return Report(token.Line, " at '"+token.Lexeme+"'", message)
}
