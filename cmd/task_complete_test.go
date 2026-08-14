package cmd

import (
	"errors"
	"io/fs"
	"os"
	"slices"
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

func TestTaskCompleteCmd_LowercaseTrackStillHitsTheGuard(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Guard the unfinished bars", "## Why\n\nThe guard follows identity, not spelling.\n")
	mustRun(t, "task", "create", "Unfinished bar", "--parent", "BIT-1", "--description", "Still todo.")

	_, err := run(t, "task", "complete", "bit-1")

	if err == nil {
		t.Fatal("bp task complete bit-1 error = nil, want the unfinished-bars guard to fire")
	}
	if !strings.Contains(err.Error(), "unfinished bars BIT-1.1") {
		t.Errorf("bp task complete bit-1 error = %q, want it to name unfinished bars BIT-1.1", err)
	}
	if _, err := os.Stat(".bit/tasks/BIT-1.md"); err != nil {
		t.Errorf("os.Stat(.bit/tasks/BIT-1.md) error = %v, want the track left where it was", err)
	}
	for _, id := range []string{"BIT-1", "bit-1"} {
		if _, err := os.Stat(".bit/completed/" + id + ".md"); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.bit/completed/%s.md) error = %v, want fs.ErrNotExist", id, err)
		}
	}
}

func TestTaskCompleteCmd_LowercaseTrackFilesUppercaseFilenames(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "File work under one name", "## Why\n\nThe write path follows identity, not spelling.\n")
	mustRun(t, "task", "create", "First bar", "--parent", "BIT-1", "--description", "Done work.")
	mustRun(t, "task", "create", "Second bar", "--parent", "BIT-1", "--description", "Done work.")
	for _, id := range []string{"BIT-1.1", "BIT-1.2", "BIT-1"} {
		mustRun(t, "task", "update", id, "-s", "done")
	}

	mustRun(t, "task", "complete", "bit-1")

	entries, err := os.ReadDir(".bit/completed")
	if err != nil {
		t.Fatalf("os.ReadDir(.bit/completed) error = %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	slices.Sort(names)
	want := []string{"BIT-1.1.md", "BIT-1.2.md", "BIT-1.md"}
	if !slices.Equal(names, want) {
		t.Errorf("os.ReadDir(.bit/completed) names = %v, want %v", names, want)
	}

	left, err := os.ReadDir(".bit/tasks")
	if err != nil {
		t.Fatalf("os.ReadDir(.bit/tasks) error = %v", err)
	}
	if len(left) != 0 {
		t.Errorf("os.ReadDir(.bit/tasks) left %d entries behind, want none", len(left))
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
