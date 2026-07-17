package cmd

import "testing"

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

func TestTaskListCmd_EmptyWhenNoTasks(t *testing.T) {
	initProject(t, "BIT")

	out := mustRun(t, "task", "list")

	if out != "" {
		t.Errorf("output = %q, want empty output when no tasks exist", out)
	}
}
