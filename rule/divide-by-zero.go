package rule

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/mgechev/revive/lint"
)

type DivideByZeroRule struct{}

// Apply applies the rule to given file.
func (r *DivideByZeroRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintDivideByZero{
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
func (r *DivideByZeroRule) Name() string {
	return "divide-by-zero"
}

type lintDivideByZero struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

func (w lintDivideByZero) Visit(n ast.Node) ast.Visitor {
	be, ok := n.(*ast.BinaryExpr)
	if !ok {
		return w
	}
	operator := be.Op
	leftVal := types.ExprString(be.Y)

	if operator == token.QUO && leftVal == "0" {
		w.onFailure(lint.Failure{
			Confidence: 1,
			Failure:    "should not divide by zero",
			Node:       be,
			Category:   "logic",
		})
		return w
	}
	return w
}
