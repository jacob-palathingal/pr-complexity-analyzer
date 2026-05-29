package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

// JSONFormatter renders results as a structured JSON payload suitable for
// piping into CI scripts, GitHub Actions step outputs, or custom dashboards.
type JSONFormatter struct{}

// jsonReport is the top-level JSON envelope.
type jsonReport struct {
	GeneratedAt string              `json:"generated_at"`
	TotalCount  int                 `json:"total_count"`
	Results     []jsonFunctionDelta `json:"results"`
}

type jsonFunctionDelta struct {
	FilePath      string `json:"file_path"`
	FunctionName  string `json:"function_name"`
	OldComplexity int    `json:"old_complexity"`
	NewComplexity int    `json:"new_complexity"`
	Delta         int    `json:"delta"`
}

func (f *JSONFormatter) Format(w io.Writer, deltas []interfaces.FunctionDelta) error {
	results := make([]jsonFunctionDelta, len(deltas))
	for i, d := range deltas {
		results[i] = jsonFunctionDelta{
			FilePath:      d.FilePath,
			FunctionName:  d.FunctionName,
			OldComplexity: d.OldComplexity,
			NewComplexity: d.NewComplexity,
			Delta:         d.Delta,
		}
	}

	report := jsonReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		TotalCount:  len(deltas),
		Results:     results,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
