package python

import (
	"os/exec"
	"testing"
)

// skipIfNoRadon skips the test if radon is not installed.
func skipIfNoRadon(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("radon"); err != nil {
		t.Skip("radon not installed; skipping integration test")
	}
}

const simpleBefore = `
def calculate(x):
    return x * 2
`

const simpleAfter = `
def calculate(x, mode):
    if mode == "double":
        return x * 2
    elif mode == "triple":
        return x * 3
    else:
        return x
`

const classBefore = `
class Auth:
    def login(self, user, password):
        if user and password:
            return True
        return False
`

const classAfter = `
class Auth:
    def login(self, user, password):
        if not user:
            raise ValueError("no user")
        if not password:
            raise ValueError("no password")
        if len(password) < 8:
            raise ValueError("too short")
        return True
`

func TestAnalyzer_Supports(t *testing.T) {
	a := New()
	cases := []struct {
		path string
		want bool
	}{
		{"auth.py", true},
		{"src/models/user.py", true},
		{"main.go", false},
		{"index.js", false},
		{"README.md", false},
	}
	for _, c := range cases {
		if got := a.Supports(c.path); got != c.want {
			t.Errorf("Supports(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestAnalyze_IncreaseDetected(t *testing.T) {
	skipIfNoRadon(t)
	a := New()

	deltas, err := a.Analyze("calc.py", simpleBefore, simpleAfter)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	var found bool
	for _, d := range deltas {
		if d.FunctionName == "calculate" {
			found = true
			if d.Delta <= 0 {
				t.Errorf("expected positive delta for calculate, got %d", d.Delta)
			}
		}
	}
	if !found {
		t.Error("function 'calculate' not found in deltas")
	}
}

func TestAnalyze_NewFile(t *testing.T) {
	skipIfNoRadon(t)
	a := New()

	// OldContent empty = brand-new file in the PR.
	deltas, err := a.Analyze("new.py", "", simpleAfter)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(deltas) == 0 {
		t.Fatal("expected deltas for new file, got none")
	}
	for _, d := range deltas {
		if d.OldComplexity != 0 {
			t.Errorf("new file: expected OldComplexity 0, got %d", d.OldComplexity)
		}
		if d.Delta != d.NewComplexity {
			t.Errorf("new file: Delta should equal NewComplexity, got delta=%d new=%d", d.Delta, d.NewComplexity)
		}
	}
}

func TestAnalyze_ClassMethod(t *testing.T) {
	skipIfNoRadon(t)
	a := New()

	deltas, err := a.Analyze("auth.py", classBefore, classAfter)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	var found bool
	for _, d := range deltas {
		if d.FunctionName == "Auth.login" {
			found = true
			if d.Delta <= 0 {
				t.Errorf("expected positive delta for Auth.login, got %d", d.Delta)
			}
		}
	}
	if !found {
		t.Error("method 'Auth.login' not found in deltas")
	}
}

func TestBuildDeltas_NoChange(t *testing.T) {
	scores := map[string]int{"foo": 3, "bar": 5}
	deltas := buildDeltas("x.py", scores, scores)

	for _, d := range deltas {
		if d.Delta != 0 {
			t.Errorf("expected 0 delta for unchanged function %s, got %d", d.FunctionName, d.Delta)
		}
	}
}

func TestBuildDeltas_RemovedFunction(t *testing.T) {
	old := map[string]int{"foo": 3, "removed": 7}
	new := map[string]int{"foo": 3}

	deltas := buildDeltas("x.py", old, new)

	var found bool
	for _, d := range deltas {
		if d.FunctionName == "removed" {
			found = true
			if d.NewComplexity != 0 {
				t.Errorf("removed function should have NewComplexity 0, got %d", d.NewComplexity)
			}
			if d.Delta != -7 {
				t.Errorf("removed function delta should be -7, got %d", d.Delta)
			}
		}
	}
	if !found {
		t.Error("removed function not in deltas")
	}
}
