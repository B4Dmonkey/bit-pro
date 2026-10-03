package task_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestTaskDeleteCmd(t *testing.T) {
	t.Run("removes file with yes flag", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Throwaway", "Delete me.")

		mustRun(t, "task", "delete", trackID, "--yes")

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(tasks/BIT-1.json) error = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("relocates instead of destroying", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Recoverable", "Deleted, not destroyed.")
		mustRun(t, "task", "update", trackID, "-s", "done")

		mustRun(t, "task", "delete", trackID, "--yes")

		if _, err := os.Stat(filepath.Join(storeDir(t), "archive", "tasks", "BIT-1.json")); err != nil {
			t.Errorf("os.Stat(archive/tasks/BIT-1.json) error = %v, want the task recoverable", err)
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(tasks/BIT-1.json) error = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("force deletes unfinished", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "A track with an unfinished bar.")
		mustRun(t, "task", "create", "Bar", "--parent", trackID, "--description", "Still todo.")

		mustRun(t, "task", "delete", trackID, "--yes", "--force")

		for _, id := range []string{trackID, firstBarID} {
			if _, err := os.Stat(filepath.Join(storeDir(t), "archive", "tasks", id+".json")); err != nil {
				t.Errorf("os.Stat(archive/tasks/%s.json) error = %v, want it relocated", id, err)
			}

			if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", id+".json")); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("os.Stat(tasks/%s.json) error = %v, want fs.ErrNotExist", id, err)
			}
		}
	})

	t.Run("refuses unfinished without force", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "A track with an unfinished bar.")
		mustRun(t, "task", "create", "Bar", "--parent", trackID, "--description", "Still todo.")

		_, err := run(t, "task", "delete", trackID, "--yes")

		var unfinished *task.UnfinishedBarsError
		if !errors.As(err, &unfinished) {
			t.Fatalf("Execute() error = %v, want *task.UnfinishedBarsError", err)
		}

		if !slices.Contains(unfinished.Bars, firstBarID) {
			t.Errorf("UnfinishedBarsError.Bars = %v, want it to contain BIT-1.1", unfinished.Bars)
		}

		for _, id := range []string{trackID, firstBarID} {
			if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", id+".json")); err != nil {
				t.Errorf("os.Stat(tasks/%s.json) error = %v, want the task to survive", id, err)
			}
		}
	})

	t.Run("prompts for confirmation", func(t *testing.T) {
		tests := []struct {
			name       string
			input      string
			wantExists bool
		}{
			{name: "y confirms", input: "y\n", wantExists: false},
			{name: "yes confirms", input: "yes\n", wantExists: false},
			{name: "uppercase Y confirms", input: "Y\n", wantExists: false},
			{name: "n declines", input: "n\n", wantExists: true},
			{name: "bare newline declines", input: "\n", wantExists: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				initProject(t, "BIT")
				createTask(t, "X", "...")

				if _, err := runWithStdin(t, tt.input, "task", "delete", trackID); err != nil {
					t.Fatalf("Execute() returned error: %v", err)
				}

				_, statErr := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.json"))

				exists := statErr == nil
				if exists != tt.wantExists {
					t.Errorf("file exists = %v, want %v (stat err: %v)", exists, tt.wantExists, statErr)
				}
			})
		}
	})

	t.Run("keeps task when confirmation unreadable", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "X", "...")

		if _, err := runWithStdin(t, "", "task", "delete", trackID); err == nil {
			t.Fatal("Execute() returned nil error, want non-nil when stdin is at EOF")
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.json")); err != nil {
			t.Errorf("os.Stat(tasks/BIT-1.json) error = %v, want the task to survive", err)
		}
	})

	t.Run("errors on unknown id", func(t *testing.T) {
		initProject(t, "BIT")

		_, err := run(t, "task", "delete", "BIT-99", "--yes")

		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Execute() error = %v, want an error wrapping fs.ErrNotExist", err)
		}

		if !strings.Contains(err.Error(), "BIT-99") {
			t.Errorf("Execute() error = %q, want it to name the task ID", err)
		}
	})

	t.Run("contains path traversal id", func(t *testing.T) {
		initProject(t, "BIT")

		readme := filepath.Join(filepath.Dir(storeDir(t)), "README.json")
		if err := os.WriteFile(readme, []byte("# real project readme\n"), 0o600); err != nil {
			t.Fatalf("writing README fixture: %v", err)
		}

		_, err := run(t, "task", "delete", "../../README", "--yes")

		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Execute() error = %v, want an error wrapping fs.ErrNotExist", err)
		}

		got, err := os.ReadFile(readme)
		if err != nil {
			t.Fatalf("reading README fixture after delete attempt: %v", err)
		}

		if string(got) != "# real project readme\n" {
			t.Errorf("README fixture = %q, want unchanged", got)
		}
	})
}
