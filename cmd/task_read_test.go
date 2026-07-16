package cmd

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskReadCmd_ShowsFullTask(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Full details", "--description", "Line one.\nLine two."})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	readCmd := NewRootCmd()
	buf := &bytes.Buffer{}
	readCmd.SetOut(buf)
	readCmd.SetArgs([]string{"task", "read", "BIT-1"})
	if err := readCmd.Execute(); err != nil {
		t.Fatalf("task read Execute() returned error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"BIT-1", "Full details", "todo", "Line one.", "Line two."} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want to contain %q", out, want)
		}
	}
}

func TestTaskReadCmd_RejectsPathTraversalID(t *testing.T) {
	// Arrange: a real .md file outside .bit/tasks/, plus a legitimate task.
	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# real project readme\n"), 0o644); err != nil {
		t.Fatalf("writing README fixture: %v", err)
	}

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Real task"})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	// Act: read an ID that escapes two levels above tasksDir to reach the README.
	readCmd := NewRootCmd()
	buf := &bytes.Buffer{}
	readCmd.SetOut(buf)
	readCmd.SetArgs([]string{"task", "read", "../../README"})
	err := readCmd.Execute()

	// Assert: the ID was contained under .bit/tasks/ (a not-found error, not a
	// successful read of the file it lexically points at) and nothing from the
	// escaped file reached the output.
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Execute() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
	if strings.Contains(buf.String(), "real project readme") {
		t.Errorf("output = %q, must not contain README fixture content", buf.String())
	}
}

func TestTaskReadCmd_ErrorsOnUnknownID(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	readCmd := NewRootCmd()
	readCmd.SetArgs([]string{"task", "read", "BIT-99"})
	if err := readCmd.Execute(); err == nil {
		t.Fatal("Execute() returned nil error, want error for unknown task ID")
	}
}
