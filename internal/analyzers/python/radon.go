package python

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/analyzers/common"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Analyzer implements interfaces.Analyzer for Python files using Radon.
type Analyzer struct{}

// New returns a ready-to-use Python/Radon Analyzer.
func New() *Analyzer {
	return &Analyzer{}
}

// Name identifies this analyzer.
func (a *Analyzer) Name() string {
	return "python/radon"
}

// Supports returns true for .py files.
func (a *Analyzer) Supports(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".py"
}

// Analyze computes cyclomatic complexity for each function in oldContent and
// newContent, then returns the delta for every function seen in either version.
func (a *Analyzer) Analyze(path, oldContent, newContent string) ([]interfaces.FunctionDelta, error) {
	oldScores, err := a.scoreContent(oldContent)
	if err != nil {
		return nil, fmt.Errorf("radon old snapshot %s: %w", path, err)
	}

	newScores, err := a.scoreContent(newContent)
	if err != nil {
		return nil, fmt.Errorf("radon new snapshot %s: %w", path, err)
	}

	return buildDeltas(path, oldScores, newScores), nil
}

// scoreContent writes content to a temp file, runs radon cc on it, and returns
// a map of fully-qualified function name -> complexity score.
func (a *Analyzer) scoreContent(content string) (map[string]int, error) {
	if strings.TrimSpace(content) == "" {
		return map[string]int{}, nil
	}

	if _, err := exec.LookPath("radon"); err != nil {
		return nil, fmt.Errorf("radon executable not found: install with `pip install radon`: %w", err)
	}

	tmp, err := os.CreateTemp("", "pr-complexity-*.py")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}

	return a.runRadon(tmp.Name())
}

// radonBlock is the JSON structure radon cc --json emits per file.
// Radon can emit classes, methods, and functions. We intentionally exclude
// class-level blocks because this tool reports per-function complexity deltas.
type radonBlock struct {
	Name       string `json:"name"`
	Complexity int    `json:"complexity"`
	Type       string `json:"type"`
	ClassName  string `json:"classname"`
}

// runRadon executes `radon cc --json --show-complexity` and parses the output.
func (a *Analyzer) runRadon(filePath string) (map[string]int, error) {
	cmd := exec.Command("radon", "cc", "--json", "--show-complexity", filePath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("radon failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	var raw map[string][]radonBlock
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse radon output: %w: raw=%s", err, string(out))
	}

	scores := make(map[string]int)
	for _, blocks := range raw {
		for _, block := range blocks {
			if block.Type != "function" && block.Type != "method" {
				continue
			}

			scores[qualifiedName(block)] = block.Complexity
		}
	}

	return scores, nil
}

// qualifiedName returns "ClassName.method" for methods, bare "name" otherwise.
func qualifiedName(block radonBlock) string {
	if block.ClassName != "" {
		return block.ClassName + "." + block.Name
	}

	return block.Name
}

// buildDeltas is kept package-local for existing tests while delegating to the
// shared deterministic implementation.
func buildDeltas(filePath string, oldScores, newScores map[string]int) []interfaces.FunctionDelta {
	return common.BuildDeltas(filePath, "python", "python/radon", oldScores, newScores)
}
