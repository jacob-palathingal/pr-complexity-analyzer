// Package diff handles all git interaction: resolving refs to commit SHAs,
// listing files changed between two refs, and reading file content at a
// specific ref. It never calls language analyzers.
package diff

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultGitTimeout = 30 * time.Second

// Client runs git commands against a repository rooted at RepoDir.
// An empty RepoDir means "current working directory".
type Client struct {
	RepoDir string
	Timeout time.Duration
}

// NewClient returns a Client for the repo at dir. Pass "" to use cwd.
func NewClient(dir string) *Client {
	return NewClientWithTimeout(dir, defaultGitTimeout)
}

// NewClientWithTimeout returns a Client with a per-git-command timeout.
func NewClientWithTimeout(dir string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = defaultGitTimeout
	}

	return &Client{
		RepoDir: dir,
		Timeout: timeout,
	}
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
// Deleted files are excluded because there is no head-side content to analyze.
func (c *Client) ChangedFiles(baseRef, headRef string) ([]string, error) {
	out, err := c.git("diff", "--name-only", "--diff-filter=d", baseRef+"..."+headRef, "--")
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
// It returns an empty string if the file did not exist at that ref.
func (c *Client) FileContentAt(ref, path string) (string, error) {
	stdout, stderr, err := c.gitRaw("show", ref+":"+path)
	if err != nil {
		if isMissingPathAtRef(stderr) {
			return "", nil
		}

		return "", fmt.Errorf("git show %s:%s: %w — %s", ref, path, err, strings.TrimSpace(stderr))
	}

	return stdout, nil
}

// UnifiedDiff returns the raw unified diff for path between baseRef and headRef.
// Useful for display; parsing uses FileContentAt on both sides instead.
func (c *Client) UnifiedDiff(baseRef, headRef, path string) (string, error) {
	cmd := c.buildCmd("diff", baseRef, headRef, "--", path)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run()

	if stderr.Len() > 0 {
		return "", fmt.Errorf("git diff stderr: %s", strings.TrimSpace(stderr.String()))
	}

	return stdout.String(), nil
}

func (c *Client) git(args ...string) (string, error) {
	stdout, stderr, err := c.gitRaw(args...)
	if err != nil {
		return "", fmt.Errorf("%w — %s", err, strings.TrimSpace(stderr))
	}

	return stdout, nil
}

func (c *Client) gitRaw(args ...string) (stdout string, stderr string, err error) {
	cmd := c.buildCmd(args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func (c *Client) buildCmd(args ...string) *exec.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), c.effectiveTimeout())

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Cancel = func() error {
		cancel()
		return nil
	}

	if c.RepoDir != "" {
		cmd.Dir = c.RepoDir
	}

	return cmd
}

func (c *Client) effectiveTimeout() time.Duration {
	if c.Timeout <= 0 {
		return defaultGitTimeout
	}

	return c.Timeout
}

func isMissingPathAtRef(stderr string) bool {
	s := strings.ToLower(stderr)

	return strings.Contains(s, "path") &&
		(strings.Contains(s, "does not exist") ||
			strings.Contains(s, "exists on disk, but not in"))
}
