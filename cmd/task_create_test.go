package cmd

import (
	"errors"
	"io/fs"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestTaskCreateCmd_WritesFirstTask(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Set up init wizard", "Add flags for prefix capture.")

	got, err := task.New(".bit").Load(trackID)
	if err != nil {
		t.Fatalf("loading BIT-1: %v", err)
	}

	want := task.Task{
		ID:     trackID,
		Title:  "Set up init wizard",
		Status: statusTodo,
		Body:   "Add flags for prefix capture.",
	}
	if !reflect.DeepEqual(*got, want) {
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

	out := mustRun(t, "task", "create", "A bar", "-d", "...", "--parent", trackID)

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

	mustRun(t, "task", "create", "A step", "-d", "...", "--parent", trackID)

	out := mustRun(t, "task", "read", firstBarID)

	want := firstBarID + "\ttodo\tA step\n"
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

	mustRun(t, "task", "create", "First step", "-d", "...", "--parent", trackID)
	mustRun(t, "task", "create", "Second step", "-d", "...", "--parent", trackID)
	mustRun(t, "task", "create", "Third step", "-d", "...", "--parent", trackID)

	out := mustRun(t, "task", "list")
	for _, id := range []string{firstBarID, secondBarID, thirdBarID} {
		if !strings.Contains(out, id) {
			t.Errorf("task list = %q, want it to contain %q", out, id)
		}
	}
}

func TestTaskCreate_AppendsToReorderedTrack(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", trackID)
	mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", trackID)
	mustRun(t, "task", "move", secondBarID, "--before", firstBarID)

	mustRun(t, "task", "create", "Third bar", "-d", "...", "--parent", trackID)

	track, err := task.New(".bit").Load(trackID)
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
}

func TestTaskCreate_AfterInsertsMidPlan(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", trackID)
	mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", trackID)

	out := mustRun(t, "task", "create", "Inserted", "-d", "...", "--parent", trackID, "--after", firstBarID)

	if out != thirdBarID+"\n" {
		t.Errorf("minted ID = %q, want %q", out, thirdBarID+"\n")
	}

	track, err := task.New(".bit").Load(trackID)
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
}

func TestTaskCreate_AfterRejectsUnknownAnchor(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "First bar", "-d", "...", "--parent", trackID)
	mustRun(t, "task", "create", "Second bar", "-d", "...", "--parent", trackID)

	if _, err := run(t, "task", "create", "Inserted", "-d", "...", "--parent", trackID, "--after", "BIT-1.9"); err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for an unknown anchor")
	}

	if _, err := os.Stat(".bit/tasks/BIT-1.3.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.bit/tasks/BIT-1.3.md) error = %v, want fs.ErrNotExist", err)
	}
}

func TestTaskCreateCmd_LowercaseParentDoesNotDestroyAnExistingBar(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")
	mustRun(t, "task", "create", "First bar", "-d", "ORIGINAL BAR ONE", "--parent", trackID)

	out := mustRun(t, "task", "create", "sneaky", "-d", "...", "--parent", "bit-1")

	if out != "BIT-1.2\n" {
		t.Errorf("minted ID = %q, want %q", out, "BIT-1.2\n")
	}

	minted, err := os.ReadFile(".bit/tasks/BIT-1.2.md")
	if err != nil {
		t.Fatalf("os.ReadFile(.bit/tasks/BIT-1.2.md) error = %v", err)
	}

	if !strings.Contains(string(minted), "id: BIT-1.2") {
		t.Errorf("BIT-1.2.md = %q, want it to contain %q", minted, "id: BIT-1.2")
	}

	survivor, err := os.ReadFile(".bit/tasks/BIT-1.1.md")
	if err != nil {
		t.Fatalf("os.ReadFile(.bit/tasks/BIT-1.1.md) error = %v", err)
	}

	if !strings.Contains(string(survivor), "ORIGINAL BAR ONE") {
		t.Errorf("BIT-1.1.md = %q, want it to still contain %q", survivor, "ORIGINAL BAR ONE")
	}

	if !strings.Contains(string(survivor), "title: First bar") {
		t.Errorf("BIT-1.1.md = %q, want it to still contain %q", survivor, "title: First bar")
	}

	entries, err := os.ReadDir(".bit/tasks")
	if err != nil {
		t.Fatalf("os.ReadDir(.bit/tasks) error = %v", err)
	}

	if len(entries) != 3 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}

		t.Errorf("os.ReadDir(.bit/tasks) names = %v, want 3 entries", names)
	}
}

func TestTaskCreateCmd_UppercasesACorruptPrefixOnRead(t *testing.T) {
	initProject(t, "BIT")

	if err := os.WriteFile(".bit/config.toml", []byte("prefix = \"bit\"\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(.bit/config.toml) error = %v", err)
	}

	out := mustRun(t, "task", "create", "first", "-d", "...")

	if out != "BIT-1\n" {
		t.Errorf("minted ID = %q, want %q", out, "BIT-1\n")
	}

	data, err := os.ReadFile(".bit/tasks/BIT-1.md")
	if err != nil {
		t.Fatalf("os.ReadFile(.bit/tasks/BIT-1.md) error = %v", err)
	}

	if !strings.Contains(string(data), "id: BIT-1") {
		t.Errorf("BIT-1.md = %q, want it to contain %q", data, "id: BIT-1")
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
