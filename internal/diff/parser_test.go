package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Shared test helpers for the diff package
func import_os_write(path, content string) {
	_ = os.WriteFile(path, []byte(content), 0644)
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	return cmd.Run()
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

// Parser tests

func TestBuildDiffs(t *testing.T) {
	client, _ := makeTestRepo(t)
	parser := NewParser(client)

	diffs, err := parser.BuildDiffs("HEAD~1", "HEAD")
	if err != nil {
		t.Fatalf("BuildDiffs: %v", err)
	}
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}

	d := diffs[0]
	if d.Path != "hello.py" {
		t.Errorf("expected path hello.py, got %q", d.Path)
	}
	if d.OldContent == "" {
		t.Error("OldContent should not be empty")
	}
	if d.NewContent == "" {
		t.Error("NewContent should not be empty")
	}
	if d.OldContent == d.NewContent {
		t.Error("OldContent and NewContent should differ")
	}
}

func TestBuildDiffs_NewFile(t *testing.T) {
	client, dir := makeTestRepo(t)

	newPath := filepath.Join(dir, "new_file.py")
	if err := os.WriteFile(newPath, []byte("def new_func(): pass\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runGit(dir, "git", "add", "."); err != nil {
		t.Fatal(err)
	}
	if err := runGit(dir, "git", "commit", "-m", "add new file"); err != nil {
		t.Fatal(err)
	}

	parser := NewParser(client)
	diffs, err := parser.BuildDiffs("HEAD~1", "HEAD")
	if err != nil {
		t.Fatalf("BuildDiffs: %v", err)
	}

	var found bool
	for _, d := range diffs {
		if d.Path == "new_file.py" {
			found = true
			if d.OldContent != "" {
				t.Errorf("new file should have empty OldContent, got %q", d.OldContent)
			}
			if d.NewContent == "" {
				t.Error("new file NewContent should not be empty")
			}
		}
	}
	if !found {
		t.Error("new_file.py not found in diffs")
	}
}

func TestBuildDiffs_EmptiedFileIsStillIncluded(t *testing.T) {
	client, dir := makeTestRepo(t)

	emptyPath := filepath.Join(dir, "hello.py")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runGit(dir, "git", "add", "."); err != nil {
		t.Fatal(err)
	}
	if err := runGit(dir, "git", "commit", "-m", "empty file"); err != nil {
		t.Fatal(err)
	}

	parser := NewParser(client)
	diffs, err := parser.BuildDiffs("HEAD~1", "HEAD")
	if err != nil {
		t.Fatalf("BuildDiffs: %v", err)
	}
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Path != "hello.py" {
		t.Fatalf("expected hello.py, got %q", diffs[0].Path)
	}
	if diffs[0].NewContent != "" {
		t.Fatalf("expected empty NewContent for emptied file, got %q", diffs[0].NewContent)
	}
}
