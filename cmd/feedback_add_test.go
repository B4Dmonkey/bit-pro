package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/B4Dmonkey/bit-pro/git/gittest"
)

const firstNote = "Happened at BIT-1.9.\n\n" +
	"The plan said: fall back to `plugin install` when `plugin update` fails.\n" +
	"The work required: deciding whether the fallback also runs `marketplace add`, " +
	"which the plan did not settle.\n"

const secondNote = "Happened at BIT-1.10.\n\n" +
	"The plan said: register the marketplace in the install script.\n" +
	"The work required: choosing whether an already-registered marketplace is an error " +
	"or a no-op, which the plan did not settle.\n"

func TestFeedbackAddCmd(t *testing.T) {
	t.Run("writes first note", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")

		out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

		if out != notePath(t, "BIT-1-001.md") {
			t.Errorf("feedback add stdout = %q, want %q", out, notePath(t, "BIT-1-001.md"))
		}

		data, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-001.md"))
		if err != nil {
			t.Fatalf("reading note: %v", err)
		}

		if string(data) != firstNote {
			t.Errorf("note = %q, want %q", data, firstNote)
		}
	})

	t.Run("second note gets next sequence", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")

		mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)
		out := mustRun(t, "feedback", "add", "BIT-1", "-d", secondNote)

		if want := notePath(t, "BIT-1-002.md"); out != want {
			t.Errorf("second feedback add stdout = %q, want %q", out, want)
		}

		second, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-002.md"))
		if err != nil {
			t.Fatalf("reading second note: %v", err)
		}

		if string(second) != secondNote {
			t.Errorf("second note = %q, want %q", second, secondNote)
		}

		first, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-001.md"))
		if err != nil {
			t.Fatalf("reading first note: %v", err)
		}

		if string(first) != firstNote {
			t.Errorf("first note = %q, want %q", first, firstNote)
		}
	})

	t.Run("lowercase track does not overwrite an existing note", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")

		mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)
		out := mustRun(t, "feedback", "add", "bit-1", "-d", secondNote)

		if want := notePath(t, "BIT-1-002.md"); out != want {
			t.Errorf("lowercase feedback add stdout = %q, want %q", out, want)
		}

		first, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-001.md"))
		if err != nil {
			t.Fatalf("reading first note: %v", err)
		}

		if string(first) != firstNote {
			t.Errorf("first note = %q, want %q", first, firstNote)
		}

		entries, err := os.ReadDir(filepath.Join(storeDir(t), "feedback"))
		if err != nil {
			t.Fatalf("reading feedback dir: %v", err)
		}

		if len(entries) != 4 {
			t.Fatalf("feedback dir holds %d files, want 4", len(entries))
		}

		for _, entry := range entries {
			stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			if stem != strings.ToUpper(stem) {
				t.Errorf("note filename = %q, want its track ID uppercase", entry.Name())
			}
		}
	})

	t.Run("accepts archived track", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
		mustRun(t, "task", "update", "BIT-1", "-s", "done")
		mustRun(t, "task", "delete", "BIT-1", "--yes")

		out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

		if out != notePath(t, "BIT-1-001.md") {
			t.Errorf("feedback add against an archived track stdout = %q, want %q", out, notePath(t, "BIT-1-001.md"))
		}

		data, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-001.md"))
		if err != nil {
			t.Fatalf("reading note: %v", err)
		}

		if string(data) != firstNote {
			t.Errorf("note = %q, want %q", data, firstNote)
		}
	})

	t.Run("accepts completed track", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
		mustRun(t, "task", "update", "BIT-1", "-s", "done")
		mustRun(t, "task", "complete", "BIT-1")

		out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

		if out != notePath(t, "BIT-1-001.md") {
			t.Errorf("feedback add against a completed track stdout = %q, want %q", out, notePath(t, "BIT-1-001.md"))
		}

		data, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-001.md"))
		if err != nil {
			t.Fatalf("reading note: %v", err)
		}

		if string(data) != firstNote {
			t.Errorf("note = %q, want %q", data, firstNote)
		}
	})

	t.Run("note survives track rewrite", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
		mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

		mustRun(t, "task", "update", "BIT-1", "-d", "## Why\n\nA wholesale rewritten scope body.\n")

		data, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-001.md"))
		if err != nil {
			t.Fatalf("reading note after track rewrite: %v", err)
		}

		if string(data) != firstNote {
			t.Errorf("note after track rewrite = %q, want %q", data, firstNote)
		}
	})

	t.Run("note survives track completion", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")
		mustRun(t, "task", "create", "A bar", "--parent", "BIT-1", "--description", "One step.")
		mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)
		mustRun(t, "task", "update", "BIT-1.1", "-s", "done")
		mustRun(t, "task", "update", "BIT-1", "-s", "done")

		mustRun(t, "task", "complete", "BIT-1")

		if _, err := os.Stat(filepath.Join(storeDir(t), "tasks", "BIT-1.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat track under tasks = %v, want fs.ErrNotExist", err)
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "completed", "BIT-1.json")); err != nil {
			t.Errorf("stat completed track = %v, want it relocated", err)
		}

		data, err := os.ReadFile(filepath.Join(storeDir(t), "feedback", "BIT-1-001.md"))
		if err != nil {
			t.Fatalf("reading note after track completion: %v", err)
		}

		if string(data) != firstNote {
			t.Errorf("note after track completion = %q, want %q", data, firstNote)
		}
	})

	t.Run("errors on unknown track", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Ship the bit plugin", "## Why\n\nThe skills only exist in this repo.\n")

		if _, err := run(t, "feedback", "add", "BIT-99", "-d", firstNote); err == nil {
			t.Fatal("feedback add against an unknown track returned no error")
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "feedback", "BIT-99-001.md")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat note = %v, want fs.ErrNotExist", err)
		}

		if _, err := os.Stat(filepath.Join(storeDir(t), "feedback")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat feedback dir = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("records the current folder's head", func(t *testing.T) {
		gittest.Isolate(t)

		dir := initProject(t, "BIT")
		createTask(t, "Track", "## Why\n\nA track.\n")
		gittest.Run(t, dir, "init", "-b", "main")
		gittest.Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "--allow-empty", "-m", "init")
		want := gittest.Run(t, dir, "rev-parse", "HEAD")

		out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

		commits := recordCommits(t, strings.TrimSpace(out))
		if len(commits) != 1 {
			t.Fatalf("commits = %v, want one", commits)
		}

		if len(want) != 40 {
			t.Fatalf("rev-parse HEAD = %q, want 40 characters", want)
		}

		if commits[0][testSHAKey] != want {
			t.Errorf("sha = %v, want %s", commits[0][testSHAKey], want)
		}

		if commits[0][testBranchKey] != "main" {
			t.Errorf("branch = %v, want main", commits[0][testBranchKey])
		}

		if _, err := time.Parse(time.RFC3339, commits[0][testAtKey].(string)); err != nil {
			t.Errorf("at = %v: %v", commits[0][testAtKey], err)
		}
	})

	t.Run("a folder outside git records no commit", func(t *testing.T) {
		gittest.Isolate(t)

		initProject(t, "BIT")
		createTask(t, "Track", "## Why\n\nA track.\n")

		out := mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)

		if commits := recordCommits(t, strings.TrimSpace(out)); len(commits) != 0 {
			t.Errorf("commits = %v, want none", commits)
		}
	})
}

func notePath(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join(storeDir(t), "feedback", name) + "\n"
}
