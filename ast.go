package glox

type Program []Stmt

type Visitor interface {
	ExprVisitor
	StmtVisitor
}

type Expr interface {
	Accept(visitor ExprVisitor) (any, error)
}

type ExprVisitor interface {
	VisitBinaryExpr(expr *BinaryExpr) (any, error)
	VisitGroupingExpr(expr *GroupingExpr) (any, error)
	VisitLiteralExpr(expr *LiteralExpr) (any, error)
	VisitUnaryExpr(expr *UnaryExpr) (any, error)
	VisitVariableExpr(expr *VariableExpr) (any, error)
	VisitAssignExpr(expr *AssignExpr) (any, error)
	VisitLogicalExpr(expr *LogicalExpr) (any, error)
	VisitCallExpr(expr *CallExpr) (any, error)
	VisitLambdaExpr(expr *LambdaExpr) (any, error)
}

type BinaryExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (b *BinaryExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitBinaryExpr(b)
}

type GroupingExpr struct {
	Expression Expr
}

func (g *GroupingExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(g)
}

type LiteralExpr struct {
	Value any
}

func (l *LiteralExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLiteralExpr(l)
}

type UnaryExpr struct {
	Operator Token
	Right    Expr
}

func (u *UnaryExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitUnaryExpr(u)
}

type VariableExpr struct {
	Name Token
}

func (i *VariableExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitVariableExpr(i)
}

type AssignExpr struct {
	Name  Token
	Value Expr
}

func (a *AssignExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitAssignExpr(a)
}

type LogicalExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (l *LogicalExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLogicalExpr(l)
}

type CallExpr struct {
	Callee    Expr
	Paren     Token
	Arguments []Expr
}

func (c *CallExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitCallExpr(c)
}

type LambdaExpr struct {
	Params []Token
	Body   []Stmt
}

func (l *LambdaExpr) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLambdaExpr(l)
}

type Stmt interface {
	Accept(visitor StmtVisitor) error
}

type StmtVisitor interface {
	VisitExpressionStmt(stmt *ExpressionStmt) error
	VisitPrintStmt(stmt *PrintStmt) error
	VisitVarDeclStmt(stmt *VarDeclStmt) error
	VisitBlockStmt(stmt *BlockStmt) error
	VisitIfStmt(stmt *IfStmt) error
	VisitWhileStmt(stmt *WhileStmt) error
	VisitBreakStmt(stmt *BreakStmt) error
	VisitContinueStmt(stmt *ContinueStmt) error
	VisitFunctionStmt(stmt *FunctionStmt) error
	VisitReturnStmt(stmt *ReturnStmt) error
}

type ExpressionStmt struct {
	Expr Expr
}

func (e *ExpressionStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitExpressionStmt(e)
}

type PrintStmt struct {
	Expr Expr
}

func (p *PrintStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitPrintStmt(p)
}

type VarDeclStmt struct {
	Name        Token
	Initializer Expr
}

func (v *VarDeclStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitVarDeclStmt(v)
}

type BlockStmt struct {
	Statements []Stmt
}

func (b *BlockStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitBlockStmt(b)
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (i *IfStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitIfStmt(i)
}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (w *WhileStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitWhileStmt(w)
}

type BreakStmt struct{}

func (b *BreakStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitBreakStmt(b)
}

type ContinueStmt struct{}

func (c *ContinueStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitContinueStmt(c)
}

type FunctionStmt struct {
	Name   Token
	Params []Token
	Body   []Stmt
}

func (f *FunctionStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitFunctionStmt(f)
}

type ReturnStmt struct {
	Keyword Token
	Value   Expr
}

func (r *ReturnStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitReturnStmt(r)
}
