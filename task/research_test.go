package task

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const (
	ttopic     = "claim-audit"
	tresearch1 = "# Audit\n\n| a | b |\n"
	tresearch2 = "# Audit\n\nrewritten\n"
)

var (
	tcommitAt = time.Date(2026, 10, 2, 10, 0, 0, 0, time.FixedZone("EDT", -4*60*60))
	tcommitA  = Commit{SHA: "aaa111", Branch: "v2", At: tcommitAt}
)

func commitJSON(sha, branch string) map[string]any {
	return map[string]any{"sha": sha, "branch": branch, "at": "2026-10-02T14:00:00Z"}
}

func researchStore(t *testing.T) (*Store, string) {
	t.Helper()

	root := t.TempDir()
	s := NewProject(root, tprefix)

	if err := s.Save(&Task{ID: tid1, Title: ttrack, Status: StatusTodo}); err != nil {
		t.Fatalf("seeding %s: %v", tid1, err)
	}

	return s, root
}

func readResearchJSON(t *testing.T, root, topic string) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(root, "research", tid1, topic+".json"))
	if err != nil {
		t.Fatalf("reading research record %s: %v", topic, err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshaling research record %s: %v", topic, err)
	}

	return got
}

func TestStoreWriteResearch(t *testing.T) {
	t.Parallel()

	t.Run("writes a json record beside the body", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)
		at := tclock1
		s.now = fixedClock(&at)

		path, err := s.WriteResearch(tid1, ttopic, tresearch1, Commit{})
		if err != nil {
			t.Fatalf("WriteResearch() returned error: %v", err)
		}

		wantPath := filepath.Join(root, "research", tid1, ttopic+".md")
		if path != wantPath {
			t.Errorf("path = %q, want %q", path, wantPath)
		}

		body, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != tresearch1 {
			t.Errorf("body = %q, want %q", body, tresearch1)
		}

		rec := readResearchJSON(t, root, ttopic)

		want := map[string]any{
			"project":    tprefix,
			"track":      tid1,
			"topic":      ttopic,
			"commits":    []any{},
			"content":    ttopic + ".md",
			"created_at": tstamp1,
			"updated_at": tstamp1,
		}
		if !reflect.DeepEqual(rec, want) {
			t.Errorf("record = %v, want %v", rec, want)
		}
	})

	t.Run("rewrite keeps created at", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)
		at := tclock1
		s.now = fixedClock(&at)

		if _, err := s.WriteResearch(tid1, ttopic, tresearch1, Commit{}); err != nil {
			t.Fatalf("first WriteResearch() returned error: %v", err)
		}

		at = tclock2

		path, err := s.WriteResearch(tid1, ttopic, tresearch2, Commit{})
		if err != nil {
			t.Fatalf("second WriteResearch() returned error: %v", err)
		}

		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != tresearch2 {
			t.Errorf("body = %q, want %q", body, tresearch2)
		}

		assertStamps(t, readResearchJSON(t, root, ttopic), tstamp1, tstamp2)
	})

	t.Run("records the creating commit", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)

		if _, err := s.WriteResearch(tid1, ttopic, tresearch1, tcommitA); err != nil {
			t.Fatalf("WriteResearch() returned error: %v", err)
		}

		got := readResearchJSON(t, root, ttopic)["commits"]
		want := []any{commitJSON("aaa111", "v2")}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("commits = %v, want %v", got, want)
		}
	})

	t.Run("commit history", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			first Commit
			next  *Commit
			want  []any
		}{
			{
				name:  "same sha keeps one entry",
				first: tcommitA,
				next:  &tcommitA,
				want:  []any{commitJSON("aaa111", "v2")},
			},
			{
				name:  "same sha new branch keeps one entry",
				first: tcommitA,
				next:  &Commit{SHA: "aaa111", Branch: "main", At: tcommitAt},
				want:  []any{commitJSON("aaa111", "v2")},
			},
			{
				name:  "moved sha appends",
				first: tcommitA,
				next:  &Commit{SHA: "bbb222", Branch: "v2", At: tcommitAt},
				want:  []any{commitJSON("aaa111", "v2"), commitJSON("bbb222", "v2")},
			},
			{
				name:  "empty commit leaves the history",
				first: tcommitA,
				next:  &Commit{},
				want:  []any{commitJSON("aaa111", "v2")},
			},
			{
				name:  "first write without a commit",
				first: Commit{},
				want:  []any{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s, root := researchStore(t)

				if _, err := s.WriteResearch(tid1, ttopic, tresearch1, tt.first); err != nil {
					t.Fatalf("first WriteResearch() returned error: %v", err)
				}

				if tt.next != nil {
					path, err := s.WriteResearch(tid1, ttopic, tresearch2, *tt.next)
					if err != nil {
						t.Fatalf("second WriteResearch() returned error: %v", err)
					}

					body, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}

					if string(body) != tresearch2 {
						t.Errorf("body = %q, want %q", body, tresearch2)
					}
				}

				got := readResearchJSON(t, root, ttopic)["commits"]
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("commits = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestStoreResearchTopics(t *testing.T) {
	t.Parallel()

	t.Run("lists records only", func(t *testing.T) {
		t.Parallel()

		s, root := researchStore(t)

		for _, topic := range []string{"alpha", "beta"} {
			if _, err := s.WriteResearch(tid1, topic, tresearch1, Commit{}); err != nil {
				t.Fatalf("WriteResearch(%s) returned error: %v", topic, err)
			}
		}

		orphan := filepath.Join(root, "research", tid1, "orphan.md")
		if err := os.WriteFile(orphan, []byte(tresearch1), fileMode); err != nil {
			t.Fatal(err)
		}

		got, err := s.ResearchTopics(tid1)
		if err != nil {
			t.Fatalf("ResearchTopics() returned error: %v", err)
		}

		if want := []string{"alpha", "beta"}; !reflect.DeepEqual(got, want) {
			t.Errorf("topics = %v, want %v", got, want)
		}
	})
}
