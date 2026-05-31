// Package runner ties together the diff engine, analyzer registry, and report
// generator into a single Run call. It is the only package that knows about
// all three layers; cmd/ calls runner.Run and handles CLI flag parsing.
package runner

import (
	"fmt"
	"io"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/diff"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces/analyzers/python"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/report"
)

// Config holds everything Run needs, sourced directly from CLI flags.
type Config struct {
	// RepoDir is the path to the git repository. Empty string = cwd.
	RepoDir string

	// BaseRef and HeadRef are the git refs to compare.
	BaseRef string
	HeadRef string

	// MinDelta filters out functions with a complexity increase below this value.
	MinDelta int

	// IncludeUnchanged adds zero-delta functions to the report.
	IncludeUnchanged bool

	// JSON switches to JSON output.
	JSON bool

	// LangFilter restricts analysis to analyzers whose Name() contains this string.
	// Empty string means all languages.
	LangFilter string
}

// registry is the list of all known language analyzers, in dispatch order.
// To add a new language: implement interfaces.Analyzer and append here.
var registry = []interfaces.Analyzer{
	python.New(),
}

// Run executes the full analysis pipeline and writes the report to w.
func Run(w io.Writer, cfg Config) error {
	// 1. Build diffs.
	gitClient := diff.NewClient(cfg.RepoDir)
	parser := diff.NewParser(gitClient)

	fileDiffs, err := parser.BuildDiffs(cfg.BaseRef, cfg.HeadRef)
	if err != nil {
		return fmt.Errorf("building diffs: %w", err)
	}

	if len(fileDiffs) == 0 {
		fmt.Fprintln(w, "No files changed between the given refs.")
		return nil
	}

	// 2. Dispatch each file to the appropriate analyzer.
	var allDeltas []interfaces.FunctionDelta
	var unsupported []string

	for _, fd := range fileDiffs {
		analyzer := findAnalyzer(fd.Path, cfg.LangFilter)
		if analyzer == nil {
			unsupported = append(unsupported, fd.Path)
			continue
		}

		deltas, err := analyzer.Analyze(fd.Path, fd.OldContent, fd.NewContent)
		if err != nil {
			return fmt.Errorf("analyzing %s: %w", fd.Path, err)
		}
		allDeltas = append(allDeltas, deltas...)
	}

	// Warn about unsupported files (don't fail — PRs often touch configs etc.).
	if len(unsupported) > 0 {
		fmt.Fprintf(w, "⚠  Skipped %d file(s) with no supported analyzer: %s\n\n",
			len(unsupported), strings.Join(unsupported, ", "))
	}

	// 3. Generate report.
	return report.Generate(w, allDeltas, report.Options{
		MinDelta:         cfg.MinDelta,
		IncludeUnchanged: cfg.IncludeUnchanged,
		JSON:             cfg.JSON,
	})
}

// findAnalyzer returns the first registered analyzer that supports path and
// matches the optional langFilter. Returns nil if none match.
func findAnalyzer(path, langFilter string) interfaces.Analyzer {
	for _, a := range registry {
		if langFilter != "" && !strings.Contains(a.Name(), langFilter) {
			continue
		}
		if a.Supports(path) {
			return a
		}
	}
	return nil
}
