package task_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

const oldTaskBody = "Old body."

func TestTaskUpdateCmd(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want task.Task
	}{
		{
			name: "no flags is a no-op",
			args: []string{taskCmdUse, updateCmd, trackID},
			want: task.Task{ID: trackID, Title: "Old title", Status: statusTodo, Body: oldTaskBody},
		},
		{
			name: "title only leaves status and body alone",
			args: []string{taskCmdUse, updateCmd, trackID, "--title", "New title"},
			want: task.Task{ID: trackID, Title: "New title", Status: statusTodo, Body: oldTaskBody},
		},
		{
			name: "description and status together",
			args: []string{taskCmdUse, updateCmd, trackID, "--description", "New body.", "--status", statusDoing},
			want: task.Task{ID: trackID, Title: "Old title", Status: statusDoing, Body: "New body."},
		},
		{
			name: "explicitly empty title is applied",
			args: []string{taskCmdUse, updateCmd, trackID, "--title", ""},
			want: task.Task{ID: trackID, Title: "", Status: statusTodo, Body: oldTaskBody},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initProject(t, "BIT")
			createTask(t, "Old title", oldTaskBody)

			mustRun(t, tt.args...)

			got, err := projectStore(t).Load(trackID)
			if err != nil {
				t.Fatalf("loading BIT-1: %v", err)
			}

			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("task = %+v, want %+v", *got, tt.want)
			}
		})
	}

	t.Run("changes phase", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "Bar", "-d", "...", "--parent", trackID,
			"--phase", "2", "--phase-label", "List & read")

		mustRun(t, taskCmdUse, updateCmd, firstBarID, "--phase", "3", "--phase-label", "Update")

		out := mustRun(t, "task", "read", firstBarID)

		firstLine := strings.SplitN(out, "\n", 2)[0]

		want := firstBarID + "\ttodo\tBar\tphase 3 — Update"
		if firstLine != want {
			t.Errorf("first line = %q, want %q", firstLine, want)
		}
	})

	t.Run("rewrites a corrupt id to canonical case", func(t *testing.T) {
		initProject(t, "BIT")
		writeRawTask(t, filepath.Join(storeDir(t), "tasks", "BIT-1.md"), "bit-1", "Corrupt frontmatter", statusTodo)

		mustRun(t, taskCmdUse, updateCmd, trackID, "-s", statusDoing)

		data, err := os.ReadFile(filepath.Join(storeDir(t), "tasks", "BIT-1.md"))
		if err != nil {
			t.Fatalf("os.ReadFile(tasks/BIT-1.md) error = %v", err)
		}

		got := string(data)
		for _, want := range []string{"id: BIT-1\n", "status: doing\n"} {
			if !strings.Contains(got, want) {
				t.Errorf("tasks/BIT-1.md = %q, want it to contain %q", got, want)
			}
		}

		entries, err := os.ReadDir(filepath.Join(storeDir(t), "tasks"))
		if err != nil {
			t.Fatalf("os.ReadDir(tasks) error = %v", err)
		}

		if len(entries) != 1 {
			t.Errorf("os.ReadDir(tasks) returned %d entries, want 1", len(entries))
		}
	})

	t.Run("errors on unknown id", func(t *testing.T) {
		initProject(t, "BIT")

		if _, err := run(t, taskCmdUse, updateCmd, "BIT-99", "--title", "X"); err == nil {
			t.Fatal("Execute() returned nil error, want non-nil for unknown ID")
		}
	})

	t.Run("revokes approval on title change", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Old title", "...")
		approve(t, trackID)

		mustRun(t, taskCmdUse, updateCmd, trackID, "--title", "New title")

		got, err := projectStore(t).Load(trackID)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		if got.Approved {
			t.Error("expected Approved = false after title change, got true")
		}
	})

	t.Run("no op preserves approval", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Old title", "...")
		approve(t, trackID)

		mustRun(t, taskCmdUse, updateCmd, trackID)

		got, err := projectStore(t).Load(trackID)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		if !got.Approved {
			t.Error("expected Approved = true after no-op update, got false")
		}
	})

	t.Run("revokes approval on body change", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Old title", oldTaskBody)
		approve(t, trackID)

		mustRun(t, taskCmdUse, updateCmd, trackID, "--description", "New body.")

		got, err := projectStore(t).Load(trackID)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		if got.Approved {
			t.Error("expected Approved = false after description change, got true")
		}
	})

	t.Run("forward status move preserves approval", func(t *testing.T) {
		tests := []struct {
			name string
			from string
			to   string
		}{
			{name: "todo to doing", to: statusDoing},
			{name: "doing to done", from: statusDoing, to: "done"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				initProject(t, "BIT")
				createTask(t, "Old title", oldTaskBody)

				if tt.from != "" {
					mustRun(t, taskCmdUse, updateCmd, trackID, "-s", tt.from)
				}

				approve(t, trackID)

				mustRun(t, taskCmdUse, updateCmd, trackID, "-s", tt.to)

				got, err := projectStore(t).Load(trackID)
				if err != nil {
					t.Fatalf("loading %s: %v", trackID, err)
				}

				if !got.Approved {
					t.Errorf("Approved = false after move to %s, want true", tt.to)
				}

				if got.Status != tt.to {
					t.Errorf("Status = %q, want %q", got.Status, tt.to)
				}
			})
		}
	})

	t.Run("status to todo revokes approval", func(t *testing.T) {
		tests := []struct {
			name string
			from string
			to   string
			want bool
		}{
			{name: "doing to todo", from: statusDoing, to: statusTodo, want: false},
			{name: "done to todo", from: statusDone, to: statusTodo, want: false},
			{name: "todo to todo", to: statusTodo, want: false},
			{name: "done to doing", from: statusDone, to: statusDoing, want: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				initProject(t, "BIT")
				createTask(t, "Old title", oldTaskBody)

				if tt.from != "" {
					mustRun(t, taskCmdUse, updateCmd, trackID, "-s", tt.from)
				}

				approve(t, trackID)

				mustRun(t, taskCmdUse, updateCmd, trackID, "-s", tt.to)

				got, err := projectStore(t).Load(trackID)
				if err != nil {
					t.Fatalf("loading %s: %v", trackID, err)
				}

				if got.Approved != tt.want {
					t.Errorf("Approved = %v after move to %s, want %v", got.Approved, tt.to, tt.want)
				}

				if got.Status != tt.to {
					t.Errorf("Status = %q, want %q", got.Status, tt.to)
				}
			})
		}
	})
}
