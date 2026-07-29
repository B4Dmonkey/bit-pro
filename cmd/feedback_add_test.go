package cmd

import (
	"os"
	"testing"
)

const firstNote = "Happened at BIT-1.9.\n\n" +
	"The plan said: fall back to `plugin install` when `plugin update` fails.\n" +
	"The work required: deciding whether the fallback also runs `marketplace add`, " +
	"which the plan did not settle.\n"

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
