package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskReadCmd_ShowsFullTask(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Full details", "Line one.\nLine two.")

	out := mustRun(t, "task", "read", "BIT-1")

	want := "BIT-1\ttodo\tFull details\n\nLine one.\nLine two."
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestTaskReadCmd_ErrorsOnUnknownID(t *testing.T) {
	initProject(t, "BIT")

	if _, err := run(t, "task", "read", "BIT-99"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown task ID")
	}
}

func TestTaskReadCmd_ContainsPathTraversalID(t *testing.T) {
	dir := initProject(t, "BIT")
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# real project readme\n"), 0o644); err != nil {
		t.Fatalf("writing README fixture: %v", err)
	}

	out, err := run(t, "task", "read", "../../README")

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Execute() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
	if strings.Contains(out, "real project readme") {
		t.Errorf("output = %q, must not contain the escaped file's content", out)
	}
}
