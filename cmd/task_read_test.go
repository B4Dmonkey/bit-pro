package cmd

import (
	"bytes"
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
