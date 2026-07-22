package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestTaskMoveCmd_ReordersParentList(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", "BIT-1")
	mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", "BIT-1")

	mustRun(t, "task", "move", "BIT-1.2", "--before", "BIT-1.1")

	out := mustRun(t, "task", "list", "--parent", "BIT-1")

	var ids []string
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		ids = append(ids, strings.SplitN(line, "\t", 2)[0])
	}
	want := []string{"BIT-1.2", "BIT-1.1"}
	if !slices.Equal(ids, want) {
		t.Errorf("parent list order = %v, want %v", ids, want)
	}
}

func TestTaskMoveCmd_RejectsBadFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "both before and after", args: []string{"task", "move", "BIT-1.2", "--before", "BIT-1.1", "--after", "BIT-1.1"}},
		{name: "neither before nor after", args: []string{"task", "move", "BIT-1.2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initProject(t, "BIT")
			createTask(t, "Track", "...")
			mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", "BIT-1")
			mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", "BIT-1")

			if _, err := run(t, tt.args...); err == nil {
				t.Fatalf("bit %s returned nil error, want non-nil", strings.Join(tt.args, " "))
			}
		})
	}
}
