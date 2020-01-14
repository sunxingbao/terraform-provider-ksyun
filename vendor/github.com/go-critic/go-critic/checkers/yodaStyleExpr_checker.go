package checkers

import (
	"go/ast"
	"go/token"

	"github.com/go-lintpack/lintpack"
	"github.com/go-lintpack/lintpack/astwalk"
	"github.com/go-toolsmith/astcopy"
	"github.com/go-toolsmith/astp"
)

func init() {
	var info lintpack.CheckerInfo
	info.Name = "yodaStyleExpr"
	info.Tags = []string{"style", "experimental"}
	info.Summary = "Detects Yoda style expressions and suggests to replace them"
	info.Before = `return nil != ptr`
	info.After = `return ptr != nil`

	collection.AddChecker(&info, func(ctx *lintpack.CheckerContext) lintpack.FileWalker {
		return astwalk.WalkerForLocalExpr(&yodaStyleExprChecker{ctx: ctx})
	})
}

type yodaStyleExprChecker struct {
	astwalk.WalkHandler
	ctx *lintpack.CheckerContext
}

func (c *yodaStyleExprChecker) VisitLocalExpr(expr ast.Expr) {
	binexpr, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return
	}
	if binexpr.Op == token.EQL || binexpr.Op == token.NEQ {
		if c.isConstExpr(binexpr.X) && !c.isConstExpr(binexpr.Y) {
			c.warn(binexpr)
		}
	}
}

func (c *yodaStyleExprChecker) isConstExpr(expr ast.Expr) bool {
	return qualifiedName(expr) == "nil" || astp.IsBasicLit(expr)
}

func (c *yodaStyleExprChecker) warn(expr *ast.BinaryExpr) {
	e := astcopy.BinaryExpr(expr)
	e.X, e.Y = e.Y, e.X
	c.ctx.Warn(expr, "consider to change order in expression to %s", e)
}
