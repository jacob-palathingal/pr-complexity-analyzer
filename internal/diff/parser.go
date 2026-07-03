package diff

import "fmt"

// Parser builds FileDiff structs by combining the git Client's ref-resolution
// and file-content retrieval.
type Parser struct {
	client *Client
}

// NewParser wraps a Client in a Parser.
func NewParser(client *Client) *Parser {
	return &Parser{client: client}
}

// BuildDiffs resolves baseRef and headRef, lists changed files, and fetches
// the content of each file at both refs.
func (p *Parser) BuildDiffs(baseRef, headRef string) ([]FileDiff, error) {
	return p.BuildDiffsFiltered(baseRef, headRef, nil)
}

// BuildDiffsFiltered is BuildDiffs with an optional path-level predicate.
// Filtering before FileContentAt avoids loading unsupported or irrelevant blobs
// in large company repositories.
func (p *Parser) BuildDiffsFiltered(baseRef, headRef string, include func(path string) bool) ([]FileDiff, error) {
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
		if include != nil && !include(path) {
			continue
		}

		oldContent, err := p.client.FileContentAt(baseRef, path)
		if err != nil {
			return nil, fmt.Errorf("reading %s at %s: %w", path, baseRef, err)
		}

		newContent, err := p.client.FileContentAt(headRef, path)
		if err != nil {
			return nil, fmt.Errorf("reading %s at %s: %w", path, headRef, err)
		}

		diffs = append(diffs, FileDiff{
			Path:       path,
			OldContent: oldContent,
			NewContent: newContent,
		})
	}

	return diffs, nil
}
