package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// Formatter is the output interface. TextFormatter, JSONFormatter, and
// MarkdownFormatter implement it.
type Formatter interface {
	Format(w io.Writer, deltas []interfaces.FunctionDelta) error
}

// TextFormatter renders a human-readable terminal table.
type TextFormatter struct{}

func (f *TextFormatter) Format(w io.Writer, deltas []interfaces.FunctionDelta) error {
	if len(deltas) == 0 {
		_, err := fmt.Fprintln(w, "No complexity increases found in this PR.")
		return err
	}

	fmt.Fprintf(w, "\n  PR Complexity Report — %d function(s) changed\n\n", len(deltas))
	writeTextTable(w, deltas)
	fmt.Fprintln(w)
	return nil
}

func writeTextTable(w io.Writer, deltas []interfaces.FunctionDelta) {
	headers := []string{"File", "Function", "Before", "After", "Delta"}
	rows := make([][]string, 0, len(deltas))
	widths := []int{len(headers[0]), len(headers[1]), len(headers[2]), len(headers[3]), len(headers[4])}

	for _, d := range deltas {
		row := []string{
			d.FilePath,
			d.FunctionName,
			formatScore(d.OldComplexity),
			formatScore(d.NewComplexity),
			formatDelta(d.Delta),
		}
		rows = append(rows, row)
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	fmt.Fprintf(w, "  %-*s  %-*s  %*s  %*s  %*s\n",
		widths[0], headers[0],
		widths[1], headers[1],
		widths[2], headers[2],
		widths[3], headers[3],
		widths[4], headers[4],
	)
	fmt.Fprintf(w, "  %s  %s  %s  %s  %s\n",
		strings.Repeat("-", widths[0]),
		strings.Repeat("-", widths[1]),
		strings.Repeat("-", widths[2]),
		strings.Repeat("-", widths[3]),
		strings.Repeat("-", widths[4]),
	)

	for _, row := range rows {
		fmt.Fprintf(w, "  %-*s  %-*s  %*s  %*s  %*s\n",
			widths[0], row[0],
			widths[1], row[1],
			widths[2], row[2],
			widths[3], row[3],
			widths[4], row[4],
		)
	}
}

func formatDelta(delta int) string {
	switch {
	case delta > 0:
		return fmt.Sprintf("+%d ▲", delta)
	case delta < 0:
		return fmt.Sprintf("%d ▼", delta)
	default:
		return "0"
	}
}

func formatScore(score int) string {
	if score == 0 {
		return "—"
	}
	return fmt.Sprintf("%d", score)
}
