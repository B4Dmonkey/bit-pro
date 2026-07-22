package cmd

import (
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

func TestTaskUpdateCmd_ErrorsOnUnknownID(t *testing.T) {
	initProject(t, "BIT")

	if _, err := run(t, "task", "update", "BIT-99", "--title", "X"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown ID")
	}
}
