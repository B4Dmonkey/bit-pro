package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	out := mustRun(t, "--help")

	if !strings.Contains(out, "bp") {
		t.Errorf("help output missing command name %q, got:\n%s", "bp", out)
	}
}

func TestRootCmd_Version(t *testing.T) {
	out := mustRun(t, "--version")

	want := "bp version " + version + "\n"
	if out != want {
		t.Errorf("version output = %q, want %q", out, want)
	}
}

func TestBitDir_OutsideWorktreeUsesRelativeDotBit(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from default .bit dir", out)
	}
}

func TestBitDir_InsideClaudeWorktreeResolvesToMainCheckout(t *testing.T) {
	root := initProject(t, "BIT")
	createTask(t, "Track", "...")

	worktree := filepath.Join(root, ".claude", "worktrees", "hazy-pondering-star")
	if err := os.MkdirAll(filepath.Join(worktree, ".bit"), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", worktree, err)
	}

	t.Chdir(worktree)

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from the main checkout's .bit", out)
	}
}

func TestBitDir_NestedWorktreeResolvesToOutermostCheckout(t *testing.T) {
	root := initProject(t, "BIT")
	createTask(t, "Track", "...")

	outer := filepath.Join(root, ".claude", "worktrees", "outer")
	if err := os.MkdirAll(filepath.Join(outer, ".bit"), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", outer, err)
	}

	nested := filepath.Join(outer, ".claude", "worktrees", "inner")
	if err := os.MkdirAll(filepath.Join(nested, ".bit"), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", nested, err)
	}

	t.Chdir(nested)

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from the outermost checkout's .bit", out)
	}
}

func TestRootCmd_RuntimeErrorOmitsUsage(t *testing.T) {
	initProject(t, "BIT")

	out, err := run(t, "task", "read", "BIT-99")
	if err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown task ID")
	}

	if strings.Contains(out, "Usage:") {
		t.Errorf("output = %q, want no usage text on a runtime failure", out)
	}
}
