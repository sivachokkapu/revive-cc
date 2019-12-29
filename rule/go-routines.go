package rule

import (
	"go/ast"

	"github.com/mgechev/revive/lint"
)

// Goroutines Rule detects the use of goroutines
type GoRoutinesRule struct{}

// Apply applies the rule to given file.
func (r *GoRoutinesRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintGoRoutines{
		file:    file,
		fileAst: fileAst,
		onFailure: func(failure lint.Failure) {
			failures = append(failures, failure)
		},
	}

	ast.Walk(walker, fileAst)

	return failures
}

// Name returns the rule name.
func (r *GoRoutinesRule) Name() string {
	return "go-routines"
}

type lintGoRoutines struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

//AST Traversal
func (w lintGoRoutines) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.GoStmt:
		w.onFailure(lint.Failure{
			Confidence: 1,
			Failure:    "should not use goroutines, will lead to non-deterministic behaviour",
			Node:       n,
			Category:   "goroutines",
		})
		return w
	}
	return w
}
