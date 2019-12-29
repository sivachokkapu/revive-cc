package rule

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"

	"github.com/mgechev/revive/lint"
)

// Read after write rule detects read after write calls
type ReadAfterWriteRule struct{}

// Apply applies the rule to given file.
func (r *ReadAfterWriteRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintReadAfterWrite{
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
func (r *ReadAfterWriteRule) Name() string {
	return "read-after-write"
}

type lintReadAfterWrite struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

// Global Variables
var writeKey string
var readKey string

// AST Traversal
func (w lintReadAfterWrite) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		writeKey = ""
		readKey = ""

		ast.Inspect(n.Body, func(writeNode ast.Node) bool {
			switch writeNode := writeNode.(type) {
			case *ast.CallExpr:
				functionExpression := nodeString(writeNode.Fun)

				if strings.Contains(functionExpression, ".") {
					functionCall := strings.Split(functionExpression, ".")

					if functionCall[1] == "PutState" {
						writeKey = nodeString(writeNode.Args[0])

						ast.Inspect(n.Body, func(readNode ast.Node) bool {
							switch readNode := readNode.(type) {
							case *ast.CallExpr:
								functionExpression = nodeString(readNode.Fun)

								if strings.Contains(functionExpression, ".") {
									functionCall := strings.Split(functionExpression, ".")

									if functionCall[1] == "GetState" {
										readKey = nodeString(readNode.Args[0])

										if readNode.Pos() > writeNode.Pos() && readKey == writeKey {
											w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "should not read after write: write",
												Node:       writeNode,
												Category:   "control flow",
											})
											w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "should not read after write: read",
												Node:       readNode,
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

//Returns string representation of a node
func nodeString(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
