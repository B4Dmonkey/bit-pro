package cmd

import (
	"testing"
)

func TestTaskUpdateCmd_ChangesTitleOnly(t *testing.T) {
	// Arrange
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

	// Act
	updateCmd := NewRootCmd()
	updateCmd.SetArgs([]string{"task", "update", "BIT-1", "--title", "New title"})
	err := updateCmd.Execute()

	// Assert
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
