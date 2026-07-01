package common

import (
	"sort"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// BuildDeltas merges old and new complexity maps into deterministic,
// metadata-rich FunctionDelta values.
func BuildDeltas(filePath, language, analyzerName string, oldScores, newScores map[string]int) []interfaces.FunctionDelta {
	seen := make(map[string]struct{}, len(oldScores)+len(newScores))

	for name := range oldScores {
		seen[name] = struct{}{}
	}

	for name := range newScores {
		seen[name] = struct{}{}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)

	deltas := make([]interfaces.FunctionDelta, 0, len(names))
	for _, name := range names {
		oldComplexity := oldScores[name]
		newComplexity := newScores[name]

		deltas = append(deltas, interfaces.FunctionDelta{
			FilePath:      filePath,
			FunctionName:  name,
			Language:      language,
			Analyzer:      analyzerName,
			OldComplexity: oldComplexity,
			NewComplexity: newComplexity,
			Delta:         newComplexity - oldComplexity,
		})
	}

	return deltas
}
