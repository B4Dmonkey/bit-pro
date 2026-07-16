package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestTaskDeleteCmd_RemovesFileWithYesFlag(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Throwaway", "--description", "Delete me."})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	deleteCmd := NewRootCmd()
	deleteCmd.SetArgs([]string{"task", "delete", "BIT-1", "--yes"})
	if err := deleteCmd.Execute(); err != nil {
		t.Fatalf("task delete Execute() returned error: %v", err)
	}

	if _, err := os.Stat(".bit/tasks/BIT-1.md"); !os.IsNotExist(err) {
		t.Errorf("os.Stat(%q) error = %v, want IsNotExist", ".bit/tasks/BIT-1.md", err)
	}
}

func TestTaskDeleteCmd_ErrorsOnUnknownID(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	deleteCmd := NewRootCmd()
	deleteCmd.SetArgs([]string{"task", "delete", "BIT-99", "--yes"})
	if err := deleteCmd.Execute(); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown ID")
	}
}

func TestTaskDeleteCmd_RejectsPathTraversalID(t *testing.T) {
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

	deleteCmd := NewRootCmd()
	deleteCmd.SetArgs([]string{"task", "delete", "../../README", "--yes"})
	err := deleteCmd.Execute()

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Execute() error = %v, want an error wrapping fs.ErrNotExist", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("reading README fixture after delete attempt: %v", err)
	}
	if string(got) != "# real project readme\n" {
		t.Errorf("README fixture = %q, want unchanged", got)
	}
}
