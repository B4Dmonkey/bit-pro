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

func TestTaskReadCmd_BodyOnly(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Full details", "Line one.\nLine two.")

	out := mustRun(t, "task", "read", "BIT-1", "--body")

	if out != "Line one.\nLine two." {
		t.Errorf("output = %q, want %q", out, "Line one.\nLine two.")
	}
}

func TestTaskReadCmd_BodyOnlyEmpty(t *testing.T) {
	initProject(t, "BIT")
	mustRun(t, "task", "create", "No body", "-d", "")

	out := mustRun(t, "task", "read", "BIT-1", "--body")

	if out != "" {
		t.Errorf("output = %q, want %q", out, "")
	}
}

func TestTaskReadCmd_ShowsPhase(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "List cmd", "-d", "...", "--parent", "BIT-1",
		"--phase", "2", "--phase-label", "List & read")

	out := mustRun(t, "task", "read", "BIT-1.1")

	firstLine := strings.SplitN(out, "\n", 2)[0]

	want := "BIT-1.1\ttodo\tList cmd\tphase 2 — List & read"
	if firstLine != want {
		t.Errorf("first line = %q, want %q", firstLine, want)
	}
}

func TestTaskReadCmd_OmitsPhaseWhenAbsent(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Title", "Body")

	out := mustRun(t, "task", "read", "BIT-1")

	want := "BIT-1\ttodo\tTitle\n\nBody"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}

	data, err := os.ReadFile(".bit/tasks/BIT-1.md")
	if err != nil {
		t.Fatalf("reading task file: %v", err)
	}

	if strings.Contains(string(data), "phase") {
		t.Errorf("task file = %q, want no phase key", data)
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
	if err := os.WriteFile(readme, []byte("# real project readme\n"), 0o600); err != nil {
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
