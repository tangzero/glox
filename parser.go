package glox

import "slices"

type Parser struct {
	Tokens        []Token
	Current       int
	LoopDepth     int
	CallableDepth int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{Tokens: tokens}
}

func (p *Parser) Parse() (Program, error) {
	return p.Program()
}

// Program -> Declaration* EOF ;
func (p *Parser) Program() (Program, error) {
	var program Program
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
func (p *Parser) Declaration() (Stmt, error) {
	if p.Match(Fun) {
		return p.FunDeclaration("function")
	}
	if p.Match(Var) {
		return p.VarDeclaration()
	}
	return p.Statement()
}

// FunDeclaration -> "fun" IDENTIFIER "(" Parameters? ")" Block ;
func (p *Parser) FunDeclaration(kind string) (_ Stmt, err error) {
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
	p.CallableDepth++
	body, err := p.Block()
	if err != nil {
		return nil, err
	}
	p.CallableDepth--
	return &FunctionStmt{
		Name:   name,
		Params: parameters,
		Body:   body.Statements,
	}, nil
}

// Parameters -> IDENTIFIER ( "," IDENTIFIER )* ;
func (p *Parser) Parameters() ([]Token, error) {
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
func (p *Parser) VarDeclaration() (_ Stmt, err error) {
	if !p.Match(Identifier) {
		return nil, p.Error(p.Peek(), "expect variable name")
	}
	name := p.Previous()
	var initializer Expr
	if p.Match(Equal) {
		initializer, err = p.Expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after variable declaration")
	}
	return &VarDeclStmt{Name: name, Initializer: initializer}, nil
}

// Statement -> IfStatement
//
//			| WhileStatement
//			| ForStatement
//			| PrintStatement
//			| Block
//	    | BreakStatement
//	    | ContinueStatement
//	    | ReturnStatement
//			| ExpressionStatement ;
func (p *Parser) Statement() (Stmt, error) {
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
	if p.Match(Return) {
		return p.ReturnStatement()
	}
	return p.ExpressionStatement()
}

// IfStatement -> "if" "(" Expression ")" Statement ( "else" Statement )? ;
func (p *Parser) IfStatement() (Stmt, error) {
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
	var elseBranch Stmt
	if p.Match(Else) {
		elseBranch, err = p.Statement()
		if err != nil {
			return nil, err
		}
	}
	return &IfStmt{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

// WhileStatement -> "while" "(" Expression ")" Statement ;
func (p *Parser) WhileStatement() (Stmt, error) {
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

	p.LoopDepth++
	body, err := p.Statement()
	if err != nil {
		return nil, err
	}
	p.LoopDepth--

	return &WhileStmt{
		Condition: condition,
		Body:      body,
	}, nil
}

// ForStatement -> "for" "(" ( VarDeclaration | ExpressionStatement | ";" ) Expression? ";" Expression? ")" Statement ;
func (p *Parser) ForStatement() (_ Stmt, err error) {
	if !p.Match(LeftParen) {
		return nil, p.Error(p.Peek(), "expect '(' after 'for'")
	}
	var initializer Stmt
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
	var condition Expr = &LiteralExpr{Value: true} // default to true
	if !p.Check(Semicolon) {
		condition, err = p.Expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after loop condition")
	}
	var increment Expr
	if !p.Check(RightParen) {
		increment, err = p.Expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(RightParen) {
		return nil, p.Error(p.Peek(), "expect ')' after for clauses")
	}

	p.LoopDepth++
	body, err := p.Statement()
	if err != nil {
		return nil, err
	}
	p.LoopDepth--

	// desugar for loop into while loop
	if increment != nil {
		body = &BlockStmt{
			Statements: []Stmt{
				body,
				&ExpressionStmt{Expr: increment},
			},
		}
	}

	body = &WhileStmt{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = &BlockStmt{
			Statements: []Stmt{
				initializer,
				body,
			},
		}
	}

	return body, nil
}

// PrintStatement -> "print" Expression ";" ;
func (p *Parser) PrintStatement() (Stmt, error) {
	expr, err := p.Expression()
	if err != nil {
		return nil, err
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after value")
	}
	return &PrintStmt{Expr: expr}, nil
}

// ReturnStatement -> "return" Expression? ";" ;
func (p *Parser) ReturnStatement() (_ Stmt, err error) {
	if p.CallableDepth == 0 {
		return nil, p.Error(p.Previous(), "unexpected 'return' outside a function/method")
	}
	keyword := p.Previous()
	var value Expr
	if !p.Check(Semicolon) {
		value, err = p.Expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after return value")
	}
	return &ReturnStmt{Keyword: keyword, Value: value}, nil
}

// ExpressionStatement -> Expression ";" ;
func (p *Parser) ExpressionStatement() (Stmt, error) {
	expr, err := p.Expression()
	if err != nil {
		return nil, err
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after expression")
	}
	return &ExpressionStmt{Expr: expr}, nil
}

// Block -> "{" Declaration* "}" ;
func (p *Parser) Block() (*BlockStmt, error) {
	var statements []Stmt
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
	return &BlockStmt{Statements: statements}, nil
}

// BreakStatement -> "break" ";" ;
func (p *Parser) BreakStatement() (Stmt, error) {
	if p.LoopDepth == 0 {
		return nil, p.Error(p.Previous(), "unexpected 'break' outside a loop")
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after 'break'")
	}
	return &BreakStmt{}, nil
}

// ContinueStatement -> "continue" ";" ;
func (p *Parser) ContinueStatement() (Stmt, error) {
	if p.LoopDepth == 0 {
		return nil, p.Error(p.Previous(), "unexpected 'continue' outside a loop")
	}
	if !p.Match(Semicolon) {
		return nil, p.Error(p.Peek(), "expect ';' after 'continue'")
	}
	return &ContinueStmt{}, nil
}

// Expression -> Assignment ;
func (p *Parser) Expression() (Expr, error) {
	return p.Assignment()
}

// Assignment -> IDENTIFIER "=" Assignment | Lambda ;
func (p *Parser) Assignment() (Expr, error) {
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
		if varExpr, ok := expr.(*VariableExpr); ok {
			return &AssignExpr{Name: varExpr.Name, Value: value}, nil
		}
		return nil, p.Error(equals, "invalid assignment target")
	}
	return expr, nil
}

// Logical -> Equality ( ( "or" | "and" ) Equality )* ;
func (p *Parser) Logical() (zero Expr, _ error) {
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
		expr = &LogicalExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Equality -> Comparison ( ( "!=" | "==" ) Comparison )* ;
func (p *Parser) Equality() (zero Expr, _ error) {
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
		expr = &BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Comparison -> Term ( ( ">" | ">=" | "<" | "<=" ) Term )* ;
func (p *Parser) Comparison() (zero Expr, _ error) {
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
		expr = &BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Term -> Factor ( ( "-" | "+" ) Factor )* ;
func (p *Parser) Term() (zero Expr, _ error) {
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
		expr = &BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Factor -> Unary ( ( "/" | "*" ) Unary )* ;
func (p *Parser) Factor() (zero Expr, _ error) {
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
		expr = &BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr, nil
}

// Unary -> ( "!" | "-" ) Unary | Call ;
func (p *Parser) Unary() (zero Expr, _ error) {
	if p.Match(Bang, Minus) {
		operator := p.Previous()
		right, err := p.Unary()
		if err != nil {
			return zero, err
		}
		return &UnaryExpr{
			Operator: operator,
			Right:    right,
		}, nil
	}
	return p.Call()
}

// Call -> Primary ( "(" Arguments? ")" )* ;
func (p *Parser) Call() (zero Expr, _ error) {
	expr, err := p.Primary()
	if err != nil {
		return zero, err
	}
	for p.Match(LeftParen) {
		var arguments []Expr
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
		expr = &CallExpr{
			Callee:    expr,
			Paren:     paren,
			Arguments: arguments,
		}
	}
	return expr, nil
}

// Arguments -> Expression ( "," Expression )* ;
func (p *Parser) Arguments() ([]Expr, error) {
	var arguments []Expr
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

// Primary -> Lambda | NUMBER | STRING | "true" | "false" | "nil" | "(" Expression ")" | IDENTIFIER ;
func (p *Parser) Primary() (zero Expr, _ error) {
	if p.Match(Fun) {
		return p.Lambda()
	}
	if p.Match(False) {
		return &LiteralExpr{Value: false}, nil
	}
	if p.Match(True) {
		return &LiteralExpr{Value: true}, nil
	}
	if p.Match(Nil) {
		return &LiteralExpr{Value: nil}, nil
	}
	if p.Match(Number, String) {
		return &LiteralExpr{Value: p.Previous().Literal}, nil
	}
	if p.Match(Identifier) {
		return &VariableExpr{Name: p.Previous()}, nil
	}
	if p.Match(LeftParen) {
		expr, err := p.Expression()
		if err != nil {
			return zero, err
		}
		if !p.Match(RightParen) {
			return zero, p.Error(p.Peek(), "expect ')' after expression")
		}
		return &GroupingExpr{Expression: expr}, nil
	}
	return zero, p.Error(p.Peek(), "expect expression")
}

// Lambda -> "fun" "(" Parameters? ")" Block ;
func (p *Parser) Lambda() (_ Expr, err error) {
	if !p.Match(LeftParen) {
		return nil, p.Error(p.Peek(), "expect '(' after 'fun'")
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
		return nil, p.Error(p.Peek(), "expect '{' before function body")
	}
	p.CallableDepth++
	body, err := p.Block()
	if err != nil {
		return nil, err
	}
	p.CallableDepth--
	return &LambdaExpr{
		Params: parameters,
		Body:   body.Statements,
	}, nil
}

func (p *Parser) Check(t TokenType) bool {
	if p.IsAtEnd() {
		return false
	}
	return p.Peek().Type == t
}

func (p *Parser) Match(types ...TokenType) bool {
	return slices.ContainsFunc(types, func(t TokenType) bool {
		return p.Check(t) && func() bool { p.Advance(); return true }()
	})
}

func (p *Parser) Advance() Token {
	if !p.IsAtEnd() {
		p.Current++
	}
	return p.Previous()
}

func (p *Parser) IsAtEnd() bool {
	return p.Peek().Type == EOF
}

func (p *Parser) Peek() Token {
	return p.Tokens[p.Current]
}

func (p *Parser) Previous() Token {
	return p.Tokens[p.Current-1]
}

func (p *Parser) Error(token Token, message string) error {
	if token.Type == EOF {
		return Report(token.Line, " at end", message)
	}
	return Report(token.Line, " at '"+token.Lexeme+"'", message)
}
