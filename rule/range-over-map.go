package rule

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"

	"github.com/mgechev/revive/lint"
)

// Range over map rule detects range map iterations
type RangeOverMapRule struct{}

// Apply applies the rule to given file.
func (r *RangeOverMapRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintRangeOverMap{
		file:    file,
		fileAst: fileAst,
		onFailure: func(failure lint.Failure) {
			failures = append(failures, failure)
		},
	}

	file.Pkg.TypeCheck()
	ast.Walk(walker, fileAst)

	return failures
}

// Name returns the rule name.
func (r *RangeOverMapRule) Name() string {
	return "range-over-map"
}

type lintRangeOverMap struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

// Global Variables
var rangeOverMapName string
var mapString string

// AST traversal logic
func (w lintRangeOverMap) Visit(node ast.Node) ast.Visitor {
	f := w.file

	rangeOverMapName = ""
	mapString = ""

	switch n := node.(type) {
	case *ast.RangeStmt:
		rangeOverMapName = types.ExprString(n.X)

		ast.Inspect(node, func(x ast.Node) bool {
			if expr, ok := x.(ast.Expr); ok {
				if tv, ok := f.Pkg.TypesInfo.Types[expr]; ok {
					if len(tv.Type.String()) > 3 {
						mapString = tv.Type.String()[0:3]
					}

					if rangeOverMapName == nodeStringRangeOverMap(expr) && n.X.Pos() == expr.Pos() && mapString == "map" {
						w.onFailure(lint.Failure{
							Confidence: 1,
							Failure:    "should not use range over map, will lead to non-deterministic behaviour",
							Node:       n,
							Category:   "control flow",
						})
					}
				}
			}
			return true
		})
	}
	return w
}

// Returns a string representation of a node
func nodeStringRangeOverMap(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
