package glox

type Program[R any] []Stmt[R]

type Visitor[R any] interface {
	ExprVisitor[R]
	StmtVisitor[R]
}

type Expr[R any] interface {
	Accept(visitor ExprVisitor[R]) (R, error)
}

type ExprVisitor[R any] interface {
	VisitBinaryExpr(expr *BinaryExpr[R]) (R, error)
	VisitGroupingExpr(expr *GroupingExpr[R]) (R, error)
	VisitLiteralExpr(expr *LiteralExpr[R]) (R, error)
	VisitUnaryExpr(expr *UnaryExpr[R]) (R, error)
	VisitVariableExpr(expr *VariableExpr[R]) (R, error)
	VisitAssignExpr(expr *AssignExpr[R]) (R, error)
	VisitLogicalExpr(expr *LogicalExpr[R]) (R, error)
}

type BinaryExpr[R any] struct {
	Left     Expr[R]
	Operator Token
	Right    Expr[R]
}

func (b *BinaryExpr[R]) Accept(visitor ExprVisitor[R]) (R, error) {
	return visitor.VisitBinaryExpr(b)
}

type GroupingExpr[R any] struct {
	Expression Expr[R]
}

func (g *GroupingExpr[R]) Accept(visitor ExprVisitor[R]) (R, error) {
	return visitor.VisitGroupingExpr(g)
}

type LiteralExpr[R any] struct {
	Value any
}

func (l *LiteralExpr[R]) Accept(visitor ExprVisitor[R]) (R, error) {
	return visitor.VisitLiteralExpr(l)
}

type UnaryExpr[R any] struct {
	Operator Token
	Right    Expr[R]
}

func (u *UnaryExpr[R]) Accept(visitor ExprVisitor[R]) (R, error) {
	return visitor.VisitUnaryExpr(u)
}

type VariableExpr[R any] struct {
	Name Token
}

func (i *VariableExpr[R]) Accept(visitor ExprVisitor[R]) (R, error) {
	return visitor.VisitVariableExpr(i)
}

type AssignExpr[R any] struct {
	Name  Token
	Value Expr[R]
}

func (a *AssignExpr[R]) Accept(visitor ExprVisitor[R]) (R, error) {
	return visitor.VisitAssignExpr(a)
}

type LogicalExpr[R any] struct {
	Left     Expr[R]
	Operator Token
	Right    Expr[R]
}

func (l *LogicalExpr[R]) Accept(visitor ExprVisitor[R]) (R, error) {
	return visitor.VisitLogicalExpr(l)
}

type Stmt[R any] interface {
	Accept(visitor StmtVisitor[R]) error
}

type StmtVisitor[R any] interface {
	VisitExpressionStmt(stmt *ExpressionStmt[R]) error
	VisitPrintStmt(stmt *PrintStmt[R]) error
	VisitVarDeclStmt(stmt *VarDeclStmt[R]) error
	VisitBlockStmt(stmt *BlockStmt[R]) error
	VisitIfStmt(stmt *IfStmt[R]) error
	VisitWhileStmt(stmt *WhileStmt[R]) error
}

type ExpressionStmt[R any] struct {
	Expr Expr[R]
}

func (e *ExpressionStmt[R]) Accept(visitor StmtVisitor[R]) error {
	return visitor.VisitExpressionStmt(e)
}

type PrintStmt[R any] struct {
	Expr Expr[R]
}

func (p *PrintStmt[R]) Accept(visitor StmtVisitor[R]) error {
	return visitor.VisitPrintStmt(p)
}

type VarDeclStmt[R any] struct {
	Name        Token
	Initializer Expr[R]
}

func (v *VarDeclStmt[R]) Accept(visitor StmtVisitor[R]) error {
	return visitor.VisitVarDeclStmt(v)
}

type BlockStmt[R any] struct {
	Statements []Stmt[R]
}

func (b *BlockStmt[R]) Accept(visitor StmtVisitor[R]) error {
	return visitor.VisitBlockStmt(b)
}

type IfStmt[R any] struct {
	Condition  Expr[R]
	ThenBranch Stmt[R]
	ElseBranch Stmt[R]
}

func (i *IfStmt[R]) Accept(visitor StmtVisitor[R]) error {
	return visitor.VisitIfStmt(i)
}

type WhileStmt[R any] struct {
	Condition Expr[R]
	Body      Stmt[R]
}

func (w *WhileStmt[R]) Accept(visitor StmtVisitor[R]) error {
	return visitor.VisitWhileStmt(w)
}
