package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestTaskListCmd_ShowsAllTasks(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createFirst := NewRootCmd()
	createFirst.SetArgs([]string{"task", "create", "First", "--description", "..."})
	if err := createFirst.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	createSecond := NewRootCmd()
	createSecond.SetArgs([]string{"task", "create", "Second", "--description", "..."})
	if err := createSecond.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	listCmd := NewRootCmd()
	buf := &bytes.Buffer{}
	listCmd.SetOut(buf)
	listCmd.SetArgs([]string{"task", "list"})
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("task list Execute() returned error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"BIT-1", "First", "BIT-2", "Second"} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want to contain %q", out, want)
		}
	}

	if strings.Index(out, "BIT-1") > strings.Index(out, "BIT-2") {
		t.Errorf("output = %q, want BIT-1 before BIT-2", out)
	}
}

func TestTaskListCmd_EmptyWhenNoTasks(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	listCmd := NewRootCmd()
	buf := &bytes.Buffer{}
	listCmd.SetOut(buf)
	listCmd.SetArgs([]string{"task", "list"})
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("task list Execute() returned error: %v", err)
	}
}
