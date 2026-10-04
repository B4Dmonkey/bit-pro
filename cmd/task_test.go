package cmd

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/spf13/cobra"
)

func taskSubcommand(t *testing.T) *cobra.Command {
	t.Helper()

	root := newRootCmd(func(context.Context, string, ...string) error { return nil })

	for _, c := range root.Commands() {
		if c.Name() == "task" {
			return c
		}
	}

	t.Fatal(`newRootCmd() has no "task" subcommand, want the cmd/task package wired in`)

	return nil
}

func TestTaskCmd(t *testing.T) {
	t.Run("subcommands are wired under root", func(t *testing.T) {
		var got []string
		for _, c := range taskSubcommand(t).Commands() {
			got = append(got, c.Name())
		}

		slices.Sort(got)

		want := []string{"delete", "list", "move", "read", "update"}
		if !slices.Equal(got, want) {
			t.Errorf("bp task subcommands = %v, want %v", got, want)
		}
	})

	t.Run("lifecycle runs through the root command", func(t *testing.T) {
		initProject(t, testPrefix)

		id := createTask(t, "Wired track", "Body.")

		mustRun(t, "task", updateCmd, id, "-s", "done")

		if out := mustRun(t, "task", "read", id); !strings.Contains(out, "done") {
			t.Errorf("bp task read %s = %q, want it to report the done status", id, out)
		}

		if err := projectStore(t).Complete(id); err != nil {
			t.Fatalf("Complete(%s) error = %v", id, err)
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "completed", id+".json")); err != nil {
			t.Errorf("os.Stat(completed/%s.json) error = %v, want the track filed as completed", id, err)
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", id+".json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(tasks/%s.json) error = %v, want fs.ErrNotExist", id, err)
		}
	})

	t.Run("works from a subfolder against the central store", func(t *testing.T) {
		dir := initProject(t, testPrefix)
		createTask(t, "Track", "...")

		sub := filepath.Join(dir, "src", "pkg")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) returned error: %v", sub, err)
		}

		t.Chdir(sub)

		if out := mustRun(t, "task", "list"); !strings.Contains(out, "BIT-1") {
			t.Errorf("bp task list from %s = %q, want it to contain BIT-1", sub, out)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() returned error: %v", err)
		}

		central := filepath.Join(home, ".local", "share", "bit", testPrefix, "tasks", "BIT-1.json")
		if _, err := os.Stat(central); err != nil {
			t.Errorf("os.Stat(%s) error = %v, want the track in the central store", central, err)
		}

		if _, err := os.Stat(filepath.Join(dir, ".bit", "tasks")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(%s/.bit/tasks) error = %v, want fs.ErrNotExist", dir, err)
		}
	})

	t.Run("outside worktree uses the project", func(t *testing.T) {
		initProject(t, testPrefix)
		createTask(t, "Track", "...")

		out := mustRun(t, "task", "list")

		if !strings.Contains(out, "BIT-1") {
			t.Errorf("output = %q, want output to contain BIT-1 from the project's store", out)
		}
	})

	t.Run("inside claude worktree resolves to main checkout", func(t *testing.T) {
		root := initProject(t, testPrefix)
		createTask(t, "Track", "...")

		worktree := filepath.Join(root, ".claude", "worktrees", "hazy-pondering-star")
		if err := os.MkdirAll(worktree, 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) returned error: %v", worktree, err)
		}

		t.Chdir(worktree)

		out := mustRun(t, "task", "list")

		if !strings.Contains(out, "BIT-1") {
			t.Errorf("output = %q, want output to contain BIT-1 from the main checkout's project", out)
		}
	})

	t.Run("nested worktree resolves to outermost checkout", func(t *testing.T) {
		root := initProject(t, testPrefix)
		createTask(t, "Track", "...")

		nested := filepath.Join(root, ".claude", "worktrees", "outer", ".claude", "worktrees", "inner")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) returned error: %v", nested, err)
		}

		t.Chdir(nested)

		out := mustRun(t, "task", "list")

		if !strings.Contains(out, "BIT-1") {
			t.Errorf("output = %q, want output to contain BIT-1 from the outermost checkout's project", out)
		}
	})

	unregistered := []struct {
		name   string
		bitDir bool
		want   error
	}{
		{name: "unregistered folder says run bp add", want: project.ErrNotRegistered},
		{name: "unregistered folder with bit says run bp migrate", bitDir: true, want: project.ErrNeedsMigrate},
	}

	for _, tt := range unregistered {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", "")
			t.Chdir(t.TempDir())

			if tt.bitDir {
				if err := os.Mkdir(".bit", 0o755); err != nil {
					t.Fatalf("os.Mkdir(.bit) returned error: %v", err)
				}
			}

			if _, err := run(t, "task", "list"); !errors.Is(err, tt.want) {
				t.Errorf("bp task list error = %v, want %v", err, tt.want)
			}
		})
	}
}
