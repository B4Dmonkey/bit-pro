package cmd

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestTaskCreateCmd_WritesFirstTask(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Set up init wizard", "Add flags for prefix capture.")

	got, err := task.New(".bit").Load("BIT-1")
	if err != nil {
		t.Fatalf("loading BIT-1: %v", err)
	}

	want := task.Task{
		ID:     "BIT-1",
		Title:  "Set up init wizard",
		Status: "todo",
		Body:   "Add flags for prefix capture.",
	}
	if *got != want {
		t.Errorf("task = %+v, want %+v", *got, want)
	}
}

func TestTaskCreateCmd_EchoesMintedID(t *testing.T) {
	initProject(t, "BIT")

	out := mustRun(t, "task", "create", "First track", "-d", "...")

	if out != "BIT-1\n" {
		t.Errorf("task create stdout = %q, want %q", out, "BIT-1\n")
	}
}

func TestTaskCreateCmd_EchoesSecondTrackID(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "First", "...")

	out := mustRun(t, "task", "create", "Second", "-d", "...")

	if out != "BIT-2\n" {
		t.Errorf("task create stdout = %q, want %q", out, "BIT-2\n")
	}
}

func TestTaskCreateCmd_EchoesChildID(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	out := mustRun(t, "task", "create", "A bar", "-d", "...", "--parent", "BIT-1")

	if out != "BIT-1.1\n" {
		t.Errorf("task create stdout = %q, want %q", out, "BIT-1.1\n")
	}
}

func TestTaskCreateCmd_AssignsNextIDWhenTasksExist(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "First", "...")
	createTask(t, "Second", "...")

	got, err := task.New(".bit").Load("BIT-2")
	if err != nil {
		t.Fatalf("loading BIT-2: %v", err)
	}
	if got.Title != "Second" {
		t.Errorf("BIT-2 title = %q, want %q", got.Title, "Second")
	}
}

func TestTaskCreateCmd_ParentMintsDottedID(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	mustRun(t, "task", "create", "A step", "-d", "...", "--parent", "BIT-1")

	out := mustRun(t, "task", "read", "BIT-1.1")
	want := "BIT-1.1\ttodo\tA step\n"
	if !strings.HasPrefix(out, want) {
		t.Errorf("task read BIT-1.1 first line = %q, want prefix %q", out, want)
	}
}

func TestTaskCreateCmd_ErrorsOnMissingParent(t *testing.T) {
	initProject(t, "BIT")

	if _, err := run(t, "task", "create", "Orphan", "-d", "...", "--parent", "BIT-99"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for a missing parent")
	}

	if _, err := os.Stat(".bit/tasks/BIT-99.1.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/tasks/BIT-99.1.md) error = %v, want fs.ErrNotExist", err)
	}
}

func TestTaskCreateCmd_SecondChildIncrements(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	mustRun(t, "task", "create", "First step", "-d", "...", "--parent", "BIT-1")
	mustRun(t, "task", "create", "Second step", "-d", "...", "--parent", "BIT-1")
	mustRun(t, "task", "create", "Third step", "-d", "...", "--parent", "BIT-1")

	out := mustRun(t, "task", "list")
	for _, id := range []string{"BIT-1.1", "BIT-1.2", "BIT-1.3"} {
		if !strings.Contains(out, id) {
			t.Errorf("task list = %q, want it to contain %q", out, id)
		}
	}
}

func TestTaskCreateCmd_ErrorsWithoutTitle(t *testing.T) {
	initProject(t, "BIT")

	if _, err := run(t, "task", "create"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for missing title argument")
	}
}

func TestTaskCreateCmd_ErrorsWithoutConfig(t *testing.T) {
	t.Chdir(t.TempDir())

	if _, err := run(t, "task", "create", "Foo"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil when config.toml is absent")
	}

	if _, err := os.Stat(".bit/tasks"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/tasks) error = %v, want fs.ErrNotExist", err)
	}
}
