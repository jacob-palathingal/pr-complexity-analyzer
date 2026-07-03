package report

import (
	"encoding/json"
	"io"
)

// JSONFormatter renders the full stable payload for CI systems and dashboards.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(w io.Writer, payload Payload) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}
