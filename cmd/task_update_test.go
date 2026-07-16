package cmd

import (
	"strings"
	"testing"
)

func TestTaskUpdateCmd_ChangesTitleOnly(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Old title", "--description", "Body text."})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	updateCmd := NewRootCmd()
	updateCmd.SetArgs([]string{"task", "update", "BIT-1", "--title", "New title"})
	err := updateCmd.Execute()

	if err != nil {
		t.Fatalf("task update Execute() returned error: %v", err)
	}

	got, err := loadTask("BIT-1")
	if err != nil {
		t.Fatalf("loadTask(%q) returned error: %v", "BIT-1", err)
	}
	if got.Title != "New title" {
		t.Errorf("Title = %q, want %q", got.Title, "New title")
	}
	if got.Status != "todo" {
		t.Errorf("Status = %q, want unchanged %q", got.Status, "todo")
	}
	if got.Body != "Body text." {
		t.Errorf("Body = %q, want unchanged %q", got.Body, "Body text.")
	}
	if got.ID != "BIT-1" {
		t.Errorf("ID = %q, want unchanged %q", got.ID, "BIT-1")
	}
}

func TestTaskUpdateCmd_ChangesDescriptionAndStatus(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Keep this title", "--description", "Old body."})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	updateCmd := NewRootCmd()
	updateCmd.SetArgs([]string{"task", "update", "BIT-1", "--description", "New body.", "--status", "doing"})
	err := updateCmd.Execute()

	if err != nil {
		t.Fatalf("task update Execute() returned error: %v", err)
	}

	got, err := loadTask("BIT-1")
	if err != nil {
		t.Fatalf("loadTask(%q) returned error: %v", "BIT-1", err)
	}
	if got.Title != "Keep this title" {
		t.Errorf("Title = %q, want unchanged %q", got.Title, "Keep this title")
	}
	if !strings.Contains(got.Body, "New body.") {
		t.Errorf("Body = %q, want to contain %q", got.Body, "New body.")
	}
	if got.Status != "doing" {
		t.Errorf("Status = %q, want %q", got.Status, "doing")
	}
}

func TestTaskUpdateCmd_NoFlagsIsANoOp(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Title", "--description", "Body."})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	before, err := loadTask("BIT-1")
	if err != nil {
		t.Fatalf("loadTask(%q) returned error: %v", "BIT-1", err)
	}

	updateCmd := NewRootCmd()
	updateCmd.SetArgs([]string{"task", "update", "BIT-1"})
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("task update Execute() returned error: %v", err)
	}

	after, err := loadTask("BIT-1")
	if err != nil {
		t.Fatalf("loadTask(%q) returned error: %v", "BIT-1", err)
	}
	if *after != *before {
		t.Errorf("task changed with no flags: before %+v, after %+v", before, after)
	}
}

func TestTaskUpdateCmd_ErrorsOnUnknownID(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	updateCmd := NewRootCmd()
	updateCmd.SetArgs([]string{"task", "update", "BIT-99", "--title", "X"})
	if err := updateCmd.Execute(); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown ID")
	}
}
