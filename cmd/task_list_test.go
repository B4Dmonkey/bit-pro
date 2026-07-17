package cmd

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestTaskListCmd_ShowsNewestFirst(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "First", "...")
	createTask(t, "Second", "...")

	out := mustRun(t, "task", "list")

	want := "BIT-2\ttodo\tSecond\nBIT-1\ttodo\tFirst\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestTaskListCmd_OrdersNumericallyNotLexically(t *testing.T) {
	initProject(t, "BIT")
	for i := 1; i <= 10; i++ {
		createTask(t, fmt.Sprintf("T%d", i), "...")
	}

	out := mustRun(t, "task", "list")

	ids := make([]string, 0, 10)
	for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
		ids = append(ids, strings.Split(line, "\t")[0])
	}

	want := []string{"BIT-10", "BIT-9", "BIT-8", "BIT-7", "BIT-6", "BIT-5", "BIT-4", "BIT-3", "BIT-2", "BIT-1"}
	if !slices.Equal(ids, want) {
		t.Errorf("ID order = %v, want %v", ids, want)
	}
}

func TestTaskListCmd_GroupsBarsUnderTheirTrack(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "One", "...")
	createTask(t, "Two", "...")
	mustRun(t, "task", "create", "One.1", "-d", "...", "--parent", "BIT-1")
	mustRun(t, "task", "create", "One.2", "-d", "...", "--parent", "BIT-1")
	mustRun(t, "task", "create", "Two.1", "-d", "...", "--parent", "BIT-2")

	out := mustRun(t, "task", "list")

	ids := make([]string, 0, 5)
	for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
		ids = append(ids, strings.Split(line, "\t")[0])
	}

	want := []string{"BIT-2", "BIT-2.1", "BIT-1", "BIT-1.1", "BIT-1.2"}
	if !slices.Equal(ids, want) {
		t.Errorf("ID order = %v, want %v", ids, want)
	}
}

func TestTaskListCmd_EmptyWhenNoTasks(t *testing.T) {
	initProject(t, "BIT")

	out := mustRun(t, "task", "list")

	if out != "" {
		t.Errorf("output = %q, want empty output when no tasks exist", out)
	}
}
