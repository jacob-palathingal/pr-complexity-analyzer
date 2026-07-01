package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/analyzers/common"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Analyzer implements interfaces.Analyzer for Go source files.
type Analyzer struct{}

// New returns a ready-to-use Go AST Analyzer.
func New() *Analyzer {
	return &Analyzer{}
}

// Name identifies this analyzer.
func (a *Analyzer) Name() string {
	return "go/ast"
}

// Supports returns true for .go source files.
func (a *Analyzer) Supports(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".go"
}

// Analyze computes cyclomatic complexity for each function in oldContent and
// newContent, then returns the delta for every function seen in either version.
func (a *Analyzer) Analyze(path, oldContent, newContent string) ([]interfaces.FunctionDelta, error) {
	oldScores, err := a.scoreContent(path, oldContent)
	if err != nil {
		return nil, fmt.Errorf("go ast old snapshot %s: %w", path, err)
	}

	newScores, err := a.scoreContent(path, newContent)
	if err != nil {
		return nil, fmt.Errorf("go ast new snapshot %s: %w", path, err)
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

// complexity returns cyclomatic complexity for a Go function body.
// It starts at 1 and increments for common Go decision points.
func complexity(body *ast.BlockStmt) int {
	score := 1

	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case nil:
			return true
		case *ast.FuncLit:
			// Do not fold nested anonymous functions into the enclosing
			// function's score. They are independent executable units.
			return false
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			score++
		case *ast.CaseClause:
			if x.List != nil {
				score++
			}
		case *ast.CommClause:
			if x.Comm != nil {
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

// buildDeltas is kept package-local for existing tests while delegating to the
// shared deterministic implementation.
func buildDeltas(filePath string, oldScores, newScores map[string]int) []interfaces.FunctionDelta {
	return common.BuildDeltas(filePath, "go", "go/ast", oldScores, newScores)
}
