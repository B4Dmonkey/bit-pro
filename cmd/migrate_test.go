package cmd

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	migrateCmdUse = "migrate"
	migratedBar   = "BIT-1.1"
)

func TestMigrateCmd(t *testing.T) {
	t.Run("copies active tasks and registers the folder", func(t *testing.T) {
		mcpSandbox(t)

		dir := t.TempDir()
		files := v1Fixture(t)
		writeV1Store(t, dir, files)
		before := hashV1Store(t, dir, files)

		t.Chdir(dir)

		want, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("project.CanonicalPath(%q) returned error: %v", dir, err)
		}

		out, err := run(t, migrateCmdUse)
		if err != nil {
			t.Fatalf("bp migrate returned error: %v", err)
		}

		wantLine := "migrated " + testPrefix + " " + want
		if first := strings.SplitN(out, "\n", 2)[0]; first != wantLine {
			t.Errorf("first output line = %q, want %q", first, wantLine)
		}

		projects := listProjects(t)
		if len(projects) != 1 {
			t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
		}

		if projects[0].Code != testPrefix || projects[0].Path != want {
			t.Errorf("project = {%q, %q}, want {%q, %q}", projects[0].Code, projects[0].Path, testPrefix, want)
		}

		s, err := project.OpenStore(t.Context(), dir)
		if err != nil {
			t.Fatalf("project.OpenStore(%q) returned error: %v", dir, err)
		}

		track, err := s.Load(testOwnTrack)
		if err != nil {
			t.Fatalf("Load(BIT-1) returned error: %v", err)
		}

		wantTrack := v1Track()
		if track.Title != wantTrack.Title || track.Approved != wantTrack.Approved ||
			!slices.Equal(track.Order, wantTrack.Order) || track.Body != wantTrack.Body {
			t.Errorf("BIT-1 = %+v, want %+v", track, wantTrack)
		}

		bar, err := s.Load(migratedBar)
		if err != nil {
			t.Fatalf("Load(BIT-1.1) returned error: %v", err)
		}

		wantBar := v1Bars()[0]
		if bar.Phase != wantBar.Phase || bar.PhaseLabel != wantBar.PhaseLabel || bar.Status != wantBar.Status {
			t.Errorf("BIT-1.1 = %+v, want %+v", bar, wantBar)
		}

		if after := hashV1Store(t, dir, files); !slices.Equal(after, before) {
			t.Errorf("source hashes changed: before %x, after %x", before, after)
		}
	})

	t.Run("a migrated project works with bp task commands", func(t *testing.T) {
		mcpSandbox(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		if out := mustRun(t, "task", "create", "Next"); out != "BIT-2\n" {
			t.Errorf("bp task create = %q, want %q", out, "BIT-2\n")
		}
	})
}

func v1Track() *task.Task {
	return &task.Task{
		ID:       testOwnTrack,
		Title:    "Skeleton",
		Status:   task.StatusDoing,
		Approved: true,
		Order:    []string{"BIT-1.2", migratedBar},
		Body:     "## Why\n\nBecause.\n\n## Verses\n\n- [ ] Verse 1 — skeleton\n- [ ] Verse 2 — more\n",
	}
}

func v1Bars() []*task.Task {
	return []*task.Task{
		{ID: migratedBar, Title: "First", Status: task.StatusDone, Phase: 1, PhaseLabel: "skeleton", Body: "one\n"},
		{ID: "BIT-1.2", Title: "Second", Status: task.StatusTodo, Phase: 1, PhaseLabel: "skeleton", Body: "two\n"},
	}
}

func v1Fixture(t *testing.T) map[string][]byte {
	t.Helper()

	files := map[string][]byte{"config.toml": []byte("prefix = \"BIT\"\n")}

	for _, tk := range append([]*task.Task{v1Track()}, v1Bars()...) {
		data, err := tk.Bytes()
		if err != nil {
			t.Fatalf("Bytes(%s) returned error: %v", tk.ID, err)
		}

		files[filepath.Join("tasks", tk.ID+".md")] = data
	}

	return files
}

func writeV1Store(t *testing.T, dir string, files map[string][]byte) {
	t.Helper()

	for name, data := range files {
		path := filepath.Join(dir, ".bit", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
		}

		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
		}
	}
}

func hashV1Store(t *testing.T, dir string, files map[string][]byte) [][sha256.Size]byte {
	t.Helper()

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}

	slices.Sort(names)

	sums := make([][sha256.Size]byte, 0, len(names))

	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, ".bit", name))
		if err != nil {
			t.Fatalf("os.ReadFile(%q) returned error: %v", name, err)
		}

		sums = append(sums, sha256.Sum256(data))
	}

	return sums
}
