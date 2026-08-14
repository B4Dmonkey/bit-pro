package cmd

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestTaskUpdateCmd(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want task.Task
	}{
		{
			name: "no flags is a no-op",
			args: []string{"task", "update", "BIT-1"},
			want: task.Task{ID: "BIT-1", Title: "Old title", Status: "todo", Body: "Old body."},
		},
		{
			name: "title only leaves status and body alone",
			args: []string{"task", "update", "BIT-1", "--title", "New title"},
			want: task.Task{ID: "BIT-1", Title: "New title", Status: "todo", Body: "Old body."},
		},
		{
			name: "description and status together",
			args: []string{"task", "update", "BIT-1", "--description", "New body.", "--status", "doing"},
			want: task.Task{ID: "BIT-1", Title: "Old title", Status: "doing", Body: "New body."},
		},
		{
			name: "explicitly empty title is applied",
			args: []string{"task", "update", "BIT-1", "--title", ""},
			want: task.Task{ID: "BIT-1", Title: "", Status: "todo", Body: "Old body."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initProject(t, "BIT")
			createTask(t, "Old title", "Old body.")

			mustRun(t, tt.args...)

			got, err := task.New(".bit").Load("BIT-1")
			if err != nil {
				t.Fatalf("loading BIT-1: %v", err)
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("task = %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func TestTaskUpdateCmd_ChangesPhase(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "Bar", "-d", "...", "--parent", "BIT-1",
		"--phase", "2", "--phase-label", "List & read")

	mustRun(t, "task", "update", "BIT-1.1", "--phase", "3", "--phase-label", "Update")

	out := mustRun(t, "task", "read", "BIT-1.1")

	firstLine := strings.SplitN(out, "\n", 2)[0]
	want := "BIT-1.1\ttodo\tBar\tphase 3 — Update"
	if firstLine != want {
		t.Errorf("first line = %q, want %q", firstLine, want)
	}
}

func TestTaskUpdateCmd_RewritesACorruptIDToCanonicalCase(t *testing.T) {
	initProject(t, "BIT")
	writeRawTask(t, ".bit/tasks/BIT-1.md", "bit-1", "Corrupt frontmatter", "todo")

	mustRun(t, "task", "update", "BIT-1", "-s", "doing")

	data, err := os.ReadFile(".bit/tasks/BIT-1.md")
	if err != nil {
		t.Fatalf("os.ReadFile(.bit/tasks/BIT-1.md) error = %v", err)
	}
	got := string(data)
	for _, want := range []string{"id: BIT-1\n", "status: doing\n"} {
		if !strings.Contains(got, want) {
			t.Errorf(".bit/tasks/BIT-1.md = %q, want it to contain %q", got, want)
		}
	}
	entries, err := os.ReadDir(".bit/tasks")
	if err != nil {
		t.Fatalf("os.ReadDir(.bit/tasks) error = %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("os.ReadDir(.bit/tasks) returned %d entries, want 1", len(entries))
	}
}

func TestTaskUpdateCmd_ErrorsOnUnknownID(t *testing.T) {
	initProject(t, "BIT")

	if _, err := run(t, "task", "update", "BIT-99", "--title", "X"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown ID")
	}
}
