package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/project"
)

func receiptPayload(dir string) string {
	return fmt.Sprintf(
		`{"session_id":"0d6f0c2b-1a26-4dbc-b324-a6401ead76a8","transcript_path":%q,"cwd":%q,"hook_event_name":"Stop"}`,
		filepath.Join(dir, "t.jsonl"), dir,
	)
}

func TestReceiptCmd(t *testing.T) {
	t.Run("unregistered cwd is a silent no-op", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")
		dir := t.TempDir()

		out, err := runWithStdin(t, receiptPayload(dir), "receipt")
		if err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want empty", out)
		}
	})

	t.Run("hidden from help", func(t *testing.T) {
		out := mustRun(t, "--help")

		if strings.Contains(out, "receipt") {
			t.Errorf("help output lists receipt:\n%s", out)
		}
	})

	t.Run("removed project is a silent no-op", func(t *testing.T) {
		dir := initProject(t, testPrefix)
		markRemoved(t, testPrefix)

		out, err := runWithStdin(t, receiptPayload(dir), "receipt")
		if err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want empty", out)
		}
	})

	t.Run("registered project is a silent no-op for now", func(t *testing.T) {
		dir := initProject(t, testPrefix)

		canonical, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
		}

		out, err := runWithStdin(t, receiptPayload(canonical), "receipt")
		if err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want empty", out)
		}
	})
}
