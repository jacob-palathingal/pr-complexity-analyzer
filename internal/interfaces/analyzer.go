// Package interfaces defines the contracts that all language analyzers must satisfy.
// Adding support for a new language means implementing Analyzer — no other package
// needs to change.
package interfaces

// Analyzer is the pluggable interface for per-language complexity analysis.
// Each implementation is responsible for:
//   - Declaring which files it can handle (Supports).
//   - Computing cyclomatic complexity for every function in a file and
//     returning the delta between the old and new versions (Analyze).
type Analyzer interface {
	// Supports returns true if this analyzer can handle the given file path.
	// Implementations typically check the file extension, e.g. ".py".
	Supports(path string) bool

	// Analyze computes per-function cyclomatic complexity for oldContent and
	// newContent and returns the delta for every function that appears in at
	// least one version.
	//
	// oldContent may be empty if the file is new in this PR.
	// newContent is always non-empty (deletions are filtered upstream).
	Analyze(path, oldContent, newContent string) ([]FunctionDelta, error)

	// Name returns a human-readable identifier for this analyzer, e.g. "python/radon".
	Name() string
}
