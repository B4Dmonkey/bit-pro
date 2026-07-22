package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskDeleteCmd_RemovesFileWithYesFlag(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Throwaway", "Delete me.")

	mustRun(t, "task", "delete", "BIT-1", "--yes")

	if _, err := os.Stat(".bit/tasks/BIT-1.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/tasks/BIT-1.md) error = %v, want fs.ErrNotExist", err)
	}
}

func TestTaskDeleteCmd_RelocatesInsteadOfDestroying(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Recoverable", "Deleted, not destroyed.")
	mustRun(t, "task", "update", "BIT-1", "-s", "done")

	mustRun(t, "task", "delete", "BIT-1", "--yes")

	if _, err := os.Stat(".bit/archive/BIT-1.md"); err != nil {
		t.Errorf("os.Stat(.bit/archive/BIT-1.md) error = %v, want the task recoverable", err)
	}
	if _, err := os.Stat(".bit/tasks/BIT-1.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/tasks/BIT-1.md) error = %v, want fs.ErrNotExist", err)
	}
}

func TestTaskDeleteCmd_PromptsForConfirmation(t *testing.T) {
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

			if _, err := runWithStdin(t, tt.input, "task", "delete", "BIT-1"); err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			_, statErr := os.Stat(".bit/tasks/BIT-1.md")
			exists := statErr == nil
			if exists != tt.wantExists {
				t.Errorf("file exists = %v, want %v (stat err: %v)", exists, tt.wantExists, statErr)
			}
		})
	}
}

func TestTaskDeleteCmd_KeepsTaskWhenConfirmationUnreadable(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "X", "...")

	if _, err := runWithStdin(t, "", "task", "delete", "BIT-1"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil when stdin is at EOF")
	}

	if _, err := os.Stat(".bit/tasks/BIT-1.md"); err != nil {
		t.Errorf("os.Stat(.bit/tasks/BIT-1.md) error = %v, want the task to survive", err)
	}
}

func TestTaskDeleteCmd_ErrorsOnUnknownID(t *testing.T) {
	initProject(t, "BIT")

	_, err := run(t, "task", "delete", "BIT-99", "--yes")

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Execute() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), "BIT-99") {
		t.Errorf("Execute() error = %q, want it to name the task ID", err)
	}
}

func TestTaskDeleteCmd_ContainsPathTraversalID(t *testing.T) {
	dir := initProject(t, "BIT")
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# real project readme\n"), 0o644); err != nil {
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
}
