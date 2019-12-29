package rule

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/mgechev/revive/lint"
)

// Blacklisted chaincode imports rule detects blacklisted import
type BlacklistedChaincodeImportsRule struct{}

// Apply applies the rule to given file.
func (r *BlacklistedChaincodeImportsRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintBlacklistedChaincodeImports{
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
func (r *BlacklistedChaincodeImportsRule) Name() string {
	return "blacklisted-chaincode-imports"
}

type lintBlacklistedChaincodeImports struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

// AST Traversal
func (w lintBlacklistedChaincodeImports) Visit(node ast.Node) ast.Visitor {
	var blacklistedImports = []string{"time", "math/rand", "net", "os"}

	switch importNode := node.(type) {
	case *ast.ImportSpec:
		for _, blacklistedImport := range blacklistedImports {
			if strings.Contains(importNode.Path.Value, blacklistedImport) && strings.Contains(importNode.Path.Value, "fabric") == false {
				w.onFailure(lint.Failure{
					Confidence: 1,
					Failure:    fmt.Sprintf("should not use the following blacklisted import: %s", importNode.Path.Value),
					Node:       importNode,
					Category:   "imports",
				})
				return w
			}
		}
	}
	return w

}
