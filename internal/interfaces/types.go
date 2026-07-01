package interfaces

// FunctionComplexity holds the cyclomatic complexity score for a single function.
type FunctionComplexity struct {
	// Name is the fully-qualified function name as reported by the analyzer,
	// e.g. "MyClass.my_method" or "standalone_func".
	Name string

	// Complexity is the cyclomatic complexity score. A score of 1 means a
	// straight-line function with no branches.
	Complexity int
}

// FunctionDelta represents the change in complexity for one function between
// the base and head refs.
type FunctionDelta struct {
	// FilePath is the repository-relative path of the file, e.g. "src/auth.py".
	FilePath string `json:"file_path"`

	// FunctionName is the function or method name.
	FunctionName string `json:"function_name"`

	// Language is the language analyzer category, e.g. "python" or "go".
	Language string `json:"language,omitempty"`

	// Analyzer is the concrete analyzer implementation, e.g. "python/radon".
	Analyzer string `json:"analyzer,omitempty"`

	// OldComplexity is the complexity at the base ref. 0 means the function
	// did not exist before.
	OldComplexity int `json:"old_complexity"`

	// NewComplexity is the complexity at the head ref. 0 means the function
	// was removed.
	NewComplexity int `json:"new_complexity"`

	// Delta is NewComplexity - OldComplexity.
	Delta int `json:"delta"`
}
