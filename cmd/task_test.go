package cmd

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func taskSubcommand(t *testing.T) *cobra.Command {
	t.Helper()

	root := newRootCmd(func(context.Context, string, ...string) error { return nil }, nothingLoaded)

	for _, c := range root.Commands() {
		if c.Name() == "task" {
			return c
		}
	}

	t.Fatal(`newRootCmd() has no "task" subcommand, want the cmd/task package wired in`)

	return nil
}

func TestTaskCmd_SubcommandsAreWiredUnderRoot(t *testing.T) {
	var got []string
	for _, c := range taskSubcommand(t).Commands() {
		got = append(got, c.Name())
	}

	slices.Sort(got)

	want := []string{"complete", "create", "delete", "list", "move", "read", "update"}
	if !slices.Equal(got, want) {
		t.Errorf("bp task subcommands = %v, want %v", got, want)
	}
}

func TestTaskCmd_LifecycleRunsThroughTheRootCommand(t *testing.T) {
	initProject(t, testPrefix)

	id := strings.TrimSpace(mustRun(t, "task", "create", "Wired track", "--description", "Body."))
	if id != "BIT-1" {
		t.Fatalf("bp task create printed %q, want BIT-1", id)
	}

	mustRun(t, "task", updateCmd, id, "-s", "done")

	if out := mustRun(t, "task", "read", id); !strings.Contains(out, "done") {
		t.Errorf("bp task read %s = %q, want it to report the done status", id, out)
	}

	mustRun(t, "task", "complete", id)

	if _, err := os.Stat(".bit/completed/" + id + ".md"); err != nil {
		t.Errorf("os.Stat(.bit/completed/%s.md) error = %v, want the track filed as completed", id, err)
	}

	if _, err := os.Stat(".bit/tasks/" + id + ".md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/tasks/%s.md) error = %v, want fs.ErrNotExist", id, err)
	}
}
