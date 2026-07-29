package cmd

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestTaskCompleteCmd_FilesTrackAndBarsUnderCompleted(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Separate completed work", "## Why\n\nFinished work needs its own home.\n")
	mustRun(t, "task", "create", "First bar", "--parent", "BIT-1", "--description", "Done work.")
	mustRun(t, "task", "create", "Second bar", "--parent", "BIT-1", "--description", "Done work.")
	for _, id := range []string{"BIT-1.1", "BIT-1.2", "BIT-1"} {
		mustRun(t, "task", "update", id, "-s", "done")
	}

	mustRun(t, "task", "complete", "BIT-1")

	for _, id := range []string{"BIT-1", "BIT-1.1", "BIT-1.2"} {
		if _, err := os.Stat(".bit/completed/" + id + ".md"); err != nil {
			t.Errorf("os.Stat(.bit/completed/%s.md) error = %v, want it filed as completed", id, err)
		}
		if _, err := os.Stat(".bit/tasks/" + id + ".md"); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.bit/tasks/%s.md) error = %v, want fs.ErrNotExist", id, err)
		}
	}
	if _, err := os.Stat(".bit/archive"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/archive) error = %v, want fs.ErrNotExist", err)
	}
}

func TestTaskCompleteCmd_ReplacesArchive(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Done thing", "Finished work.")
	mustRun(t, "task", "update", "BIT-1", "-s", "done")

	out, _ := run(t, "task", "archive", "BIT-1")

	if _, err := os.Stat(".bit/tasks/BIT-1.md"); err != nil {
		t.Errorf("os.Stat(.bit/tasks/BIT-1.md) error = %v, want the track left where it was", err)
	}
	if strings.Contains(out, "archive") {
		t.Errorf("task archive output = %q, want no archive subcommand in it", out)
	}
}
