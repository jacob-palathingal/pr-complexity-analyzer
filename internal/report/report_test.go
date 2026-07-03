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

	if err := Generate(&buf, testDeltas, Options{
		Format:            "json",
		Threshold:         5,
		BreachedFunctions: 2,
		ExitCode:          1,
		LangFilter:        "python",
		Include:           []string{"src/**"},
		Exclude:           []string{"src/generated/**"},
	}); err != nil {
		t.Fatalf("Generate (json): %v", err)
	}

	var payload Payload
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}

	if payload.SchemaVersion != "1.0" {
		t.Fatalf("expected schema_version 1.0, got %q", payload.SchemaVersion)
	}

	if payload.Summary.ExitCode != 1 {
		t.Fatalf("expected exit_code 1, got %d", payload.Summary.ExitCode)
	}

	if payload.Summary.BreachedFunctions != 2 {
		t.Fatalf("expected breached_functions 2, got %d", payload.Summary.BreachedFunctions)
	}

	if payload.Filters.Language != "python" {
		t.Fatalf("expected language filter python, got %q", payload.Filters.Language)
	}

	if len(payload.Results) == 0 {
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

func TestMarkdownEscapesTableSpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	deltas := []interfaces.FunctionDelta{
		{FilePath: "weird|file.py", FunctionName: "func`name", OldComplexity: 1, NewComplexity: 3, Delta: 2},
	}
	if err := Generate(&buf, deltas, Options{Format: "markdown"}); err != nil {
		t.Fatalf("Generate markdown: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "weird\\|file.py") {
		t.Fatalf("expected escaped pipe in markdown output, got %q", out)
	}
	if !strings.Contains(out, "func\\`name") {
		t.Fatalf("expected escaped backtick in markdown output, got %q", out)
	}
}
