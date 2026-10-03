package migrate

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	headSHA     = "9f3c5e7a0b1d2c3e4f5a6b7c8d9e0f1a2b3c4d5e"
	headBranch  = "v2"
	revParse    = "rev-parse HEAD"
	symbolicRef = "symbolic-ref --short -q HEAD"
)

type fakeResult struct {
	out string
	err error
}

type fakeGit struct {
	results map[string]fakeResult
	dirs    []string
}

func (f *fakeGit) run(_ context.Context, dir string, args ...string) (string, error) {
	f.dirs = append(f.dirs, dir)

	r, ok := f.results[strings.Join(args, " ")]
	if !ok {
		return "", errors.New("unexpected git " + strings.Join(args, " "))
	}

	return r.out, r.err
}

func TestRun(t *testing.T) {
	fixed := time.Date(2026, 10, 2, 10, 0, 0, 0, time.FixedZone("EDT", -4*60*60))
	now := func() time.Time { return fixed }

	t.Run("stamps every record with the session head", func(t *testing.T) {
		dir := writeFixture(t)
		fake := &fakeGit{results: map[string]fakeResult{
			revParse:    {out: headSHA + "\n"},
			symbolicRef: {out: headBranch + "\n"},
		}}

		runMigrate(t, Options{Dir: dir, Git: fake.run, Now: now})

		data, root := dirs(t)

		for _, path := range taskRecords(root) {
			rec := readJSON(t, path)
			if rec["branch"] != headBranch || rec["commit"] != headSHA {
				t.Errorf("%s = {branch %v, commit %v}, want {%s, %s}", path, rec["branch"], rec["commit"], headBranch, headSHA)
			}
		}

		want := []any{map[string]any{"sha": headSHA, "branch": headBranch, "at": "2026-10-02T14:00:00Z"}}

		for _, path := range sharedRecords(data, root) {
			if got := readJSON(t, path)["commits"]; !reflect.DeepEqual(got, want) {
				t.Errorf("%s commits = %v, want %v", path, got, want)
			}
		}

		for _, d := range fake.dirs {
			if d != dir {
				t.Errorf("git dir = %q, want %q", d, dir)
			}
		}

		if len(fake.dirs) == 0 {
			t.Error("git was never called")
		}
	})

	t.Run("a folder with no git leaves git fields empty", func(t *testing.T) {
		dir := writeFixture(t)
		fake := &fakeGit{results: map[string]fakeResult{}}

		runMigrate(t, Options{Dir: dir, Git: fake.run, Now: now})

		data, root := dirs(t)

		for _, path := range taskRecords(root) {
			rec := readJSON(t, path)
			if rec["branch"] != "" || rec["commit"] != "" {
				t.Errorf("%s = {branch %v, commit %v}, want empty", path, rec["branch"], rec["commit"])
			}
		}

		for _, path := range sharedRecords(data, root) {
			if got := readJSON(t, path)["commits"]; !reflect.DeepEqual(got, []any{}) {
				t.Errorf("%s commits = %v, want []", path, got)
			}
		}
	})
}

func runMigrate(t *testing.T, opts Options) {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	t.Cleanup(func() { _ = sqlDB.Close() })

	if _, err := Run(t.Context(), orm.New(sqlDB), opts); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
}

func writeFixture(t *testing.T) string {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	files := map[string][]byte{
		"config.toml": []byte("prefix = \"BIT\"\n"),
		filepath.Join("feedback", "BIT-1-001.md"):      []byte("## What happened\n\nA note.\n"),
		filepath.Join("research", "BIT-1", "index.md"): []byte("## Findings\n"),
		filepath.Join("retro", "album-proposals.md"):   []byte("## Proposal 1\n"),
	}

	for place, tk := range map[string]*task.Task{
		"tasks":     {ID: "BIT-1", Title: "Track", Status: task.StatusDoing, Order: []string{"BIT-1.1"}},
		"completed": {ID: "BIT-1.1", Title: "Bar", Status: task.StatusDone},
	} {
		raw, err := tk.Bytes()
		if err != nil {
			t.Fatalf("Bytes(%s) returned error: %v", tk.ID, err)
		}

		files[filepath.Join(place, tk.ID+".md")] = raw
	}

	dir := t.TempDir()

	for name, raw := range files {
		path := filepath.Join(dir, ".bit", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
		}

		if err := os.WriteFile(path, raw, 0o600); err != nil {
			t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
		}
	}

	return dir
}

func dirs(t *testing.T) (data, root string) {
	t.Helper()

	data, err := store.Dir()
	if err != nil {
		t.Fatalf("store.Dir() returned error: %v", err)
	}

	root, err = store.ProjectDir("BIT")
	if err != nil {
		t.Fatalf("store.ProjectDir() returned error: %v", err)
	}

	return data, root
}

func taskRecords(root string) []string {
	return []string{
		filepath.Join(root, "tasks", "BIT-1.json"),
		filepath.Join(root, "completed", "BIT-1.1.json"),
	}
}

func sharedRecords(data, root string) []string {
	return []string{
		filepath.Join(data, "feedback", "BIT-1-001.json"),
		filepath.Join(root, "research", "BIT-1", "index.json"),
		filepath.Join(data, "retro", "BIT-album-proposals.json"),
	}
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	var rec map[string]any
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("json.Unmarshal(%q) returned error: %v", path, err)
	}

	return rec
}
