package task_test

import (
	"slices"
	"strings"
	"testing"
)

func TestTaskMoveCmd(t *testing.T) {
	t.Run("reorders parent list", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		createBar(t, trackID, "First bar", "...")
		createBar(t, trackID, "Second bar", "...")

		mustRun(t, taskCmdUse, "move", secondBarID, "--before", firstBarID)

		out := mustRun(t, taskCmdUse, "list", "--parent", trackID)

		var ids []string
		for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
			ids = append(ids, strings.SplitN(line, "\t", 2)[0])
		}

		want := []string{secondBarID, firstBarID}
		if !slices.Equal(ids, want) {
			t.Errorf("parent list order = %v, want %v", ids, want)
		}
	})

	t.Run("rejects bad flags", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
		}{
			{
				name: "both before and after",
				args: []string{taskCmdUse, "move", secondBarID, "--before", firstBarID, "--after", firstBarID},
			},
			{name: "neither before nor after", args: []string{taskCmdUse, "move", secondBarID}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				initProject(t, "BIT")
				createTask(t, "Track", "...")
				createBar(t, trackID, "First bar", "...")
				createBar(t, trackID, "Second bar", "...")

				if _, err := run(t, tt.args...); err == nil {
					t.Fatalf("bit %s returned nil error, want non-nil", strings.Join(tt.args, " "))
				}
			})
		}
	})
}
