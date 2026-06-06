package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

var testDeltas = []interfaces.FunctionDelta{
	{FilePath: "auth.py", FunctionName: "login", OldComplexity: 2, NewComplexity: 7, Delta: 5},
	{FilePath: "auth.py", FunctionName: "logout", OldComplexity: 1, NewComplexity: 1, Delta: 0},
	{FilePath: "utils.py", FunctionName: "parse", OldComplexity: 3, NewComplexity: 8, Delta: 5},
	{FilePath: "models.py", FunctionName: "validate", OldComplexity: 4, NewComplexity: 2, Delta: -2},
	{FilePath: "new.py", FunctionName: "fresh", OldComplexity: 0, NewComplexity: 3, Delta: 3},
}

func TestGenerate_TextOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Generate(&buf, testDeltas, Options{Format: "text"}); err != nil {
		t.Fatalf("Generate (text): %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "login") {
		t.Error("expected 'login' in text output")
	}
}

func TestGenerate_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Generate(&buf, testDeltas, Options{Format: "json"}); err != nil {
		t.Fatalf("Generate (json): %v", err)
	}
	var report jsonReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}
	if len(report.Results) == 0 {
		t.Error("expected non-empty Results in JSON output")
	}
}

func TestGenerate_MarkdownOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Generate(&buf, testDeltas, Options{Format: "markdown"}); err != nil {
		t.Fatalf("Generate (markdown): %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "|") {
		t.Error("expected markdown table pipes in output")
	}
}

func TestGenerate_DefaultIsText(t *testing.T) {
	var buf bytes.Buffer
	// Empty Format string should default to text.
	if err := Generate(&buf, testDeltas, Options{}); err != nil {
		t.Fatalf("Generate (default): %v", err)
	}
	// Text output contains the table header word "File" — JSON does not.
	if !strings.Contains(buf.String(), "File") {
		t.Error("default format should be text")
	}
}

func TestFilter_ExcludesUnchanged(t *testing.T) {
	result := filter(testDeltas, Options{IncludeUnchanged: false})
	for _, d := range result {
		if d.Delta == 0 {
			t.Errorf("unchanged function %s should be filtered out", d.FunctionName)
		}
	}
}

func TestFilter_MinDelta(t *testing.T) {
	result := filter(testDeltas, Options{MinDelta: 4})
	for _, d := range result {
		if d.Delta < 4 {
			t.Errorf("function %s has delta %d, below MinDelta 4", d.FunctionName, d.Delta)
		}
	}
}

func TestSortDeltas(t *testing.T) {
	deltas := []interfaces.FunctionDelta{
		{FunctionName: "a", Delta: 1},
		{FunctionName: "b", Delta: 5},
		{FunctionName: "c", Delta: 3},
	}
	sortDeltas(deltas)
	if deltas[0].Delta != 5 || deltas[1].Delta != 3 || deltas[2].Delta != 1 {
		t.Errorf("wrong sort order: %v", deltas)
	}
}
