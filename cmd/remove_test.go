package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
)

func TestRemoveCmd(t *testing.T) {
	t.Run("archives open tracks and soft deletes", func(t *testing.T) {
		dir := initProject(t, testPrefix)
		s := projectStore(t)
		root := storeDir(t)

		seedTrack(t, s, "open", task.StatusTodo, task.StatusTodo)
		seedTrack(t, s, "finished", task.StatusDone, task.StatusDone)
		seedTrack(t, s, "completed", task.StatusDone, task.StatusDone)

		if err := s.Complete("BIT-3"); err != nil {
			t.Fatalf("Complete(BIT-3) returned error: %v", err)
		}

		if _, err := s.WriteResearch(testNewTrackID, "index", "x", task.Commit{}); err != nil {
			t.Fatalf("WriteResearch() returned error: %v", err)
		}

		path, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
		}

		out, err := runWithStdin(t, "y\n", removeCmdUse)
		if err != nil {
			t.Fatalf("bp remove returned error: %v", err)
		}

		if !strings.Contains(out, "Outstanding work:") {
			t.Errorf("output = %q, want it to list outstanding work", out)
		}

		lines := outLines(out)
		for _, id := range []string{testNewTrackID, "BIT-1.1"} {
			if !lines[id] {
				t.Errorf("output = %q, want a line for %s", out, id)
			}
		}

		for _, id := range []string{"BIT-2", "BIT-2.1"} {
			if lines[id] {
				t.Errorf("output = %q, want no line for %s", out, id)
			}
		}

		if want := "removed BIT " + path; !strings.HasSuffix(strings.TrimSpace(out), want) {
			t.Errorf("output = %q, want it to end with %q", out, want)
		}

		if left, _ := filepath.Glob(filepath.Join(root, "tasks", "*.json")); len(left) != 0 {
			t.Errorf("tasks/ still holds %v, want no records", left)
		}

		for _, id := range []string{"BIT-1", "BIT-1.1", "BIT-2", "BIT-2.1"} {
			for _, ext := range []string{".json", ".md"} {
				assertExists(t, filepath.Join(root, "archive", "tasks", id+ext))
			}
		}

		assertExists(t, filepath.Join(root, "completed", "BIT-3.json"))
		assertExists(t, filepath.Join(root, "research", testNewTrackID, "index.md"))

		if p := loadProject(t, path); !p.Removed {
			t.Errorf("project %s Removed = false, want true", path)
		}
	})

	t.Run("decline changes nothing", func(t *testing.T) {
		tests := []struct {
			name  string
			stdin string
		}{
			{name: "explicit no", stdin: "n\n"},
			{name: "empty line", stdin: "\n"},
			{name: "eof", stdin: ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				dir := initProject(t, testPrefix)
				s := projectStore(t)
				root := storeDir(t)

				seedTrack(t, s, "open", task.StatusTodo, task.StatusTodo)

				before, _ := filepath.Glob(filepath.Join(root, "tasks", "*"))

				out, err := runWithStdin(t, tt.stdin, removeCmdUse)
				if err != nil {
					t.Fatalf("bp remove returned error: %v", err)
				}

				if !strings.HasSuffix(strings.TrimSpace(out), "not removed") {
					t.Errorf("output = %q, want it to end with %q", out, "not removed")
				}

				after, _ := filepath.Glob(filepath.Join(root, "tasks", "*"))
				if strings.Join(after, "\n") != strings.Join(before, "\n") {
					t.Errorf("tasks/ = %v, want unchanged %v", after, before)
				}

				path, err := project.CanonicalPath(dir)
				if err != nil {
					t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
				}

				if p := loadProject(t, path); p.Removed {
					t.Errorf("project %s Removed = true, want false", path)
				}
			})
		}
	})

	t.Run("no outstanding work skips the list", func(t *testing.T) {
		initProject(t, testPrefix)
		s := projectStore(t)

		seedTrack(t, s, "finished", task.StatusDone, task.StatusDone)
		seedTrack(t, s, "completed", task.StatusDone, task.StatusDone)

		if err := s.Complete("BIT-2"); err != nil {
			t.Fatalf("Complete(BIT-2) returned error: %v", err)
		}

		out, err := runWithStdin(t, "n\n", removeCmdUse)
		if err != nil {
			t.Fatalf("bp remove returned error: %v", err)
		}

		if strings.Contains(out, "Outstanding work:") {
			t.Errorf("output = %q, want no outstanding work list", out)
		}

		if !strings.Contains(out, "[y/N]") {
			t.Errorf("output = %q, want the confirmation prompt", out)
		}
	})

	t.Run("then the folder says removed", func(t *testing.T) {
		initProject(t, testPrefix)

		if _, err := runWithStdin(t, "y\n", removeCmdUse); err != nil {
			t.Fatalf("bp remove returned error: %v", err)
		}

		if _, err := run(t, "task", "list"); !errors.Is(err, project.ErrRemoved) {
			t.Errorf("bp task list error = %v, want %v", err, project.ErrRemoved)
		}

		if _, err := runWithStdin(t, "y\n", removeCmdUse); !errors.Is(err, project.ErrRemoved) {
			t.Errorf("second bp remove error = %v, want %v", err, project.ErrRemoved)
		}
	})
}

func seedTrack(t *testing.T, s *task.Store, title, trackStatus, barStatus string) {
	t.Helper()

	track, err := s.Create(task.CreateParams{Title: title})
	if err != nil {
		t.Fatalf("Create(%q) returned error: %v", title, err)
	}

	bar, err := s.Create(task.CreateParams{Title: title + " bar", Parent: track.ID})
	if err != nil {
		t.Fatalf("Create(%q bar) returned error: %v", title, err)
	}

	setStatus(t, s, track.ID, trackStatus)
	setStatus(t, s, bar.ID, barStatus)
}

func setStatus(t *testing.T, s *task.Store, id, status string) {
	t.Helper()

	if _, err := s.Update(id, task.Patch{Status: &status}); err != nil {
		t.Fatalf("Update(%s, status %s) returned error: %v", id, status, err)
	}
}

func outLines(out string) map[string]bool {
	ids := make(map[string]bool)

	for _, line := range strings.Split(out, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			ids[fields[0]] = true
		}
	}

	return ids
}

func assertExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Errorf("os.Stat(%s) returned error: %v", path, err)
	}
}

func loadProject(t *testing.T, path string) project.Project {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	projects, err := project.Load(t.Context(), orm.New(sqlDB))
	if err != nil {
		t.Fatalf("project.Load() returned error: %v", err)
	}

	for _, p := range projects {
		if p.Path == path {
			return p
		}
	}

	t.Fatalf("no project registered at %s", path)

	return project.Project{}
}
