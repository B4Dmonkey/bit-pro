package cmd

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

const firstNote = "Happened at BIT-1.9.\n\n" +
	"The plan said: fall back to `plugin install` when `plugin update` fails.\n" +
	"The work required: deciding whether the fallback also runs `marketplace add`, " +
	"which the plan did not settle.\n"

const secondNote = "Happened at BIT-1.10.\n\n" +
	"The plan said: register the marketplace in the install script.\n" +
	"The work required: choosing whether an already-registered marketplace is an error " +
	"or a no-op, which the plan did not settle.\n"

func TestFeedbackAddCmd_WritesFirstNote(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")

	out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

	if want := ".bit/feedback/BIT-1-001.md\n"; out != want {
		t.Errorf("feedback add stdout = %q, want %q", out, want)
	}

	data, err := os.ReadFile(".bit/feedback/BIT-1-001.md")
	if err != nil {
		t.Fatalf("reading note: %v", err)
	}
	if string(data) != firstNote {
		t.Errorf("note = %q, want %q", data, firstNote)
	}
}

func TestFeedbackAddCmd_SecondNoteGetsNextSequence(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")

	mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)
	out := mustRun(t, "feedback", "add", "BIT-1", "-d", secondNote)

	if want := ".bit/feedback/BIT-1-002.md\n"; out != want {
		t.Errorf("second feedback add stdout = %q, want %q", out, want)
	}

	second, err := os.ReadFile(".bit/feedback/BIT-1-002.md")
	if err != nil {
		t.Fatalf("reading second note: %v", err)
	}
	if string(second) != secondNote {
		t.Errorf("second note = %q, want %q", second, secondNote)
	}

	first, err := os.ReadFile(".bit/feedback/BIT-1-001.md")
	if err != nil {
		t.Fatalf("reading first note: %v", err)
	}
	if string(first) != firstNote {
		t.Errorf("first note = %q, want %q", first, firstNote)
	}
}

func TestFeedbackAddCmd_AcceptsArchivedTrack(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
	mustRun(t, "task", "update", "BIT-1", "-s", "done")
	mustRun(t, "task", "archive", "BIT-1")

	out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

	if want := ".bit/feedback/BIT-1-001.md\n"; out != want {
		t.Errorf("feedback add against an archived track stdout = %q, want %q", out, want)
	}

	data, err := os.ReadFile(".bit/feedback/BIT-1-001.md")
	if err != nil {
		t.Fatalf("reading note: %v", err)
	}
	if string(data) != firstNote {
		t.Errorf("note = %q, want %q", data, firstNote)
	}
}

func TestFeedbackAddCmd_AcceptsCompletedTrack(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
	mustRun(t, "task", "update", "BIT-1", "-s", "done")
	mustRun(t, "task", "complete", "BIT-1")

	out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

	if want := ".bit/feedback/BIT-1-001.md\n"; out != want {
		t.Errorf("feedback add against a completed track stdout = %q, want %q", out, want)
	}

	data, err := os.ReadFile(".bit/feedback/BIT-1-001.md")
	if err != nil {
		t.Fatalf("reading note: %v", err)
	}
	if string(data) != firstNote {
		t.Errorf("note = %q, want %q", data, firstNote)
	}
}

func TestFeedbackAddCmd_NoteSurvivesTrackRewrite(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
	mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

	mustRun(t, "task", "update", "BIT-1", "-d", "## Why\n\nA wholesale rewritten scope body.\n")

	data, err := os.ReadFile(".bit/feedback/BIT-1-001.md")
	if err != nil {
		t.Fatalf("reading note after track rewrite: %v", err)
	}
	if string(data) != firstNote {
		t.Errorf("note after track rewrite = %q, want %q", data, firstNote)
	}
}

func TestFeedbackAddCmd_NoteSurvivesTrackArchive(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
	mustRun(t, "task", "create", "A bar", "--parent", "BIT-1", "--description", "One step.")
	mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)
	mustRun(t, "task", "update", "BIT-1.1", "-s", "done")
	mustRun(t, "task", "update", "BIT-1", "-s", "done")

	mustRun(t, "task", "archive", "BIT-1")

	if _, err := os.Stat(".bit/tasks/BIT-1.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stat track under tasks = %v, want fs.ErrNotExist", err)
	}
	if _, err := os.Stat(".bit/archive/BIT-1.md"); err != nil {
		t.Errorf("stat archived track = %v, want it relocated", err)
	}

	data, err := os.ReadFile(".bit/feedback/BIT-1-001.md")
	if err != nil {
		t.Fatalf("reading note after track archive: %v", err)
	}
	if string(data) != firstNote {
		t.Errorf("note after track archive = %q, want %q", data, firstNote)
	}
}

func TestFeedbackAddCmd_ErrorsOnUnknownTrack(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")

	if _, err := run(t, "feedback", "add", "BIT-99", "-d", firstNote); err == nil {
		t.Fatal("feedback add against an unknown track returned no error")
	}

	if _, err := os.Stat(".bit/feedback/BIT-99-001.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stat note = %v, want fs.ErrNotExist", err)
	}

	if _, err := os.Stat(".bit/feedback"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stat feedback dir = %v, want fs.ErrNotExist", err)
	}
}
