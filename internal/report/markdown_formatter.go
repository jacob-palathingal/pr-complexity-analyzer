package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// MarkdownFormatter renders a GitHub-flavored Markdown table suitable for
// posting as a PR comment.
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Format(w io.Writer, deltas []interfaces.FunctionDelta) error {
	if len(deltas) == 0 {
		_, err := fmt.Fprintln(w, "✅ **No complexity increases found in this PR.**")
		return err
	}

	fmt.Fprintf(w, "## PR Complexity Report\n\n")
	fmt.Fprintf(w, "**%d function(s) changed complexity**\n\n", len(deltas))
	fmt.Fprintln(w, "| File | Function | Before | After | Delta |")
	fmt.Fprintln(w, "|------|----------|-------:|------:|------:|")

	for _, d := range deltas {
		fmt.Fprintf(w, "| `%s` | `%s` | %s | %s | %s |\n",
			d.FilePath,
			d.FunctionName,
			markdownScore(d.OldComplexity),
			markdownScore(d.NewComplexity),
			markdownDelta(d.Delta),
		)
	}

	fmt.Fprintln(w)

	// Summary line for quick scanning.
	increased := 0
	for _, d := range deltas {
		if d.Delta > 0 {
			increased++
		}
	}
	if increased > 0 {
		fmt.Fprintf(w, "> ⚠️ %d function(s) increased in complexity. ", increased)
		fmt.Fprintf(w, "Consider breaking complex functions into smaller pieces.\n")
	}

	return nil
}

func markdownDelta(delta int) string {
	switch {
	case delta > 0:
		return fmt.Sprintf("**+%d** ⬆️", delta)
	case delta < 0:
		return fmt.Sprintf("%d ⬇️", delta)
	default:
		return "0"
	}
}

func markdownScore(score int) string {
	if score == 0 {
		return strings.Repeat("—", 1)
	}
	return fmt.Sprintf("%d", score)
}
