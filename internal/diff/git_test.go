package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func import_os_write(path, content string) {
	_ = os.WriteFile(path, []byte(content), 0644)
}

func makeTestRepo(t *testing.T) (*Client, string) {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", args, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")
	run("git", "checkout", "-b", "main")

	import_os_write(filepath.Join(dir, "hello.py"), "def hi():\n    pass\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "initial")

	import_os_write(filepath.Join(dir, "hello.py"), "def hi():\n    x = 1\n    return x\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "modify")

	return NewClient(dir), dir
}

func TestResolveRef(t *testing.T) {
	client, _ := makeTestRepo(t)
	sha, err := client.ResolveRef("HEAD")
	if err != nil {
		t.Fatalf("ResolveRef(HEAD): %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("expected 40-char SHA, got %q", sha)
	}
}

func TestChangedFiles(t *testing.T) {
	client, _ := makeTestRepo(t)
	files, err := client.ChangedFiles("HEAD~1", "HEAD")
	if err != nil {
		t.Fatalf("ChangedFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "hello.py" {
		t.Errorf("expected [hello.py], got %v", files)
	}
}

func TestFileContentAt(t *testing.T) {
	client, _ := makeTestRepo(t)
	content, err := client.FileContentAt("HEAD~1", "hello.py")
	if err != nil {
		t.Fatalf("FileContentAt: %v", err)
	}
	if !strings.Contains(content, "def hi") {
		t.Errorf("expected file content, got %q", content)
	}
}

func TestFileContentAt_NewFile(t *testing.T) {
	client, _ := makeTestRepo(t)
	content, err := client.FileContentAt("HEAD~1", "does_not_exist.py")
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty string for missing file, got %q", content)
	}
}
