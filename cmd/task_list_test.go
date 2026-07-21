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

	want := "BIT-2\ttodo\tSecond\t\nBIT-1\ttodo\tFirst\t\n"
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

func TestTaskListCmd_FiltersToParentBars(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "One", "...")
	createTask(t, "Two", "...")
	mustRun(t, "task", "create", "One.1", "-d", "...", "--parent", "BIT-1")
	mustRun(t, "task", "create", "One.2", "-d", "...", "--parent", "BIT-1")
	mustRun(t, "task", "create", "Two.1", "-d", "...", "--parent", "BIT-2")

	out := mustRun(t, "task", "list", "--parent", "BIT-1")

	want := "BIT-1.1\ttodo\tOne.1\t\nBIT-1.2\ttodo\tOne.2\t\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestTaskListCmd_ParentWithNoBars(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Lonely", "...")

	out := mustRun(t, "task", "list", "--parent", "BIT-9")

	if out != "" {
		t.Errorf("output = %q, want empty output for a parent with no bars", out)
	}
}

func TestTaskListCmd_ShowsPhaseOnBars(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "Bar one", "-d", "...", "--parent", "BIT-1",
		"--phase", "1", "--phase-label", "First")
	mustRun(t, "task", "create", "Bar two", "-d", "...", "--parent", "BIT-1",
		"--phase", "2", "--phase-label", "Second")

	out := mustRun(t, "task", "list")

	want := "BIT-1\ttodo\tTrack\t\n" +
		"BIT-1.1\ttodo\tBar one\tphase 1 — First\n" +
		"BIT-1.2\ttodo\tBar two\tphase 2 — Second\n"
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
