// Package goast implements cyclomatic complexity analysis for Go files using
// the standard library's go/ast parser. It does not shell out to external tools,
// which keeps Go analysis fast and portable inside CI and Docker.
package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Analyzer implements interfaces.Analyzer for Go source files.
type Analyzer struct{}

// New returns a ready-to-use Go AST Analyzer.
func New() *Analyzer { return &Analyzer{} }

// Name identifies this analyzer.
func (a *Analyzer) Name() string { return "go/ast" }

// Supports returns true for .go source files and false for generated test files.
func (a *Analyzer) Supports(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".go"
}

// Analyze computes cyclomatic complexity for each function in oldContent and
// newContent, then returns the delta for every function seen in either version.
func (a *Analyzer) Analyze(path, oldContent, newContent string) ([]interfaces.FunctionDelta, error) {
	oldScores, err := a.scoreContent(path, oldContent)
	if err != nil {
		return nil, fmt.Errorf("go ast (old) %s: %w", path, err)
	}

	newScores, err := a.scoreContent(path, newContent)
	if err != nil {
		return nil, fmt.Errorf("go ast (new) %s: %w", path, err)
	}

	return buildDeltas(path, oldScores, newScores), nil
}

func (a *Analyzer) scoreContent(path, content string) (map[string]int, error) {
	if strings.TrimSpace(content) == "" {
		return map[string]int{}, nil
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	scores := make(map[string]int)
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		scores[qualifiedName(fn)] = complexity(fn.Body)
	}
	return scores, nil
}

func qualifiedName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	return receiverName(fn.Recv.List[0].Type) + "." + fn.Name.Name
}

func receiverName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverName(t.X)
	case *ast.IndexExpr:
		return receiverName(t.X)
	case *ast.IndexListExpr:
		return receiverName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return "receiver"
	}
}

// complexity returns cyclomatic complexity for a Go function body. It starts at
// 1 for the straight-line path and increments for common Go decision points:
// if/for/range, switch/select cases, and boolean &&/|| operators.
func complexity(body *ast.BlockStmt) int {
	score := 1

	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case nil:
			return true
		case *ast.FuncLit:
			// Do not fold nested anonymous functions into the enclosing function's
			// score. They are independent executable units.
			return false
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			score++
		case *ast.CaseClause:
			if x.List != nil { // default does not add a branch.
				score++
			}
		case *ast.CommClause:
			if x.Comm != nil { // default does not add a branch.
				score++
			}
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				score++
			}
		}
		return true
	})

	return score
}

func buildDeltas(filePath string, old, new map[string]int) []interfaces.FunctionDelta {
	seen := make(map[string]struct{})
	for k := range old {
		seen[k] = struct{}{}
	}
	for k := range new {
		seen[k] = struct{}{}
	}

	deltas := make([]interfaces.FunctionDelta, 0, len(seen))
	for name := range seen {
		oldC := old[name]
		newC := new[name]
		deltas = append(deltas, interfaces.FunctionDelta{
			FilePath:      filePath,
			FunctionName:  name,
			OldComplexity: oldC,
			NewComplexity: newC,
			Delta:         newC - oldC,
		})
	}
	return deltas
}
