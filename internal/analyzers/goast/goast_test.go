package goast

import (
	"testing"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

const goBefore = `
package sample

type Server struct{}

func handle(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

func (s *Server) Route(method string) int {
	return 1
}
`

const goAfter = `
package sample

type Server struct{}

func handle(ok bool, retry bool) int {
	if ok && retry {
		return 1
	}
	if ok || retry {
		return 2
	}
	return 0
}

func (s *Server) Route(method string) int {
	switch method {
	case "GET":
		return 1
	case "POST":
		return 2
	default:
		return 0
	}
}
`

func TestAnalyzer_Supports(t *testing.T) {
	a := New()
	cases := []struct {
		path string
		want bool
	}{
		{"main.go", true},
		{"internal/runner/runner.go", true},
		{"script.py", false},
		{"README.md", false},
	}

	for _, c := range cases {
		if got := a.Supports(c.path); got != c.want {
			t.Errorf("Supports(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestAnalyze_IncreaseDetected(t *testing.T) {
	a := New()
	deltas, err := a.Analyze("sample.go", goBefore, goAfter)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	assertDeltaPositive(t, deltas, "handle")
	assertDeltaPositive(t, deltas, "Server.Route")
}

func TestAnalyze_NewFile(t *testing.T) {
	a := New()
	deltas, err := a.Analyze("new.go", "", goAfter)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(deltas) == 0 {
		t.Fatal("expected deltas for new file")
	}
	for _, d := range deltas {
		if d.OldComplexity != 0 {
			t.Errorf("new file OldComplexity = %d, want 0", d.OldComplexity)
		}
		if d.Delta != d.NewComplexity {
			t.Errorf("delta should equal new complexity, got delta=%d new=%d", d.Delta, d.NewComplexity)
		}
	}
}

func TestAnalyze_SyntaxError(t *testing.T) {
	a := New()
	_, err := a.Analyze("bad.go", "package bad\n", "package bad\nfunc broken(")
	if err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestComplexity_SkipsNestedFunctionLiterals(t *testing.T) {
	a := New()
	content := `
package sample

func outer(ok bool) int {
	fn := func(x int) int {
		if x > 0 {
			return x
		}
		return 0
	}
	return fn(1)
}
`
	scores, err := a.scoreContent("sample.go", content)
	if err != nil {
		t.Fatalf("scoreContent: %v", err)
	}
	if got := scores["outer"]; got != 1 {
		t.Fatalf("outer complexity = %d, want 1", got)
	}
}

func assertDeltaPositive(t *testing.T, deltas []interfaces.FunctionDelta, name string) {
	t.Helper()
	for _, d := range deltas {
		if d.FunctionName == name {
			if d.Delta <= 0 {
				t.Fatalf("%s delta = %d, want positive", name, d.Delta)
			}
			return
		}
	}
	t.Fatalf("%s not found in deltas", name)
}
