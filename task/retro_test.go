package task

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRetroName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		label, code, name, want string
	}{
		{label: "adds the prefix", code: tprefix, name: "album-proposals", want: "BIT-album-proposals"},
		{label: "keeps an existing prefix", code: tprefix, name: "BIT-49-proposals", want: "BIT-49-proposals"},
		{label: "keeps a repeated code", code: "ACME", name: "ACME-1-ACME-4-proposals", want: "ACME-1-ACME-4-proposals"},
		{label: "is case-sensitive", code: tprefix, name: "bit-49-proposals", want: "BIT-bit-49-proposals"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			t.Parallel()

			if got := RetroName(tt.code, tt.name); got != tt.want {
				t.Errorf("RetroName(%q, %q) = %q, want %q", tt.code, tt.name, got, tt.want)
			}
		})
	}
}

func retroStore(t *testing.T) (*Store, string) {
	t.Helper()

	d := t.TempDir()

	return NewProject(filepath.Join(d, tprefix), tprefix).WithDataRoot(d), d
}

func TestStoreWriteRetro(t *testing.T) {
	t.Run("replaces a re-run and keeps created_at", func(t *testing.T) {
		t.Parallel()

		s, root := retroStore(t)
		at := tclock1
		s.now = fixedClock(&at)

		if _, err := s.WriteRetro("album-proposals", "first\n", Commit{}); err != nil {
			t.Fatalf("first WriteRetro() returned error: %v", err)
		}

		at = tclock2

		name, err := s.WriteRetro("album-proposals", "second\n", Commit{})
		if err != nil {
			t.Fatalf("second WriteRetro() returned error: %v", err)
		}

		if name != "BIT-album-proposals" {
			t.Errorf("name = %q, want %q", name, "BIT-album-proposals")
		}

		body, err := os.ReadFile(filepath.Join(root, "retro", name+".md"))
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != "second\n" {
			t.Errorf("body = %q, want %q", body, "second\n")
		}

		raw, err := os.ReadFile(filepath.Join(root, "retro", name+".json"))
		if err != nil {
			t.Fatal(err)
		}

		var rec map[string]any
		if err := json.Unmarshal(raw, &rec); err != nil {
			t.Fatal(err)
		}

		assertStamps(t, rec, tstamp1, tstamp2)
	})

	t.Run("refuses a path-like name", func(t *testing.T) {
		t.Parallel()

		for _, name := range []string{"", "../x", "a/b"} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				s, root := retroStore(t)

				if _, err := s.WriteRetro(name, "body\n", Commit{}); err == nil {
					t.Errorf("WriteRetro(%q) returned nil error", name)
				}

				entries, err := os.ReadDir(filepath.Join(root, "retro"))
				if err == nil && len(entries) != 0 {
					t.Errorf("retro dir holds %d entries, want none", len(entries))
				}
			})
		}
	})
}
