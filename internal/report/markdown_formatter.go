package report

import (
	"fmt"
	"io"
	"strings"
)

// MarkdownFormatter renders a GitHub-flavored Markdown table suitable for PR comments.
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Format(w io.Writer, payload Payload) error {
	deltas := payload.Results

	if len(deltas) == 0 {
		_, err := fmt.Fprintln(w, "✅ **No complexity increases found in this PR.**")
		return err
	}

	fmt.Fprintln(w, "## PR Complexity Report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "**%d function(s) reported** from **%d analyzed function(s)**.\n\n", len(deltas), payload.Summary.AnalyzedFunctions)

	if payload.Summary.Threshold > 0 {
		fmt.Fprintf(w, "- Threshold: `+%d`\n", payload.Summary.Threshold)
		fmt.Fprintf(w, "- Breaches: `%d`\n", payload.Summary.BreachedFunctions)
		fmt.Fprintf(w, "- Max delta: `+%d`\n\n", payload.Summary.MaxDelta)
	}

	fmt.Fprintln(w, "| File | Function | Before | After | Delta |")
	fmt.Fprintln(w, "|------|----------|-------:|------:|------:|")

	for _, delta := range deltas {
		fmt.Fprintf(
			w,
			"| `%s` | `%s` | %s | %s | %s |\n",
			escapeMarkdownCode(delta.FilePath),
			escapeMarkdownCode(delta.FunctionName),
			markdownScore(delta.OldComplexity),
			markdownScore(delta.NewComplexity),
			markdownDelta(delta.Delta),
		)
	}

	fmt.Fprintln(w)

	if payload.Summary.BreachedFunctions > 0 {
		fmt.Fprintf(w, "> ❌ %d function(s) exceeded the configured complexity threshold.\n", payload.Summary.BreachedFunctions)
		return nil
	}

	increased := 0
	for _, delta := range deltas {
		if delta.Delta > 0 {
			increased++
		}
	}

	if increased > 0 {
		fmt.Fprintf(w, "> ⚠️ %d function(s) increased in complexity. Consider breaking complex functions into smaller pieces.\n", increased)
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
		return "—"
	}

	return fmt.Sprintf("%d", score)
}

func escapeMarkdownCode(value string) string {
	value = strings.ReplaceAll(value, "`", "\\`")
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")

	return value
}
