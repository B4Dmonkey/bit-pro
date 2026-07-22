package cmd

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

func TestTaskArchiveCmd_RelocatesTask(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Done thing", "Finished work.")
	mustRun(t, "task", "update", "BIT-1", "-s", "done")

	mustRun(t, "task", "archive", "BIT-1")

	if _, err := os.Stat(".bit/archive/BIT-1.md"); err != nil {
		t.Errorf("os.Stat(.bit/archive/BIT-1.md) error = %v, want the task relocated", err)
	}
	if _, err := os.Stat(".bit/tasks/BIT-1.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/tasks/BIT-1.md) error = %v, want fs.ErrNotExist", err)
	}
}

func TestTaskArchiveCmd_ForceArchivesUnfinished(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "A track with an unfinished bar.")
	mustRun(t, "task", "create", "Bar", "--parent", "BIT-1", "--description", "Still todo.")

	mustRun(t, "task", "archive", "BIT-1", "--force")

	for _, id := range []string{"BIT-1", "BIT-1.1"} {
		if _, err := os.Stat(".bit/archive/" + id + ".md"); err != nil {
			t.Errorf("os.Stat(.bit/archive/%s.md) error = %v, want it relocated", id, err)
		}
		if _, err := os.Stat(".bit/tasks/" + id + ".md"); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.bit/tasks/%s.md) error = %v, want fs.ErrNotExist", id, err)
		}
	}
}
