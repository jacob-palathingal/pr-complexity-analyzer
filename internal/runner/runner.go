package runner

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/analyzers/goast"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/analyzers/python"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/diff"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/pathfilter"
	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/report"
)

// Exit code constants.
const (
	ExitOK        = 0 // No threshold breaches, or no threshold configured.
	ExitThreshold = 1 // At least one function met or exceeded --threshold.
	ExitToolError = 2 // Tool error.
)

const defaultTimeout = 30 * time.Second

// Config holds all run parameters sourced from CLI flags.
type Config struct {
	RepoDir string
	BaseRef string
	HeadRef string

	MinDelta         int
	IncludeUnchanged bool

	// Format selects the output formatter: "text" (default), "json", "markdown".
	Format string

	// LangFilter restricts analysis to a language: python, py, go, golang, all, or empty.
	LangFilter string

	// Threshold is the delta that triggers ExitThreshold. 0 = disabled.
	Threshold int

	// Timeout is the per-git-command timeout.
	Timeout time.Duration

	// Include and Exclude are repo-relative glob filters.
	Include []string
	Exclude []string

	// IncludeTests includes files like *_test.go, *_test.py, test/, and tests/.
	IncludeTests bool
}

// Result carries the recommended exit code and breach count.
type Result struct {
	ExitCode          int
	BreachedFunctions int
}

var registry = []interfaces.Analyzer{
	python.New(),
	goast.New(),
}

// Run executes the full pipeline and returns a Result with the exit code.
func Run(w io.Writer, cfg Config) (Result, error) {
	cfg = cfg.withDefaults()

	if err := cfg.Validate(); err != nil {
		return Result{ExitCode: ExitToolError}, err
	}

	gitClient := diff.NewClientWithTimeout(cfg.RepoDir, cfg.Timeout)
	parser := diff.NewParser(gitClient)

	fileDiffs, err := parser.BuildDiffsFiltered(cfg.BaseRef, cfg.HeadRef, func(path string) bool {
		if !pathfilter.ShouldAnalyze(path, pathfilter.Options{
			Include:      cfg.Include,
			Exclude:      cfg.Exclude,
			IncludeTests: cfg.IncludeTests,
		}) {
			return false
		}

		return findAnalyzer(path, cfg.LangFilter) != nil
	})
	if err != nil {
		return Result{ExitCode: ExitToolError}, fmt.Errorf("building diffs: %w", err)
	}

	var allDeltas []interfaces.FunctionDelta
	for _, fd := range fileDiffs {
		analyzer := findAnalyzer(fd.Path, cfg.LangFilter)
		if analyzer == nil {
			continue
		}

		deltas, err := analyzer.Analyze(fd.Path, fd.OldContent, fd.NewContent)
		if err != nil {
			return Result{ExitCode: ExitToolError}, fmt.Errorf("analyzing %s: %w", fd.Path, err)
		}

		allDeltas = append(allDeltas, deltas...)
	}

	result := thresholdResult(cfg, allDeltas)

	if err := report.Generate(w, allDeltas, report.Options{
		MinDelta:          cfg.MinDelta,
		IncludeUnchanged:  cfg.IncludeUnchanged,
		Format:            cfg.Format,
		GeneratedAt:       time.Now().UTC(),
		Threshold:         cfg.Threshold,
		BreachedFunctions: result.BreachedFunctions,
		ExitCode:          result.ExitCode,
		LangFilter:        normalizeLangFilter(cfg.LangFilter),
		Include:           cfg.Include,
		Exclude:           cfg.Exclude,
	}); err != nil {
		return Result{ExitCode: ExitToolError}, err
	}

	if result.ExitCode == ExitThreshold && normalizeFormat(cfg.Format) != "json" {
		fmt.Fprintf(w, "\n✗ %d function(s) exceeded the complexity threshold of +%d\n", result.BreachedFunctions, cfg.Threshold)
	}

	return result, nil
}

func (cfg Config) withDefaults() Config {
	cfg.Format = normalizeFormat(cfg.Format)

	if strings.TrimSpace(cfg.HeadRef) == "" {
		cfg.HeadRef = "HEAD"
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}

	return cfg
}

func (cfg Config) Validate() error {
	if strings.TrimSpace(cfg.BaseRef) == "" {
		return fmt.Errorf("base ref is required")
	}

	if strings.TrimSpace(cfg.HeadRef) == "" {
		return fmt.Errorf("head ref is required")
	}

	if cfg.Threshold < 0 {
		return fmt.Errorf("threshold must be >= 0")
	}

	if cfg.MinDelta < 0 {
		return fmt.Errorf("min-delta must be >= 0")
	}

	if cfg.Timeout < 0 {
		return fmt.Errorf("timeout must be >= 0")
	}

	switch normalizeFormat(cfg.Format) {
	case "text", "json", "markdown":
	default:
		return fmt.Errorf("unsupported format %q (supported: text, json, markdown)", cfg.Format)
	}

	switch normalizeLangFilter(cfg.LangFilter) {
	case "", "python", "go":
	default:
		return fmt.Errorf("unsupported language %q (supported: all, python, py, go, golang)", cfg.LangFilter)
	}

	return nil
}

func thresholdResult(cfg Config, deltas []interfaces.FunctionDelta) Result {
	if cfg.Threshold <= 0 {
		return Result{ExitCode: ExitOK}
	}

	var breached int
	for _, delta := range deltas {
		if delta.Delta >= cfg.Threshold {
			breached++
		}
	}

	if breached == 0 {
		return Result{ExitCode: ExitOK}
	}

	return Result{
		ExitCode:          ExitThreshold,
		BreachedFunctions: breached,
	}
}

// applyThreshold is kept for tests and for callers that want the old behavior:
// compute threshold status and write the human-readable breach message.
func applyThreshold(w io.Writer, cfg Config, deltas []interfaces.FunctionDelta) Result {
	result := thresholdResult(cfg, deltas)

	if result.ExitCode == ExitThreshold && normalizeFormat(cfg.Format) != "json" {
		fmt.Fprintf(w, "\n✗ %d function(s) exceeded the complexity threshold of +%d\n", result.BreachedFunctions, cfg.Threshold)
	}

	return result
}

func findAnalyzer(path, langFilter string) interfaces.Analyzer {
	langFilter = normalizeLangFilter(langFilter)

	for _, analyzer := range registry {
		if langFilter != "" && !matchesAnalyzer(analyzer, langFilter) {
			continue
		}

		if analyzer.Supports(path) {
			return analyzer
		}
	}

	return nil
}

func normalizeFormat(format string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		return "text"
	}

	return format
}

func normalizeLangFilter(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))

	switch lang {
	case "", "all":
		return ""
	case "py":
		return "python"
	case "golang":
		return "go"
	default:
		return lang
	}
}

func matchesAnalyzer(analyzer interfaces.Analyzer, langFilter string) bool {
	if langFilter == "" {
		return true
	}

	name := strings.ToLower(analyzer.Name())

	if strings.Contains(name, langFilter) {
		return true
	}

	if langFilter == "go" && strings.HasPrefix(name, "go/") {
		return true
	}

	if langFilter == "python" && strings.HasPrefix(name, "python/") {
		return true
	}

	return false
}
