// Package diff handles all git interaction: resolving refs to commit SHAs,
// listing files changed between two refs, and reading file content at a
// specific ref. It never calls language analyzers — it only produces raw
// file content for other packages to consume.
package diff

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Client runs git commands against a repository rooted at RepoDir.
// An empty RepoDir means "current working directory".
type Client struct {
	RepoDir string
}

// NewClient returns a Client for the repo at dir. Pass "" to use cwd.
func NewClient(dir string) *Client {
	return &Client{RepoDir: dir}
}

// ResolveRef validates that ref exists and returns its full commit SHA.
func (c *Client) ResolveRef(ref string) (string, error) {
	out, err := c.git("rev-parse", "--verify", ref)
	if err != nil {
		return "", fmt.Errorf("cannot resolve ref %q: %w", ref, err)
	}
	return strings.TrimSpace(out), nil
}

// ChangedFiles returns the list of files that differ between baseRef and headRef.
// Deleted files are excluded — there is nothing to analyze in a deletion.
func (c *Client) ChangedFiles(baseRef, headRef string) ([]string, error) {
	// --diff-filter=d excludes deleted files (D), --name-only gives just paths.
	out, err := c.git("diff", "--name-only", "--diff-filter=d", baseRef+"..."+headRef)
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// FileContentAt returns the full content of path at the given git ref.
// Returns an empty string (not an error) if the file did not exist at that ref.
func (c *Client) FileContentAt(ref, path string) (string, error) {
	out, err := c.git("show", ref+":"+path)
	if err != nil {
		// Exit code 128 means the object doesn't exist (new file in the PR).
		if strings.Contains(err.Error(), "exit status 128") {
			return "", nil
		}
		return "", fmt.Errorf("git show %s:%s: %w", ref, path, err)
	}
	return out, nil
}

// UnifiedDiff returns the raw unified diff for path between baseRef and headRef.
// Useful for display; parsing uses FileContentAt on both sides instead.
func (c *Client) UnifiedDiff(baseRef, headRef, path string) (string, error) {
	// git diff can exit 1 when there are differences — that's expected.
	cmd := c.buildCmd("diff", baseRef, headRef, "--", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run() // intentionally ignore exit code
	if stderr.Len() > 0 {
		return "", fmt.Errorf("git diff stderr: %s", stderr.String())
	}
	return stdout.String(), nil
}

// git runs a git command and returns combined stdout. Stderr is captured and
// returned as part of the error message.
func (c *Client) git(args ...string) (string, error) {
	cmd := c.buildCmd(args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w — %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func (c *Client) buildCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	if c.RepoDir != "" {
		cmd.Dir = c.RepoDir
	}
	return cmd
}
