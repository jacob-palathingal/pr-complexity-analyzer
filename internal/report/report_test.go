package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

var testDeltas = []interfaces.FunctionDelta{
	{
		FilePath:      "auth.py",
		FunctionName:  "login",
		Language:      "python",
		Analyzer:      "python/radon",
		OldComplexity: 2,
		NewComplexity: 7,
		Delta:         5,
	},
	{
		FilePath:      "auth.py",
		FunctionName:  "logout",
		Language:      "python",
		Analyzer:      "python/radon",
		OldComplexity: 1,
		NewComplexity: 1,
		Delta:         0,
	},
	{
		FilePath:      "utils.py",
		FunctionName:  "parse",
		Language:      "python",
		Analyzer:      "python/radon",
		OldComplexity: 3,
		NewComplexity: 8,
		Delta:         5,
	},
	{
		FilePath:      "models.py",
		FunctionName:  "validate",
		Language:      "python",
		Analyzer:      "python/radon",
		OldComplexity: 4,
		NewComplexity: 2,
		Delta:         -2,
	},
	{
		FilePath:      "new.py",
		FunctionName:  "fresh",
		Language:      "python",
		Analyzer:      "python/radon",
		OldComplexity: 0,
		NewComplexity: 3,
		Delta:         3,
	},
}

func TestGenerateTextOutput(t *testing.T) {
	var buf bytes.Buffer

	err := Generate(&buf, testDeltas, Options{
		Format: "text",
	})
	if err != nil {
		t.Fatalf("Generate text output: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "PR Complexity Report") {
		t.Fatalf("expected report heading, got %q", output)
	}

	if !strings.Contains(output, "login") {
		t.Fatalf("expected login function in output, got %q", output)
	}

	if !strings.Contains(output, "Analyzed: 5 function(s)") {
		t.Fatalf("expected analyzed-function summary, got %q", output)
	}

	if strings.Contains(output, "logout") {
		t.Fatalf("unchanged function should be excluded by default, got %q", output)
	}
}

func TestGenerateJSONOutputUsesStableSchema(t *testing.T) {
	var buf bytes.Buffer

	generatedAt := time.Date(
		2026,
		time.July,
		6,
		12,
		30,
		0,
		0,
		time.UTC,
	)

	err := Generate(&buf, testDeltas, Options{
		Format:            "json",
		GeneratedAt:       generatedAt,
		Threshold:         5,
		BreachedFunctions: 2,
		ExitCode:          1,
		LangFilter:        "python",
		MinDelta:          1,
		IncludeUnchanged:  false,
		Include:           []string{"src/**"},
		Exclude:           []string{"src/generated/**"},
	})
	if err != nil {
		t.Fatalf("Generate JSON output: %v", err)
	}

	var payload Payload
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf(
			"unmarshal JSON output: %v\noutput: %s",
			err,
			buf.String(),
		)
	}

	if payload.SchemaVersion != "1.0" {
		t.Fatalf(
			"schema version = %q, want %q",
			payload.SchemaVersion,
			"1.0",
		)
	}

	if !payload.GeneratedAt.Equal(generatedAt) {
		t.Fatalf(
			"generated_at = %s, want %s",
			payload.GeneratedAt,
			generatedAt,
		)
	}

	if payload.Summary.AnalyzedFunctions != len(testDeltas) {
		t.Fatalf(
			"analyzed functions = %d, want %d",
			payload.Summary.AnalyzedFunctions,
			len(testDeltas),
		)
	}

	if payload.Summary.ReportedFunctions != 3 {
		t.Fatalf(
			"reported functions = %d, want 3",
			payload.Summary.ReportedFunctions,
		)
	}

	if payload.Summary.Threshold != 5 {
		t.Fatalf(
			"threshold = %d, want 5",
			payload.Summary.Threshold,
		)
	}

	if payload.Summary.BreachedFunctions != 2 {
		t.Fatalf(
			"breached functions = %d, want 2",
			payload.Summary.BreachedFunctions,
		)
	}

	if payload.Summary.MaxDelta != 5 {
		t.Fatalf(
			"max delta = %d, want 5",
			payload.Summary.MaxDelta,
		)
	}

	if payload.Summary.ExitCode != 1 {
		t.Fatalf(
			"exit code = %d, want 1",
			payload.Summary.ExitCode,
		)
	}

	if payload.Filters.Language != "python" {
		t.Fatalf(
			"language filter = %q, want python",
			payload.Filters.Language,
		)
	}

	if payload.Filters.MinDelta != 1 {
		t.Fatalf(
			"minimum delta = %d, want 1",
			payload.Filters.MinDelta,
		)
	}

	if len(payload.Filters.Include) != 1 ||
		payload.Filters.Include[0] != "src/**" {
		t.Fatalf(
			"unexpected include filters: %#v",
			payload.Filters.Include,
		)
	}

	if len(payload.Filters.Exclude) != 1 ||
		payload.Filters.Exclude[0] != "src/generated/**" {
		t.Fatalf(
			"unexpected exclude filters: %#v",
			payload.Filters.Exclude,
		)
	}

	if len(payload.Results) != 3 {
		t.Fatalf(
			"result count = %d, want 3",
			len(payload.Results),
		)
	}

	for _, result := range payload.Results {
		if result.Delta < 1 {
			t.Fatalf(
				"result with delta below minimum was included: %+v",
				result,
			)
		}

		if result.Language == "" {
			t.Fatalf(
				"result is missing language metadata: %+v",
				result,
			)
		}

		if result.Analyzer == "" {
			t.Fatalf(
				"result is missing analyzer metadata: %+v",
				result,
			)
		}
	}
}

