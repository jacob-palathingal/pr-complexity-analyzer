package python

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Analyzer implements interfaces.Analyzer for Python files using Radon.
type Analyzer struct{}

// New returns a ready-to-use Python/Radon Analyzer.
func New() *Analyzer { return &Analyzer{} }

// Name identifies this analyzer.
func (a *Analyzer) Name() string { return "python/radon" }

// Supports returns true for .py files.
func (a *Analyzer) Supports(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".py"
}

// Analyze computes cyclomatic complexity for each function in oldContent and
// newContent, then returns the delta for every function seen in either version.
func (a *Analyzer) Analyze(path, oldContent, newContent string) ([]interfaces.FunctionDelta, error) {
	oldScores, err := a.scoreContent(oldContent)
	if err != nil {
		return nil, fmt.Errorf("radon (old) %s: %w", path, err)
	}

	newScores, err := a.scoreContent(newContent)
	if err != nil {
		return nil, fmt.Errorf("radon (new) %s: %w", path, err)
	}

	return buildDeltas(path, oldScores, newScores), nil
}

// scoreContent writes content to a temp file, runs radon cc on it, and returns
// a map of function name → complexity score.
func (a *Analyzer) scoreContent(content string) (map[string]int, error) {
	if strings.TrimSpace(content) == "" {
		return map[string]int{}, nil
	}

	tmp, err := os.CreateTemp("", "pr-complexity-*.py")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(content); err != nil {
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}

	return a.runRadon(tmp.Name())
}

// radonBlock is the JSON structure radon cc --json emits per file.
// Each element in the array is one function/method/class.
type radonBlock struct {
	Name       string `json:"name"`
	Complexity int    `json:"complexity"`
	Type       string `json:"type"` // "function", "method", "class"
	ClassName  string `json:"classname"`
}

// runRadon executes `radon cc --json --show-complexity <file>` and parses
// the output into a name→complexity map.
func (a *Analyzer) runRadon(filePath string) (map[string]int, error) {
	cmd := exec.Command("radon", "cc", "--json", "--show-complexity", filePath)
	out, err := cmd.Output()
	if err != nil {
		// radon exits non-zero on syntax errors. Wrap with output for context.
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("radon exited %d: %s", ee.ExitCode(), string(ee.Stderr))
		}
		return nil, fmt.Errorf("radon: %w", err)
	}

	// radon --json returns {"<filepath>": [block, block, ...]}
	var raw map[string][]radonBlock
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse radon output: %w — raw: %s", err, string(out))
	}

	scores := make(map[string]int)
	for _, blocks := range raw {
		for _, b := range blocks {
			key := qualifiedName(b)
			scores[key] = b.Complexity
		}
	}
	return scores, nil
}

// qualifiedName returns "ClassName.method" for methods, bare "name" otherwise.
func qualifiedName(b radonBlock) string {
	if b.ClassName != "" {
		return b.ClassName + "." + b.Name
	}
	return b.Name
}

// buildDeltas merges old and new score maps into a slice of FunctionDelta,
// one entry per unique function name seen in either map.
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
