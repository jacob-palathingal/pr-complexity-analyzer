package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
	"github.com/olekukonko/tablewriter"
)

// Formatter is the output interface. TextFormatter and JSONFormatter implement it.
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

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"File", "Function", "Before", "After", "Delta"})
	table.SetBorder(false)
	table.SetColumnSeparator("  ")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(true)
	table.SetAutoWrapText(false)

	// Alignment overrides: Before/After/Delta are right-aligned (numeric).
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,  // File
		tablewriter.ALIGN_LEFT,  // Function
		tablewriter.ALIGN_RIGHT, // Before
		tablewriter.ALIGN_RIGHT, // After
		tablewriter.ALIGN_RIGHT, // Delta
	})

	for _, d := range deltas {
		deltaStr := formatDelta(d.Delta)
		table.Append([]string{
			d.FilePath,
			d.FunctionName,
			formatScore(d.OldComplexity),
			formatScore(d.NewComplexity),
			deltaStr,
		})
	}

	table.Render()
	fmt.Fprintln(w)
	return nil
}

func formatDelta(delta int) string {
	switch {
	case delta > 0:
		return fmt.Sprintf("+%d ▲", delta)
	case delta < 0:
		return fmt.Sprintf("%d ▼", delta)
	default:
		return "  0"
	}
}

func formatScore(score int) string {
	if score == 0 {
		return strings.Repeat(" ", 3) + "—"
	}
	return fmt.Sprintf("%d", score)
}
