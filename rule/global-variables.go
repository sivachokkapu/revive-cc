package rule

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"

	"github.com/mgechev/revive/lint"
)

// Global Variables Rule detects Global State Variables
type GlobalVariablesRule struct{}

// Apply applies the rule to given file.
func (r *GlobalVariablesRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintGlobalVariables{
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
func (r *GlobalVariablesRule) Name() string {
	return "global-variables"
}

type lintGlobalVariables struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

// Global Variables
var localVars []string
var globalVar bool

// AST Traversal
func (w lintGlobalVariables) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		ast.Inspect(n.Body, func(localNode ast.Node) bool {
			switch localNode := localNode.(type) {
			case *ast.GenDecl:
				if localNode.Tok == token.VAR {
					for _, varDecl := range localNode.Specs {
						ast.Inspect(varDecl, func(localVarNode ast.Node) bool {
							switch localVarNode := localVarNode.(type) {
							case *ast.ValueSpec:
								for _, name := range localVarNode.Names {
									localVars = append(localVars, nodeString(name))
								}
							}
							return true
						})
					}
				}
			}
			return true
		})
	case *ast.GenDecl:
		if n.Tok == token.VAR {
			for _, varDecl := range n.Specs {
				ast.Inspect(varDecl, func(varNode ast.Node) bool {
					switch varNode := varNode.(type) {
					case *ast.ValueSpec:
						for _, name := range varNode.Names {
							globalVar = true
							for _, localVar := range localVars {
								if nodeString(name) == localVar {
									globalVar = false
								}
							}
							if globalVar == true {
								w.onFailure(lint.Failure{
									Confidence: 1,
									Failure:    fmt.Sprintf("global variable detected: %s; should not use global variables, will lead to non-deterministic behaviour", name),
									Node:       n,
									Category:   "variables",
								})
							}
						}
					}
					return true
				})
			}
		}
	}

	return w
}

// Returns string representation of node
func nodeStringGlobalVars(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
