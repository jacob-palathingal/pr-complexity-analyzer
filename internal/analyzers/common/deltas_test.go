package common

import "testing"

func TestBuildDeltasIsDeterministicAndComplete(t *testing.T) {
	oldScores := map[string]int{
		"Removed": 2,
		"Changed": 5,
	}

	newScores := map[string]int{
		"Changed": 8,
		"Added":   1,
	}

	deltas := BuildDeltas(
		"service.go",
		"go",
		"go/ast",
		oldScores,
		newScores,
	)

	if len(deltas) != 3 {
		t.Fatalf(
			"delta count = %d, want 3",
			len(deltas),
		)
	}

	expectedNames := []string{
		"Added",
		"Changed",
		"Removed",
	}

	for index, expectedName := range expectedNames {
		if deltas[index].FunctionName != expectedName {
			t.Fatalf(
				"delta %d function = %q, want %q",
				index,
				deltas[index].FunctionName,
				expectedName,
			)
		}
	}

	added := deltas[0]
	if added.OldComplexity != 0 ||
		added.NewComplexity != 1 ||
		added.Delta != 1 {
		t.Fatalf(
			"unexpected added-function delta: %+v",
			added,
		)
	}

	changed := deltas[1]
	if changed.OldComplexity != 5 ||
		changed.NewComplexity != 8 ||
		changed.Delta != 3 {
		t.Fatalf(
			"unexpected changed-function delta: %+v",
			changed,
		)
	}

	removed := deltas[2]
	if removed.OldComplexity != 2 ||
		removed.NewComplexity != 0 ||
		removed.Delta != -2 {
		t.Fatalf(
			"unexpected removed-function delta: %+v",
			removed,
		)
	}
}

func TestBuildDeltasIncludesReportMetadata(t *testing.T) {
	deltas := BuildDeltas(
		"src/auth.py",
		"python",
		"python/radon",
		map[string]int{
			"authenticate": 2,
		},
		map[string]int{
			"authenticate": 6,
		},
	)

	if len(deltas) != 1 {
		t.Fatalf(
			"delta count = %d, want 1",
			len(deltas),
		)
	}

	delta := deltas[0]

	if delta.FilePath != "src/auth.py" {
		t.Fatalf(
			"file path = %q, want src/auth.py",
			delta.FilePath,
		)
	}

	if delta.Language != "python" {
		t.Fatalf(
			"language = %q, want python",
			delta.Language,
		)
	}

	if delta.Analyzer != "python/radon" {
		t.Fatalf(
			"analyzer = %q, want python/radon",
			delta.Analyzer,
		)
	}
}

func TestBuildDeltasHandlesEmptyInputs(t *testing.T) {
	deltas := BuildDeltas(
		"empty.go",
		"go",
		"go/ast",
		map[string]int{},
		map[string]int{},
	)

	if len(deltas) != 0 {
		t.Fatalf(
			"delta count = %d, want 0",
			len(deltas),
		)
	}
}
