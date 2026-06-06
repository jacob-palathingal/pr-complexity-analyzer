package runner

import (
	"bytes"
	"testing"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

func TestApplyThreshold_NoThreshold(t *testing.T) {
	cfg := Config{Threshold: 0}
	result := applyThreshold(&bytes.Buffer{}, cfg, []interfaces.FunctionDelta{
		{FunctionName: "big_func", Delta: 50},
	})
	if result.ExitCode != ExitOK {
		t.Errorf("threshold=0 should always return ExitOK, got %d", result.ExitCode)
	}
	if result.BreachedFunctions != 0 {
		t.Errorf("expected 0 breached functions, got %d", result.BreachedFunctions)
	}
}

func TestApplyThreshold_NotMet(t *testing.T) {
	cfg := Config{Threshold: 5}
	result := applyThreshold(&bytes.Buffer{}, cfg, []interfaces.FunctionDelta{
		{FunctionName: "a", Delta: 3},
		{FunctionName: "b", Delta: 1},
	})
	if result.ExitCode != ExitOK {
		t.Errorf("no breach should return ExitOK, got %d", result.ExitCode)
	}
}

func TestApplyThreshold_ExactlyMet(t *testing.T) {
	// Delta == Threshold should trigger (>= not >).
	cfg := Config{Threshold: 5}
	result := applyThreshold(&bytes.Buffer{}, cfg, []interfaces.FunctionDelta{
		{FunctionName: "exact", Delta: 5},
	})
	if result.ExitCode != ExitThreshold {
		t.Errorf("delta == threshold should trigger ExitThreshold, got %d", result.ExitCode)
	}
	if result.BreachedFunctions != 1 {
		t.Errorf("expected 1 breached function, got %d", result.BreachedFunctions)
	}
}

func TestApplyThreshold_MultipleBreaches(t *testing.T) {
	cfg := Config{Threshold: 5}
	result := applyThreshold(&bytes.Buffer{}, cfg, []interfaces.FunctionDelta{
		{FunctionName: "a", Delta: 3},  // under
		{FunctionName: "b", Delta: 7},  // over
		{FunctionName: "c", Delta: 5},  // exact
		{FunctionName: "d", Delta: -2}, // negative, never counts
	})
	if result.ExitCode != ExitThreshold {
		t.Errorf("expected ExitThreshold, got %d", result.ExitCode)
	}
	if result.BreachedFunctions != 2 {
		t.Errorf("expected 2 breached functions, got %d", result.BreachedFunctions)
	}
}

func TestApplyThreshold_NegativeDeltaIgnored(t *testing.T) {
	cfg := Config{Threshold: 3}
	result := applyThreshold(&bytes.Buffer{}, cfg, []interfaces.FunctionDelta{
		{FunctionName: "simplified", Delta: -8},
	})
	if result.ExitCode != ExitOK {
		t.Errorf("negative delta should never trigger threshold, got exit %d", result.ExitCode)
	}
}

func TestApplyThreshold_WritesBreachMessage(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{Threshold: 3}
	applyThreshold(&buf, cfg, []interfaces.FunctionDelta{
		{FunctionName: "complex", Delta: 6},
	})
	if buf.Len() == 0 {
		t.Error("expected breach message written to output")
	}
}

func TestApplyThreshold_JSONDoesNotWriteMessage(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{Threshold: 3, Format: "json"}
	result := applyThreshold(&buf, cfg, []interfaces.FunctionDelta{
		{FunctionName: "complex", Delta: 6},
	})
	if result.ExitCode != ExitThreshold {
		t.Fatalf("expected ExitThreshold, got %d", result.ExitCode)
	}
	if buf.Len() != 0 {
		t.Fatalf("json threshold handling should not write prose to stdout, got %q", buf.String())
	}
}

func TestNormalizeFormat_DefaultsToText(t *testing.T) {
	if got := normalizeFormat(""); got != "text" {
		t.Fatalf("normalizeFormat empty = %q, want text", got)
	}
}
