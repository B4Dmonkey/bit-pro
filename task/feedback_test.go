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

func feedbackStore(t *testing.T) (*Store, string) {
	t.Helper()

	d := t.TempDir()
	s := NewProject(filepath.Join(d, tprefix), tprefix).WithDataRoot(d)

	if err := s.Save(&Task{ID: tid1, Title: ttrack, Status: StatusTodo}); err != nil {
		t.Fatalf("seeding %s: %v", tid1, err)
	}

	return s, d
}

func TestStoreAddNote(t *testing.T) {
	t.Run("writes a json record beside the body", func(t *testing.T) {
		t.Parallel()

		s, root := feedbackStore(t)
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

		s, root := feedbackStore(t)

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

		s, root := feedbackStore(t)

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

		s, root := feedbackStore(t)

		if err := s.Relocate(tid1, false); err != nil {
			t.Fatalf("Relocate() returned error: %v", err)
		}

		if _, err := s.AddNote(tid1, tnote, Commit{}); err != nil {
			t.Fatalf("AddNote() returned error: %v", err)
		}

		readNoteJSON(t, root, tid1+"-001")
	})
	t.Run("two projects share the folder", func(t *testing.T) {
		t.Parallel()

		d := t.TempDir()

		for _, code := range []string{"BIT", "EX"} {
			s := NewProject(filepath.Join(d, code), code).WithDataRoot(d)
			track := code + "-1"

			if err := s.Save(&Task{ID: track, Title: ttrack, Status: StatusTodo}); err != nil {
				t.Fatalf("seeding %s: %v", track, err)
			}

			path, err := s.AddNote(track, tnote, Commit{})
			if err != nil {
				t.Fatalf("AddNote(%s) returned error: %v", track, err)
			}

			if want := filepath.Join(d, "feedback", track+"-001.md"); path != want {
				t.Errorf("path = %q, want %q", path, want)
			}
		}

		for _, code := range []string{"BIT", "EX"} {
			if got := readNoteJSON(t, d, code+"-1-001")[kproject]; got != code {
				t.Errorf("%s-1-001 project = %v, want %s", code, got, code)
			}
		}
	})

	t.Run("refuses without a data root", func(t *testing.T) {
		t.Chdir(t.TempDir())

		dir := t.TempDir()
		s := NewProject(dir, tprefix)

		if err := s.Save(&Task{ID: tid1, Title: ttrack, Status: StatusTodo}); err != nil {
			t.Fatalf("seeding %s: %v", tid1, err)
		}

		if _, err := s.AddNote(tid1, tnote, Commit{}); err == nil {
			t.Fatal("AddNote() returned no error")
		}

		entries, err := os.ReadDir(".")
		if err != nil {
			t.Fatal(err)
		}

		if len(entries) != 0 {
			t.Errorf("cwd holds %v, want nothing", entries)
		}
	})
}