func TestGenerateJSONOutputContainsNoHumanProse(t *testing.T) {
	var buf bytes.Buffer

	err := Generate(&buf, testDeltas, Options{
		Format:            "json",
		Threshold:         5,
		BreachedFunctions: 2,
		ExitCode:          1,
	})
	if err != nil {
		t.Fatalf("Generate JSON output: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "exceeded the complexity threshold") {
		t.Fatalf(
			"JSON output must not contain human-readable threshold prose: %q",
			output,
		)
	}

	var payload Payload
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestGenerateMarkdownOutput(t *testing.T) {
	var buf bytes.Buffer

	err := Generate(&buf, testDeltas, Options{
		Format:            "markdown",
		Threshold:         5,
		BreachedFunctions: 2,
		ExitCode:          1,
	})
	if err != nil {
		t.Fatalf("Generate Markdown output: %v", err)
	}

	output := buf.String()

	expectedStrings := []string{
		"## PR Complexity Report",
		"| File | Function | Before | After | Delta |",
		"Threshold: `+5`",
		"Breaches: `2`",
		"exceeded the configured complexity threshold",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected Markdown output to contain %q, got %q",
				expected,
				output,
			)
		}
	}
}

func TestGenerateDefaultsToText(t *testing.T) {
	var buf bytes.Buffer

	if err := Generate(&buf, testDeltas, Options{}); err != nil {
		t.Fatalf("Generate default output: %v", err)
	}

	if !strings.Contains(buf.String(), "PR Complexity Report") {
		t.Fatalf(
			"default format should be text, got %q",
			buf.String(),
		)
	}
}

func TestGenerateEmptyJSONOutput(t *testing.T) {
	var buf bytes.Buffer

	if err := Generate(&buf, nil, Options{
		Format: "json",
	}); err != nil {
		t.Fatalf("Generate empty JSON output: %v", err)
	}

	var payload Payload
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal empty JSON output: %v", err)
	}

	if payload.Summary.AnalyzedFunctions != 0 {
		t.Fatalf(
			"analyzed functions = %d, want 0",
			payload.Summary.AnalyzedFunctions,
		)
	}

	if payload.Summary.ReportedFunctions != 0 {
		t.Fatalf(
			"reported functions = %d, want 0",
			payload.Summary.ReportedFunctions,
		)
	}

	if len(payload.Results) != 0 {
		t.Fatalf(
			"results = %#v, want empty results",
			payload.Results,
		)
	}
}

func TestFilterExcludesUnchangedByDefault(t *testing.T) {
	results := filter(testDeltas, Options{
		IncludeUnchanged: false,
	})

	for _, result := range results {
		if result.Delta == 0 {
			t.Fatalf(
				"unchanged function %q should have been excluded",
				result.FunctionName,
			)
		}
	}
}

func TestFilterIncludesUnchangedWhenRequested(t *testing.T) {
	results := filter(testDeltas, Options{
		IncludeUnchanged: true,
	})

	var found bool

	for _, result := range results {
		if result.FunctionName == "logout" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal(
			"expected unchanged logout function when IncludeUnchanged is true",
		)
	}
}

func TestFilterAppliesMinimumDelta(t *testing.T) {
	results := filter(testDeltas, Options{
		MinDelta: 4,
	})

	if len(results) != 2 {
		t.Fatalf(
			"result count = %d, want 2",
			len(results),
		)
	}

	for _, result := range results {
		if result.Delta < 4 {
			t.Fatalf(
				"function %q has delta %d below minimum 4",
				result.FunctionName,
				result.Delta,
			)
		}
	}
}

func TestSortDeltasUsesDeterministicTieBreaking(t *testing.T) {
	deltas := []interfaces.FunctionDelta{
		{
			FilePath:     "b.go",
			FunctionName: "z",
			Delta:        5,
		},
		{
			FilePath:     "a.go",
			FunctionName: "z",
			Delta:        5,
		},
		{
			FilePath:     "a.go",
			FunctionName: "a",
			Delta:        5,
		},
		{
			FilePath:     "c.go",
			FunctionName: "c",
			Delta:        3,
		},
	}

	sortDeltas(deltas)

	expectedOrder := []struct {
		file     string
		function string
		delta    int
	}{
		{"a.go", "a", 5},
		{"a.go", "z", 5},
		{"b.go", "z", 5},
		{"c.go", "c", 3},
	}

	for index, expected := range expectedOrder {
		actual := deltas[index]

		if actual.FilePath != expected.file ||
			actual.FunctionName != expected.function ||
			actual.Delta != expected.delta {
			t.Fatalf(
				"result %d = %+v, want file=%q function=%q delta=%d",
				index,
				actual,
				expected.file,
				expected.function,
				expected.delta,
			)
		}
	}
}

func TestMarkdownEscapesTableSpecialCharacters(t *testing.T) {
	var buf bytes.Buffer

	deltas := []interfaces.FunctionDelta{
		{
			FilePath:      "weird|file.py",
			FunctionName:  "func`name",
			OldComplexity: 1,
			NewComplexity: 3,
			Delta:         2,
		},
	}

	if err := Generate(&buf, deltas, Options{
		Format: "markdown",
	}); err != nil {
		t.Fatalf("Generate Markdown output: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "weird\\|file.py") {
		t.Fatalf(
			"expected escaped pipe in Markdown output, got %q",
			output,
		)
	}

	if !strings.Contains(output, "func\\`name") {
		t.Fatalf(
			"expected escaped backtick in Markdown output, got %q",
			output,
		)
	}
}
