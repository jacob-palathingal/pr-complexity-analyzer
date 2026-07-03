package report

import (
	"io"
	"sort"
	"time"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Options controls which deltas are included and how they are rendered.
type Options struct {
	MinDelta         int
	IncludeUnchanged bool
	Format           string

	GeneratedAt       time.Time
	Threshold         int
	BreachedFunctions int
	ExitCode          int

	LangFilter string
	Include    []string
	Exclude    []string
}

// Payload is the stable report contract used by all formatters.
type Payload struct {
	SchemaVersion string                     `json:"schema_version"`
	GeneratedAt   time.Time                  `json:"generated_at"`
	Summary       Summary                    `json:"summary"`
	Filters       Filters                    `json:"filters"`
	Results       []interfaces.FunctionDelta `json:"results"`
}

// Summary captures run-level metadata for CI and dashboards.
type Summary struct {
	AnalyzedFunctions int `json:"analyzed_functions"`
	ReportedFunctions int `json:"reported_functions"`
	Threshold         int `json:"threshold"`
	BreachedFunctions int `json:"breached_functions"`
	MaxDelta          int `json:"max_delta"`
	ExitCode          int `json:"exit_code"`
}

// Filters records the effective filters used for this run.
type Filters struct {
	Language         string   `json:"language,omitempty"`
	MinDelta         int      `json:"min_delta"`
	IncludeUnchanged bool     `json:"include_unchanged"`
	Include          []string `json:"include,omitempty"`
	Exclude          []string `json:"exclude,omitempty"`
}

// Generate filters, sorts, builds a payload, and writes the report to w.
func Generate(w io.Writer, deltas []interfaces.FunctionDelta, opts Options) error {
	opts = opts.withDefaults()

	filtered := filter(deltas, opts)
	sortDeltas(filtered)

	payload := Payload{
		SchemaVersion: "1.0",
		GeneratedAt:   opts.GeneratedAt,
		Summary:       buildSummary(deltas, filtered, opts),
		Filters: Filters{
			Language:         opts.LangFilter,
			MinDelta:         opts.MinDelta,
			IncludeUnchanged: opts.IncludeUnchanged,
			Include:          opts.Include,
			Exclude:          opts.Exclude,
		},
		Results: filtered,
	}

	var formatter Formatter
	switch opts.Format {
	case "json":
		formatter = &JSONFormatter{}
	case "markdown":
		formatter = &MarkdownFormatter{}
	default:
		formatter = &TextFormatter{}
	}

	return formatter.Format(w, payload)
}

func (opts Options) withDefaults() Options {
	if opts.Format == "" {
		opts.Format = "text"
	}

	if opts.GeneratedAt.IsZero() {
		opts.GeneratedAt = time.Now().UTC()
	}

	return opts
}

func buildSummary(allDeltas, reported []interfaces.FunctionDelta, opts Options) Summary {
	summary := Summary{
		AnalyzedFunctions: len(allDeltas),
		ReportedFunctions: len(reported),
		Threshold:         opts.Threshold,
		BreachedFunctions: opts.BreachedFunctions,
		ExitCode:          opts.ExitCode,
	}

	for _, delta := range allDeltas {
		if delta.Delta > summary.MaxDelta {
			summary.MaxDelta = delta.Delta
		}
	}

	return summary
}

func filter(deltas []interfaces.FunctionDelta, opts Options) []interfaces.FunctionDelta {
	out := make([]interfaces.FunctionDelta, 0, len(deltas))

	for _, delta := range deltas {
		if !opts.IncludeUnchanged && delta.Delta == 0 {
			continue
		}

		if delta.Delta < opts.MinDelta {
			continue
		}

		out = append(out, delta)
	}

	return out
}

func sortDeltas(deltas []interfaces.FunctionDelta) {
	sort.Slice(deltas, func(i, j int) bool {
		if deltas[i].Delta != deltas[j].Delta {
			return deltas[i].Delta > deltas[j].Delta
		}

		if deltas[i].FilePath != deltas[j].FilePath {
			return deltas[i].FilePath < deltas[j].FilePath
		}

		return deltas[i].FunctionName < deltas[j].FunctionName
	})
}
