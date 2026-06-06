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

// Exit code constants.
const (
	ExitOK        = 0 // No threshold breaches, or no threshold configured.
	ExitThreshold = 1 // At least one function met or exceeded --threshold.
	ExitError     = 2 // Tool error (bad ref, missing git, analyzer failure).
)

// Config holds all run parameters sourced from CLI flags and config file.
type Config struct {
	RepoDir          string
	BaseRef          string
	HeadRef          string
	MinDelta         int
	IncludeUnchanged bool
	// Format selects the output formatter: "text" (default), "json", "markdown".
	Format     string
	LangFilter string
	// Threshold is the delta that triggers ExitThreshold. 0 = disabled.
	Threshold int
}

// Result carries the recommended exit code and breach count.
type Result struct {
	ExitCode          int
	BreachedFunctions int
}

var registry = []interfaces.Analyzer{
	python.New(),
}

// Run executes the full pipeline and returns a Result with the exit code.
func Run(w io.Writer, cfg Config) (Result, error) {
	gitClient := diff.NewClient(cfg.RepoDir)
	parser := diff.NewParser(gitClient)

	fileDiffs, err := parser.BuildDiffs(cfg.BaseRef, cfg.HeadRef)
	if err != nil {
		return Result{ExitCode: ExitError}, fmt.Errorf("building diffs: %w", err)
	}

	if len(fileDiffs) == 0 {
		fmt.Fprintln(w, "No files changed between the given refs.")
		return Result{ExitCode: ExitOK}, nil
	}

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
			return Result{ExitCode: ExitError}, fmt.Errorf("analyzing %s: %w", fd.Path, err)
		}
		allDeltas = append(allDeltas, deltas...)
	}

	if len(unsupported) > 0 {
		fmt.Fprintf(w, "⚠  Skipped %d file(s) with no supported analyzer: %s\n\n",
			len(unsupported), strings.Join(unsupported, ", "))
	}

	if err := report.Generate(w, allDeltas, report.Options{
		MinDelta:         cfg.MinDelta,
		IncludeUnchanged: cfg.IncludeUnchanged,
		Format:           cfg.Format,
	}); err != nil {
		return Result{ExitCode: ExitError}, err
	}

	return applyThreshold(w, cfg, allDeltas), nil
}

// applyThreshold counts breaches and prints a summary line if any are found.
func applyThreshold(w io.Writer, cfg Config, deltas []interfaces.FunctionDelta) Result {
	if cfg.Threshold <= 0 {
		return Result{ExitCode: ExitOK}
	}
	var breached int
	for _, d := range deltas {
		if d.Delta >= cfg.Threshold {
			breached++
		}
	}
	if breached > 0 {
		fmt.Fprintf(w, "\n✗  %d function(s) exceeded the complexity threshold of +%d\n",
			breached, cfg.Threshold)
		return Result{ExitCode: ExitThreshold, BreachedFunctions: breached}
	}
	return Result{ExitCode: ExitOK}
}

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
