package task

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const tnote = "## What the plan said\n\n...\n"

func readNoteJSON(t *testing.T, root, id string) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(root, "feedback", id+".json"))
	if err != nil {
		t.Fatalf("reading note record %s: %v", id, err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshaling note record %s: %v", id, err)
	}

	return got
}

func TestStoreAddNote(t *testing.T) {
	t.Parallel()

	t.Run("writes a json record beside the body", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)
		at := tclock1
		s.now = fixedClock(&at)

		path, err := s.AddNote(tid1, tnote, tcommitA)
		if err != nil {
			t.Fatalf("AddNote() returned error: %v", err)
		}

		wantPath := filepath.Join(root, "feedback", tid1+"-001.md")
		if path != wantPath {
			t.Errorf("path = %q, want %q", path, wantPath)
		}

		body, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != tnote {
			t.Errorf("body = %q, want %q", body, tnote)
		}

		rec := readNoteJSON(t, root, tid1+"-001")

		want := map[string]any{
			kproject:   tprefix,
			"id":       tid1 + "-001",
			"track":    tid1,
			"seq":      float64(1),
			"commits":  []any{commitJSON("aaa111", "v2")},
			kcontent:   tid1 + "-001.md",
			kcreatedAt: tstamp1,
			kupdatedAt: tstamp1,
		}
		if !reflect.DeepEqual(rec, want) {
			t.Errorf("record = %v, want %v", rec, want)
		}
	})

	t.Run("second note counts records only", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)

		if _, err := s.AddNote(tid1, tnote, Commit{}); err != nil {
			t.Fatalf("first AddNote() returned error: %v", err)
		}

		stray := filepath.Join(root, "feedback", tid1+"-003.md")
		if err := os.WriteFile(stray, []byte(tnote), 0o600); err != nil {
			t.Fatal(err)
		}

		path, err := s.AddNote(tid1, tnote, Commit{})
		if err != nil {
			t.Fatalf("second AddNote() returned error: %v", err)
		}

		if want := filepath.Join(root, "feedback", tid1+"-002.md"); path != want {
			t.Errorf("path = %q, want %q", path, want)
		}

		rec := readNoteJSON(t, root, tid1+"-002")
		if rec["id"] != tid1+"-002" || rec["seq"] != float64(2) {
			t.Errorf("id, seq = %v, %v, want %s-002, 2", rec["id"], rec["seq"], tid1)
		}
	})

	t.Run("no commit writes an empty list", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)

		if _, err := s.AddNote(tid1, tnote, Commit{}); err != nil {
			t.Fatalf("AddNote() returned error: %v", err)
		}

		got := readNoteJSON(t, root, tid1+"-001")["commits"]
		if !reflect.DeepEqual(got, []any{}) {
			t.Errorf("commits = %v, want []", got)
		}
	})

	t.Run("accepts an archived track", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)

		if err := s.Relocate(tid1, false); err != nil {
			t.Fatalf("Relocate() returned error: %v", err)
		}

		if _, err := s.AddNote(tid1, tnote, Commit{}); err != nil {
			t.Fatalf("AddNote() returned error: %v", err)
		}

		readNoteJSON(t, root, tid1+"-001")
	})
}
