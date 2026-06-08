package runner

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestFindAnalyzer_LanguageAliases(t *testing.T) {
	if got := findAnalyzer("service.py", "py"); got == nil || got.Name() != "python/radon" {
		t.Fatalf("py alias should resolve python analyzer, got %#v", got)
	}
	if got := findAnalyzer("main.go", "golang"); got == nil || got.Name() != "go/ast" {
		t.Fatalf("golang alias should resolve go analyzer, got %#v", got)
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

func TestRun_GoFileEndToEndJSONThreshold(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "checkout", "-b", "main")

	writeFile(t, filepath.Join(dir, "main.go"), `package main

func handler(ok bool) int {
	return 1
}
`)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")

	writeFile(t, filepath.Join(dir, "main.go"), `package main

func handler(ok bool, retry bool) int {
	if ok && retry {
		return 1
	}
	return 0
}
`)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "increase complexity")

	var buf bytes.Buffer
	result, err := Run(&buf, Config{
		RepoDir:   dir,
		BaseRef:   "HEAD~1",
		HeadRef:   "HEAD",
		Format:    "json",
		Threshold: 2,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.ExitCode != ExitThreshold {
		t.Fatalf("expected threshold exit, got %d", result.ExitCode)
	}
	if strings.Contains(buf.String(), "exceeded the complexity threshold") {
		t.Fatalf("json output should not include threshold prose: %q", buf.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON, got error %v and output %q", err, buf.String())
	}
	if !strings.Contains(buf.String(), "handler") {
		t.Fatalf("expected handler in JSON output, got %q", buf.String())
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(out))
	}
}
