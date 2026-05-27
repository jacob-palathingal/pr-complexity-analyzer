package diff

// FileDiff holds the before and after content of a single file that was
// touched by the PR. Either side can be empty:
//   - OldContent == "" means the file was newly created in this PR.
//   - NewContent == "" means the file was deleted (callers filter these out,
//     but the type supports it for completeness).
type FileDiff struct {
	// Path is the repository-relative file path, e.g. "src/auth/login.py".
	Path string

	// OldContent is the full file content at the base ref.
	OldContent string

	// NewContent is the full file content at the head ref.
	NewContent string
}
