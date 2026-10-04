package task_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestTaskReadCmd(t *testing.T) {
	t.Run("shows full task", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Full details", "Line one.\nLine two.")

		out := mustRun(t, "task", "read", "BIT-1")

		want := "BIT-1\ttodo\tFull details\n\nLine one.\nLine two."
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("body only", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Full details", "Line one.\nLine two.")

		out := mustRun(t, "task", "read", "BIT-1", "--body")

		if out != "Line one.\nLine two." {
			t.Errorf("output = %q, want %q", out, "Line one.\nLine two.")
		}
	})

	t.Run("body only empty", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "No body", "")

		out := mustRun(t, "task", "read", "BIT-1", "--body")

		if out != "" {
			t.Errorf("output = %q, want %q", out, "")
		}
	})

	t.Run("shows phase", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		createWith(t, task.CreateParams{
			Title: "List cmd", Parent: trackID, Phase: 2, PhaseLabel: "List & read",
		})

		out := mustRun(t, "task", "read", "BIT-1.1")

		firstLine := strings.SplitN(out, "\n", 2)[0]

		want := "BIT-1.1\ttodo\tList cmd\tphase 2 — List & read"
		if firstLine != want {
			t.Errorf("first line = %q, want %q", firstLine, want)
		}
	})

	t.Run("omits phase when absent", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Title", "Body")

		out := mustRun(t, "task", "read", "BIT-1")

		want := "BIT-1\ttodo\tTitle\n\nBody"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("errors on unknown id", func(t *testing.T) {
		initProject(t, "BIT")

		if _, err := run(t, "task", "read", "BIT-99"); err == nil {
			t.Fatal("Execute() returned nil error, want non-nil for unknown task ID")
		}
	})

	t.Run("contains path traversal id", func(t *testing.T) {
		initProject(t, "BIT")

		readme := filepath.Join(filepath.Dir(storeDir(t)), "README.json")
		if err := os.WriteFile(readme, []byte("# real project readme\n"), 0o600); err != nil {
			t.Fatalf("writing README fixture: %v", err)
		}

		out, err := run(t, "task", "read", "../../README")

		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Execute() error = %v, want an error wrapping fs.ErrNotExist", err)
		}

		if strings.Contains(out, "real project readme") {
			t.Errorf("output = %q, must not contain the escaped file's content", out)
		}
	})
}
