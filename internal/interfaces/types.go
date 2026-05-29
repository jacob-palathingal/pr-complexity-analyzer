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
	FilePath string

	// FunctionName is the function or method name.
	FunctionName string

	// OldComplexity is the complexity at the base ref. 0 means the function
	// did not exist before (it was added in this PR).
	OldComplexity int

	// NewComplexity is the complexity at the head ref.
	NewComplexity int

	// Delta is NewComplexity - OldComplexity. Positive means more complex.
	// Negative means simplified. Zero means no change.
	Delta int
}
