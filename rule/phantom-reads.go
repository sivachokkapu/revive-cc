package rule

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"

	"github.com/mgechev/revive/lint"
)

// Phantom reads rule detects phantom reads of ledger
type PhantomReadsRule struct{}

// Apply applies the rule to given file.
func (r *PhantomReadsRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintPhantomReads{
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
func (r *PhantomReadsRule) Name() string {
	return "phantom-reads"
}

type lintPhantomReads struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

// Global Variables
var writeQueryOrKey string
var getQueryOrKey string

// AST Traversal
func (w lintPhantomReads) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		writeKey = ""
		getQueryOrKey = ""

		ast.Inspect(n.Body, func(phantomReadNode ast.Node) bool {
			switch phantomReadNode := phantomReadNode.(type) {
			case *ast.CallExpr:
				functionExpression := nodeStringPhantomReads(phantomReadNode.Fun)

				if strings.Contains(functionExpression, ".") {
					functionCall := strings.Split(functionExpression, ".")

					if functionCall[1] == "GetHistoryForKey" || functionCall[1] == "GetQueryResult" {
						getQueryOrKey = nodeStringPhantomReads(phantomReadNode.Args[0])

						ast.Inspect(n.Body, func(writeNode ast.Node) bool {
							switch writeNode := writeNode.(type) {
							case *ast.CallExpr:
								functionExpression = nodeStringPhantomReads(writeNode.Fun)

								if strings.Contains(functionExpression, ".") {
									functionCall := strings.Split(functionExpression, ".")

									if functionCall[1] == "PutState" {
										writeQueryOrKey = nodeStringPhantomReads(writeNode.Args[0])

										if writeNode.Pos() > phantomReadNode.Pos() && writeQueryOrKey == getQueryOrKey {
											w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "data obtained from phantom reads should not be used to write new data or update data on the ledger: write",
												Node:       phantomReadNode,
												Category:   "control flow",
											})
											w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "data obtained from phantom reads should not be used to write new data or update data on the ledger: read",
												Node:       writeNode,
												Category:   "control flow",
											})

											return true
										}
									}
								}
							}
							return true
						})
					}
				}
			}
			return true
		})

	}
	return w
}

// Returns the string representation of a node
func nodeStringPhantomReads(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
