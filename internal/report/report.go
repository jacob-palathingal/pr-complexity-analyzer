package report

import (
	"io"
	"sort"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Options controls which deltas are included and how they are rendered.
type Options struct {
	// MinDelta filters out functions whose delta is below this value.
	MinDelta int

	// IncludeUnchanged includes zero-delta functions when true.
	IncludeUnchanged bool

	// Format selects the output formatter: "text" (default), "json", "markdown".
	Format string
}

// Generate filters, sorts, and writes the report to w.
func Generate(w io.Writer, deltas []interfaces.FunctionDelta, opts Options) error {
	filtered := filter(deltas, opts)
	sortDeltas(filtered)

	var f Formatter
	switch opts.Format {
	case "json":
		f = &JSONFormatter{}
	case "markdown":
		f = &MarkdownFormatter{}
	default:
		f = &TextFormatter{}
	}
	return f.Format(w, filtered)
}

func filter(deltas []interfaces.FunctionDelta, opts Options) []interfaces.FunctionDelta {
	out := make([]interfaces.FunctionDelta, 0, len(deltas))
	for _, d := range deltas {
		if !opts.IncludeUnchanged && d.Delta == 0 {
			continue
		}
		if d.Delta < opts.MinDelta {
			continue
		}
		out = append(out, d)
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
