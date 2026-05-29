// Package report sorts FunctionDelta results and dispatches to a Formatter.
package report

import (
	"io"
	"sort"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Options controls which deltas are included and how they're rendered.
type Options struct {
	// MinDelta filters out functions whose complexity change is below this value.
	// 0 (default) includes all changed functions.
	MinDelta int

	// IncludeUnchanged includes functions with Delta == 0 when true.
	IncludeUnchanged bool

	// JSON switches to machine-readable output.
	JSON bool
}

// Generate filters deltas by opts, sorts them by descending Delta (then by
// file+function name for stability), and writes a report to w.
func Generate(w io.Writer, deltas []interfaces.FunctionDelta, opts Options) error {
	filtered := filter(deltas, opts)
	sortDeltas(filtered)

	var f Formatter
	if opts.JSON {
		f = &JSONFormatter{}
	} else {
		f = &TextFormatter{}
	}
	return f.Format(w, filtered)
}

// filter returns only the deltas that satisfy opts.
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

// sortDeltas sorts by Delta descending, then FilePath+FunctionName ascending.
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
