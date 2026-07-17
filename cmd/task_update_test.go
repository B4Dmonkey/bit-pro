package cmd

import (
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
			if *got != tt.want {
				t.Errorf("task = %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func TestTaskUpdateCmd_ErrorsOnUnknownID(t *testing.T) {
	initProject(t, "BIT")

	if _, err := run(t, "task", "update", "BIT-99", "--title", "X"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown ID")
	}
}
