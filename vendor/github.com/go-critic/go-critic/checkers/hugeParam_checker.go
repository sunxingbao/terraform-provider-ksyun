package checkers

import (
	"go/ast"

	"github.com/go-lintpack/lintpack"
	"github.com/go-lintpack/lintpack/astwalk"
)

func init() {
	var info lintpack.CheckerInfo
	info.Name = "hugeParam"
	info.Tags = []string{"performance"}
	info.Params = lintpack.CheckerParams{
		"sizeThreshold": {
			Value: 80,
			Usage: "size in bytes that makes the warning trigger",
		},
	}
	info.Summary = "Detects params that incur excessive amount of copying"
	info.Before = `func f(x [1024]int) {}`
	info.After = `func f(x *[1024]int) {}`

	collection.AddChecker(&info, func(ctx *lintpack.CheckerContext) lintpack.FileWalker {
		return astwalk.WalkerForFuncDecl(&hugeParamChecker{
			ctx:           ctx,
			sizeThreshold: int64(info.Params.Int("sizeThreshold")),
		})
	})
}

type hugeParamChecker struct {
	astwalk.WalkHandler
	ctx *lintpack.CheckerContext

	sizeThreshold int64
}

func (c *hugeParamChecker) VisitFuncDecl(decl *ast.FuncDecl) {
	// TODO(quasilyte): maybe it's worthwhile to permit skipping
	// test files for this checker?
	if decl.Recv != nil {
		c.checkParams(decl.Recv.List)
	}
	c.checkParams(decl.Type.Params.List)
}

func (c *hugeParamChecker) checkParams(params []*ast.Field) {
	for _, p := range params {
		for _, id := range p.Names {
			typ := c.ctx.TypesInfo.TypeOf(id)
			size := c.ctx.SizesInfo.Sizeof(typ)
			if size >= c.sizeThreshold {
				c.warn(id, size)
			}
		}
	}
}

func (c *hugeParamChecker) warn(cause *ast.Ident, size int64) {
	c.ctx.Warn(cause, "%s is heavy (%d bytes); consider passing it by pointer",
		cause, size)
}
