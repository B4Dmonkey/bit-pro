package task_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTaskCompleteCmd(t *testing.T) {
	t.Run("files track and bars under completed", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Separate completed work", "## Why\n\nFinished work needs its own home.\n")
		mustRun(t, "task", "create", "First bar", "--parent", trackID, "--description", "Done work.")
		mustRun(t, "task", "create", "Second bar", "--parent", trackID, "--description", "Done work.")

		for _, id := range []string{firstBarID, secondBarID, trackID} {
			mustRun(t, "task", "update", id, "-s", "done")
		}

		mustRun(t, "task", "complete", trackID)

		for _, id := range []string{trackID, firstBarID, secondBarID} {
			if _, err := os.Stat(filepath.Join(storeDir(t), "completed", id+".md")); err != nil {
				t.Errorf("os.Stat(completed/%s.md) error = %v, want it filed as completed", id, err)
			}

			if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", id+".md")); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("os.Stat(tasks/%s.md) error = %v, want fs.ErrNotExist", id, err)
			}
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "archive")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(archive) error = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("lowercase track still hits the guard", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Guard the unfinished bars", "## Why\n\nThe guard follows identity, not spelling.\n")
		mustRun(t, "task", "create", "Unfinished bar", "--parent", trackID, "--description", "Still todo.")

		_, err := run(t, "task", "complete", "bit-1")
		if err == nil {
			t.Fatal("bp task complete bit-1 error = nil, want the unfinished-bars guard to fire")
		}

		if !strings.Contains(err.Error(), "unfinished bars BIT-1.1") {
			t.Errorf("bp task complete bit-1 error = %q, want it to name unfinished bars BIT-1.1", err)
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.md")); err != nil {
			t.Errorf("os.Stat(tasks/BIT-1.md) error = %v, want the track left where it was", err)
		}

		for _, id := range []string{trackID, "bit-1"} {
			if _, err := os.Stat(filepath.Join(storeDir(t), "completed", id+".md")); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("os.Stat(completed/%s.md) error = %v, want fs.ErrNotExist", id, err)
			}
		}
	})

	t.Run("lowercase track files uppercase filenames", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "File work under one name", "## Why\n\nThe write path follows identity, not spelling.\n")
		mustRun(t, "task", "create", "First bar", "--parent", trackID, "--description", "Done work.")
		mustRun(t, "task", "create", "Second bar", "--parent", trackID, "--description", "Done work.")

		for _, id := range []string{firstBarID, secondBarID, trackID} {
			mustRun(t, "task", "update", id, "-s", "done")
		}

		mustRun(t, "task", "complete", "bit-1")

		entries, err := os.ReadDir(filepath.Join(storeDir(t), "completed"))
		if err != nil {
			t.Fatalf("os.ReadDir(completed) error = %v", err)
		}

		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}

		slices.Sort(names)

		want := []string{"BIT-1.1.md", "BIT-1.2.md", "BIT-1.md"}
		if !slices.Equal(names, want) {
			t.Errorf("os.ReadDir(completed) names = %v, want %v", names, want)
		}

		left, err := os.ReadDir(filepath.Join(storeDir(t), "tasks"))
		if err != nil {
			t.Fatalf("os.ReadDir(tasks) error = %v", err)
		}

		if len(left) != 0 {
			t.Errorf("os.ReadDir(tasks) left %d entries behind, want none", len(left))
		}
	})

	t.Run("hand edited lowercase id still hits the guard", func(t *testing.T) {
		initProject(t, "BIT")
		writeRawTask(t, filepath.Join(storeDir(t), "tasks", "BIT-1.md"), "bit-1", "Guard the unfinished bars", statusTodo)
		writeRawTask(t, filepath.Join(storeDir(t), "tasks", "BIT-1.1.md"), "bit-1.1", "Unfinished bar", statusTodo)

		_, err := run(t, "task", "complete", trackID)
		if err == nil {
			t.Fatal("bp task complete BIT-1 error = nil, want the unfinished-bars guard to fire")
		}

		if !strings.Contains(err.Error(), "unfinished bars BIT-1.1") {
			t.Errorf("bp task complete BIT-1 error = %q, want it to name unfinished bars BIT-1.1", err)
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "completed")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(completed) error = %v, want fs.ErrNotExist", err)
		}

		for _, id := range []string{trackID, firstBarID} {
			if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", id+".md")); err != nil {
				t.Errorf("os.Stat(tasks/%s.md) error = %v, want the task left where it was", id, err)
			}
		}
	})

	t.Run("replaces archive", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Done thing", "Finished work.")
		mustRun(t, "task", "update", trackID, "-s", "done")

		out, _ := run(t, "task", "archive", trackID)

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.md")); err != nil {
			t.Errorf("os.Stat(tasks/BIT-1.md) error = %v, want the track left where it was", err)
		}

		if strings.Contains(out, "archive") {
			t.Errorf("task archive output = %q, want no archive subcommand in it", out)
		}
	})
}

func writeRawTask(t *testing.T, path, id, title, status string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}

	body := "---\nid: " + id + "\ntitle: " + title + "\nstatus: " + status + "\n---\nHand-written.\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) error = %v", path, err)
	}
}
