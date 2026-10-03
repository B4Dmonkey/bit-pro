package task_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestTaskCreateCmd(t *testing.T) {
	t.Run("writes first task", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Set up init wizard", "Add flags for prefix capture.")

		got, err := projectStore(t).Load(trackID)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		got.Project, got.CreatedAt, got.UpdatedAt = "", time.Time{}, time.Time{}

		want := task.Task{
			ID:     trackID,
			Title:  "Set up init wizard",
			Status: statusTodo,
			Body:   "Add flags for prefix capture.",
		}
		if !reflect.DeepEqual(*got, want) {
			t.Errorf("task = %+v, want %+v", *got, want)
		}
	})

	t.Run("echoes minted id", func(t *testing.T) {
		initProject(t, "BIT")

		out := mustRun(t, "task", "create", "First track", "-d", "...")

		if out != "BIT-1\n" {
			t.Errorf("task create stdout = %q, want %q", out, "BIT-1\n")
		}
	})

	t.Run("echoes second track id", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "First", "...")

		out := mustRun(t, "task", "create", "Second", "-d", "...")

		if out != "BIT-2\n" {
			t.Errorf("task create stdout = %q, want %q", out, "BIT-2\n")
		}
	})

	t.Run("echoes child id", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")

		out := mustRun(t, "task", "create", "A bar", "-d", "...", "--parent", trackID)

		if out != "BIT-1.1\n" {
			t.Errorf("task create stdout = %q, want %q", out, "BIT-1.1\n")
		}
	})

	t.Run("assigns next id when tasks exist", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "First", "...")
		createTask(t, "Second", "...")

		got, err := projectStore(t).Load("BIT-2")
		if err != nil {
			t.Fatalf("loading BIT-2: %v", err)
		}

		if got.Title != "Second" {
			t.Errorf("BIT-2 title = %q, want %q", got.Title, "Second")
		}
	})

	t.Run("parent mints dotted id", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")

		mustRun(t, "task", "create", "A step", "-d", "...", "--parent", trackID)

		out := mustRun(t, "task", "read", firstBarID)

		want := firstBarID + "\ttodo\tA step\n"
		if !strings.HasPrefix(out, want) {
			t.Errorf("task read BIT-1.1 first line = %q, want prefix %q", out, want)
		}
	})

	t.Run("errors on missing parent", func(t *testing.T) {
		initProject(t, "BIT")

		if _, err := run(t, "task", "create", "Orphan", "-d", "...", "--parent", "BIT-99"); err == nil {
			t.Fatal("Execute() returned nil error, want non-nil for a missing parent")
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-99.1.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(tasks/BIT-99.1.json) error = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("second child increments", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")

		mustRun(t, "task", "create", "First step", "-d", "...", "--parent", trackID)
		mustRun(t, "task", "create", "Second step", "-d", "...", "--parent", trackID)
		mustRun(t, "task", "create", "Third step", "-d", "...", "--parent", trackID)

		out := mustRun(t, "task", "list")
		for _, id := range []string{firstBarID, secondBarID, thirdBarID} {
			if !strings.Contains(out, id) {
				t.Errorf("task list = %q, want it to contain %q", out, id)
			}
		}
	})

	t.Run("appends to reordered track", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", trackID)
		mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", trackID)
		mustRun(t, "task", "move", secondBarID, "--before", firstBarID)

		mustRun(t, "task", "create", "Third bar", "-d", "...", "--parent", trackID)

		track, err := projectStore(t).Load(trackID)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		want := []string{secondBarID, firstBarID, thirdBarID}
		if !slices.Equal(track.Order, want) {
			t.Errorf("BIT-1 order = %v, want %v", track.Order, want)
		}

		out := mustRun(t, "task", "list", "--parent", trackID)

		var ids []string
		for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
			ids = append(ids, strings.SplitN(line, "\t", 2)[0])
		}

		if len(ids) == 0 || ids[len(ids)-1] != thirdBarID {
			t.Errorf("parent list = %v, want it to end with BIT-1.3", ids)
		}
	})

	t.Run("after inserts mid plan", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", trackID)
		mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", trackID)

		out := mustRun(t, "task", "create", "Inserted", "-d", "...", "--parent", trackID, "--after", firstBarID)

		if out != thirdBarID+"\n" {
			t.Errorf("minted ID = %q, want %q", out, thirdBarID+"\n")
		}

		track, err := projectStore(t).Load(trackID)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		want := []string{firstBarID, thirdBarID, secondBarID}
		if !slices.Equal(track.Order, want) {
			t.Errorf("BIT-1 order = %v, want %v", track.Order, want)
		}

		listOut := mustRun(t, "task", "list", "--parent", trackID)

		var ids []string
		for line := range strings.SplitSeq(strings.TrimSpace(listOut), "\n") {
			ids = append(ids, strings.SplitN(line, "\t", 2)[0])
		}

		if !slices.Equal(ids, want) {
			t.Errorf("parent list = %v, want %v", ids, want)
		}
	})

	t.Run("after rejects unknown anchor", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", trackID)
		mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", trackID)

		_, err := run(t, "task", "create", "Inserted", "-d", "...", "--parent", trackID, "--after", "BIT-1.9")
		if err == nil {
			t.Fatal("Execute() returned nil error, want non-nil for an unknown anchor")
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.3.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(tasks/BIT-1.3.json) error = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("lowercase parent does not destroy an existing bar", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "First bar", "-d", "ORIGINAL BAR ONE", "--parent", trackID)

		out := mustRun(t, "task", "create", "sneaky", "-d", "...", "--parent", "bit-1")

		if out != "BIT-1.2\n" {
			t.Errorf("minted ID = %q, want %q", out, "BIT-1.2\n")
		}

		minted, err := os.ReadFile(filepath.Join(storeDir(t), "tasks", "BIT-1.2.json"))
		if err != nil {
			t.Fatalf("os.ReadFile(tasks/BIT-1.2.json) error = %v", err)
		}

		if !strings.Contains(string(minted), `"id": "BIT-1.2"`) {
			t.Errorf("BIT-1.2.json = %q, want it to contain %q", minted, `"id": "BIT-1.2"`)
		}

		survivor, err := os.ReadFile(filepath.Join(storeDir(t), "tasks", "BIT-1.1.md"))
		if err != nil {
			t.Fatalf("os.ReadFile(tasks/BIT-1.1.md) error = %v", err)
		}

		if !strings.Contains(string(survivor), "ORIGINAL BAR ONE") {
			t.Errorf("BIT-1.1.md = %q, want it to still contain %q", survivor, "ORIGINAL BAR ONE")
		}

		record, err := os.ReadFile(filepath.Join(storeDir(t), "tasks", "BIT-1.1.json"))
		if err != nil {
			t.Fatalf("os.ReadFile(tasks/BIT-1.1.json) error = %v", err)
		}

		if !strings.Contains(string(record), `"title": "First bar"`) {
			t.Errorf("BIT-1.1.json = %q, want it to still contain %q", record, `"title": "First bar"`)
		}

		records, err := filepath.Glob(filepath.Join(storeDir(t), "tasks", "*.json"))
		if err != nil {
			t.Fatalf("filepath.Glob(tasks/*.json) error = %v", err)
		}

		if len(records) != 3 {
			t.Errorf("tasks records = %v, want 3", records)
		}
	})

	t.Run("errors without title", func(t *testing.T) {
		initProject(t, "BIT")

		if _, err := run(t, "task", "create"); err == nil {
			t.Fatal("Execute() returned nil error, want non-nil for missing title argument")
		}
	})

	t.Run("errors without config", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")
		t.Chdir(t.TempDir())

		if _, err := run(t, "task", "create", "Foo"); err == nil {
			t.Fatal("Execute() returned nil error, want non-nil when the folder is unregistered")
		}

		if _, err := os.Stat(".bit/tasks"); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.bit/tasks) error = %v, want fs.ErrNotExist", err)
		}
	})
}
