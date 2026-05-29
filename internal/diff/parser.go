package diff

import "fmt"

// Parser builds FileDiff structs by combining the git Client's ref-resolution
// and file-content retrieval. It is the entry point for the rest of the
// application — call BuildDiffs, get back a slice ready for analyzers.
type Parser struct {
	client *Client
}

// NewParser wraps a Client in a Parser.
func NewParser(client *Client) *Parser {
	return &Parser{client: client}
}

// BuildDiffs resolves baseRef and headRef, lists changed files, and fetches
// the content of each file at both refs. It skips files where NewContent is
// empty (pure deletions).
func (p *Parser) BuildDiffs(baseRef, headRef string) ([]FileDiff, error) {
	if _, err := p.client.ResolveRef(baseRef); err != nil {
		return nil, fmt.Errorf("base ref: %w", err)
	}
	if _, err := p.client.ResolveRef(headRef); err != nil {
		return nil, fmt.Errorf("head ref: %w", err)
	}

	paths, err := p.client.ChangedFiles(baseRef, headRef)
	if err != nil {
		return nil, err
	}

	diffs := make([]FileDiff, 0, len(paths))
	for _, path := range paths {
		oldContent, err := p.client.FileContentAt(baseRef, path)
		if err != nil {
			return nil, fmt.Errorf("reading %s at %s: %w", path, baseRef, err)
		}

		newContent, err := p.client.FileContentAt(headRef, path)
		if err != nil {
			return nil, fmt.Errorf("reading %s at %s: %w", path, headRef, err)
		}

		// Skip pure deletions.
		if newContent == "" {
			continue
		}

		diffs = append(diffs, FileDiff{
			Path:       path,
			OldContent: oldContent,
			NewContent: newContent,
		})
	}

	return diffs, nil
}
