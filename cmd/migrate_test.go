package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/git/gittest"
	"github.com/B4Dmonkey/bit-pro/migrate"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	migrateCmdUse = "migrate"
	migratedBar   = "BIT-1.1"
	splitBar      = "BIT-39.1"
	listedBar     = "BIT-10.1"
	activeTitle   = "Active"
	notedTrack    = "BIT-19"
	activeTrack   = "BIT-44"
	researchTrack = "BIT-49"
)

func TestMigrateCmd(t *testing.T) {
	t.Run("copies active tasks and registers the folder", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

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
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		if id := createTask(t, "Next", ""); id != testOwnTrack2 {
			t.Errorf("created ID = %q, want %q", id, testOwnTrack2)
		}
	})

	t.Run("keeps completed and archived records where they were", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Files(t, map[string][]*task.Task{
			testTasksDir: {{ID: "BIT-40", Title: activeTitle, Status: task.StatusTodo}},
			"completed": {
				{ID: "BIT-39", Title: "Split", Status: task.StatusDoing, Order: []string{splitBar}},
				{ID: splitBar, Title: "Done bar", Status: task.StatusDone},
				{ID: "BIT-10", Title: "Partial", Status: task.StatusDone, Order: []string{listedBar}},
				{ID: listedBar, Title: "Listed", Status: task.StatusDone},
				{ID: "BIT-10.9", Title: "Unlisted", Status: task.StatusDone},
			},
			filepath.Join("archive", testTasksDir): {{ID: "BIT-39.13", Title: "Deleted bar", Status: task.StatusTodo}},
		}))
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		root, err := store.ProjectDir(testPrefix)
		if err != nil {
			t.Fatalf("store.ProjectDir(%q) returned error: %v", testPrefix, err)
		}

		split := readMigratedRecord(t, filepath.Join(root, "completed", "BIT-39.json"))
		if split.Status != task.StatusDoing || !slices.Equal(split.Order, []string{splitBar}) {
			t.Errorf("completed/BIT-39 = {%q, %v}, want {%q, [BIT-39.1]}", split.Status, split.Order, task.StatusDoing)
		}

		if _, err := os.Stat(filepath.Join(root, "archive", "tasks", "BIT-39.13.json")); err != nil {
			t.Errorf("archive/tasks/BIT-39.13.json: %v", err)
		}

		if _, err := os.Stat(filepath.Join(root, "completed", "BIT-39.13.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("completed/BIT-39.13.json stat error = %v, want not exist", err)
		}

		partial := readMigratedRecord(t, filepath.Join(root, "completed", "BIT-10.json"))
		if !slices.Equal(partial.Order, []string{listedBar}) {
			t.Errorf("completed/BIT-10 order = %v, want [BIT-10.1]", partial.Order)
		}

		if _, err := os.Stat(filepath.Join(root, "completed", "BIT-10.9.json")); err != nil {
			t.Errorf("completed/BIT-10.9.json: %v", err)
		}

		out := mustRun(t, "task", "list")
		if !strings.Contains(out, "BIT-40") || strings.Count(out, "BIT-") != 1 {
			t.Errorf("bp task list = %q, want only BIT-40", out)
		}
	})

	t.Run("an archived track's id is never re-minted", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Files(t, map[string][]*task.Task{
			testTasksDir:                           {{ID: testOwnTrack2, Title: activeTitle, Status: task.StatusTodo}},
			filepath.Join("archive", testTasksDir): {{ID: "BIT-7", Title: "Deleted", Status: task.StatusTodo}},
		}))
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		if id := createTask(t, "T", ""); id != "BIT-8" {
			t.Errorf("created ID = %q, want %q", id, "BIT-8")
		}
	})

	t.Run("moves feedback notes with their numbers", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Files(t, map[string][]*task.Task{
			testTasksDir:                           {{ID: activeTrack, Title: activeTitle, Status: task.StatusDoing}},
			filepath.Join("archive", testTasksDir): {{ID: notedTrack, Title: "Deleted", Status: task.StatusTodo}},
		})
		notes := map[string]struct {
			track string
			seq   float64
			body  []byte
		}{
			"BIT-19-001.md": {notedTrack, 1, []byte("## What the plan said\n\nUse a map.\n\n## What happened\n\nA slice.\n")},
			"BIT-19-002.md": {notedTrack, 2, []byte("## What the plan said\n\nNo flag.\n\n## What happened\n\nA --dry-run.\n")},
			"BIT-44-005.md": {activeTrack, 5, []byte("## What the plan said\n\nOne bar.\n\n## What happened\n\nIt took two.\n")},
		}

		for name, n := range notes {
			files[filepath.Join("feedback", name)] = n.body
		}

		writeV1Store(t, dir, files)
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		d := dataDir(t)

		for name, n := range notes {
			path := filepath.Join(d, "feedback", name)

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
			}

			if string(got) != string(n.body) {
				t.Errorf("%s = %q, want %q", name, got, n.body)
			}

			rec := readRecord(t, path)
			if rec["project"] != testPrefix || rec["track"] != n.track || rec["seq"] != n.seq {
				t.Errorf("%s record = {%v, %v, %v}, want {%s, %s, %v}",
					name, rec["project"], rec["track"], rec["seq"], testPrefix, n.track, n.seq)
			}
		}

		path, err := projectStore(t).AddNote(notedTrack, "next", task.Commit{})
		if err != nil {
			t.Fatalf("AddNote(%s) error = %v", notedTrack, err)
		}

		if !strings.HasSuffix(path, "BIT-19-003.md") {
			t.Errorf("AddNote path = %q, want a path ending in BIT-19-003.md", path)
		}

		if _, err := os.Stat(filepath.Join(d, testPrefix, "feedback")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat project feedback dir = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("carries research topics", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Files(t, map[string][]*task.Task{
			testTasksDir: {{ID: researchTrack, Title: activeTitle, Status: task.StatusDoing}},
			"completed":  {{ID: activeTrack, Title: "Done", Status: task.StatusDone}},
		})
		indexBody := []byte("## Findings\n\n| a | b |\n|---|---|\n| 1 | 2 |\n\nSee [x](x.md).\n")
		files[filepath.Join(testResearchDir, researchTrack, "index.md")] = indexBody
		files[filepath.Join(testResearchDir, researchTrack, "claim-audit-2026-10-02.md")] = []byte("audit\n")
		files[filepath.Join(testResearchDir, activeTrack, "index.md")] = []byte("done track\n")
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		session := mcpSession(t, dir)

		got := callTool(t, session, researchReadTool, map[string]any{testTrackKey: researchTrack})

		want := []any{"claim-audit-2026-10-02", testIndexTopic}
		if !reflect.DeepEqual(got[testTopicsKey], want) {
			t.Errorf("topics = %v, want %v", got[testTopicsKey], want)
		}

		got = callTool(t, session, researchReadTool, map[string]any{
			testTrackKey: researchTrack,
			testTopicKey: testIndexTopic,
		})
		if got[testBodyKey] != string(indexBody) {
			t.Errorf("body = %q, want %q", got[testBodyKey], indexBody)
		}

		root, err := store.ProjectDir(testPrefix)
		if err != nil {
			t.Fatalf("store.ProjectDir(%q) returned error: %v", testPrefix, err)
		}

		rec := readRecord(t, filepath.Join(root, testResearchDir, activeTrack, "index.md"))
		if rec["project"] != testPrefix || rec["track"] != activeTrack {
			t.Errorf("BIT-44 index record = {%v, %v}, want {%s, BIT-44}", rec["project"], rec["track"], testPrefix)
		}
	})

	t.Run("stores retro proposals under the prefix rule", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Files(t, map[string][]*task.Task{
			testTasksDir: {{ID: "BIT-12", Title: activeTitle, Status: task.StatusDoing}},
		})
		albumBody := []byte("## Proposal 1\n\nAlbum.\n")
		files[filepath.Join("retro", "album-proposals.md")] = albumBody
		files[filepath.Join("retro", "BIT-12-proposals.md")] = []byte("## Proposal 1\n\nTrack.\n")
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		var got struct {
			Proposals []map[string]string `json:"proposals"`
		}

		decodeToolResult(t, mcpSession(t, dir), retroListTool, map[string]any{}, &got)

		want := []map[string]string{
			{testNameKey: "BIT-12-proposals", testProjectKey: testPrefix},
			{testNameKey: "BIT-album-proposals", testProjectKey: testPrefix},
		}
		if !reflect.DeepEqual(got.Proposals, want) {
			t.Errorf("proposals = %v, want %v", got.Proposals, want)
		}

		path := filepath.Join(dataDir(t), "retro", "BIT-album-proposals.md")

		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
		}

		if string(body) != string(albumBody) {
			t.Errorf("BIT-album-proposals.md = %q, want %q", body, albumBody)
		}
	})

	t.Run("migrates a folder outside git", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Fixture(t)
		files[filepath.Join("feedback", "BIT-1-001.md")] = []byte("## What happened\n\nA note.\n")
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		root, err := store.ProjectDir(testPrefix)
		if err != nil {
			t.Fatalf("store.ProjectDir(%q) returned error: %v", testPrefix, err)
		}

		rec := readRecord(t, filepath.Join(root, testTasksDir, testOwnTrack+".md"))
		if rec["branch"] != "" || rec["commit"] != "" {
			t.Errorf("BIT-1 = {branch %v, commit %v}, want empty", rec["branch"], rec["commit"])
		}

		note := readRecord(t, filepath.Join(dataDir(t), "feedback", "BIT-1-001.md"))
		if !reflect.DeepEqual(note["commits"], []any{}) {
			t.Errorf("BIT-1-001 commits = %v, want []", note["commits"])
		}
	})

	t.Run("a copy that doesn't match leaves nothing behind", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Files(t, map[string][]*task.Task{
			testTasksDir: {{ID: testOwnTrack, Title: activeTitle, Status: task.StatusDoing}},
		})
		files[filepath.Join("feedback", "BIT-1-1.md")] = []byte("## What happened\n\nHand-numbered.\n")
		writeV1Store(t, dir, files)
		before := hashV1Store(t, dir, files)
		t.Chdir(dir)

		_, err := run(t, migrateCmdUse)
		if !errors.Is(err, migrate.ErrVerify) {
			t.Fatalf("bp migrate error = %v, want migrate.ErrVerify", err)
		}

		if !strings.Contains(err.Error(), "feedback/BIT-1-1.md") {
			t.Errorf("bp migrate error = %q, want it to name feedback/BIT-1-1.md", err)
		}

		if projects := listProjects(t); len(projects) != 0 {
			t.Errorf("ListProjects() = %v, want none", projects)
		}

		d := dataDir(t)

		if _, err := os.Stat(filepath.Join(d, testPrefix)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat %s project dir = %v, want fs.ErrNotExist", testPrefix, err)
		}

		if notes, _ := filepath.Glob(filepath.Join(d, "feedback", testPrefix+"-*")); len(notes) != 0 {
			t.Errorf("feedback = %v, want no %s notes", notes, testPrefix)
		}

		if stages, _ := filepath.Glob(filepath.Join(d, ".migrate-*")); len(stages) != 0 {
			t.Errorf("staging left behind: %v", stages)
		}

		if after := hashV1Store(t, dir, files); !slices.Equal(after, before) {
			t.Errorf("source hashes changed: before %x, after %x", before, after)
		}
	})

	t.Run("stops on unknown files", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Fixture(t)
		files[filepath.Join("archive", "BIT-3.md")] = []byte("flat archive\n")
		files["notes.txt"] = []byte("notes\n")
		files[filepath.Join(testResearchDir, testOwnTrack, "diagram.png")] = []byte("png\n")
		files[".DS_Store"] = []byte("finder\n")
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		_, err := run(t, migrateCmdUse)
		if !errors.Is(err, migrate.ErrUnknownFiles) {
			t.Fatalf("bp migrate error = %v, want migrate.ErrUnknownFiles", err)
		}

		want := ".DS_Store\n  archive/BIT-3.md\n  notes.txt\n  research/BIT-1/diagram.png"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("bp migrate error = %q, want it to list %q", err, want)
		}

		if projects := listProjects(t); len(projects) != 0 {
			t.Errorf("ListProjects() = %v, want none", projects)
		}

		d := dataDir(t)

		if _, err := os.Stat(filepath.Join(d, testPrefix)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat %s project dir = %v, want fs.ErrNotExist", testPrefix, err)
		}

		if stages, _ := filepath.Glob(filepath.Join(d, ".migrate-*")); len(stages) != 0 {
			t.Errorf("staging left behind: %v", stages)
		}
	})

	t.Run("a re-run says already migrated and changes nothing", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		writeV1Store(t, dir, map[string][]byte{"notes.txt": []byte("notes\n")})

		d := dataDir(t)
		before := snapshotData(t, d)

		out, err := run(t, migrateCmdUse)
		if err != nil {
			t.Fatalf("second bp migrate returned error: %v", err)
		}

		if want := "already migrated\n"; out != want {
			t.Errorf("output = %q, want %q", out, want)
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Errorf("ListProjects() returned %d projects, want 1", len(projects))
		}

		if after := snapshotData(t, d); !reflect.DeepEqual(after, before) {
			t.Errorf("data dir changed:\nbefore %v\nafter  %v", before, after)
		}
	})

	t.Run("refuses a removed project's folder", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)
		markRemoved(t, "BIT")

		d := dataDir(t)
		before := snapshotData(t, d)

		_, err := run(t, migrateCmdUse)
		if !errors.Is(err, project.ErrRemoved) {
			t.Fatalf("second bp migrate error = %v, want %v", err, project.ErrRemoved)
		}

		if !strings.Contains(err.Error(), "bp add") {
			t.Errorf("error %q does not mention bp add", err)
		}

		projects := listProjects(t)
		if len(projects) != 1 || projects[0].Removed == 0 {
			t.Errorf("ListProjects() = %+v, want one removed row", projects)
		}

		if after := snapshotData(t, d); !reflect.DeepEqual(after, before) {
			t.Errorf("data dir changed:\nbefore %v\nafter  %v", before, after)
		}
	})

	t.Run("stops on task files it can't copy exactly", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Fixture(t)

		extra, err := (&task.Task{ID: testOwnTrack2, Title: "Extra", Status: task.StatusTodo}).Bytes()
		if err != nil {
			t.Fatalf("Bytes(BIT-2) returned error: %v", err)
		}

		extra = []byte(strings.Replace(string(extra), "---\n", "---\npriority: high\n", 1))
		files[filepath.Join(testTasksDir, "BIT-2.md")] = extra

		crlf, err := (&task.Task{ID: "BIT-3", Title: "Windows", Status: task.StatusDone}).Bytes()
		if err != nil {
			t.Fatalf("Bytes(BIT-3) returned error: %v", err)
		}

		files[filepath.Join(testCompletedDir, "BIT-3.md")] = []byte(strings.ReplaceAll(string(crlf), "\n", "\r\n"))
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		_, err = run(t, migrateCmdUse)
		if !errors.Is(err, migrate.ErrTaskFiles) {
			t.Fatalf("bp migrate error = %v, want migrate.ErrTaskFiles", err)
		}

		for _, want := range []string{"tasks/BIT-2.md: does not round-trip", "completed/BIT-3.md"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("bp migrate error = %q, want it to name %q", err, want)
			}
		}

		if projects := listProjects(t); len(projects) != 0 {
			t.Errorf("ListProjects() = %v, want none", projects)
		}

		d := dataDir(t)

		if _, err := os.Stat(filepath.Join(d, testPrefix)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat %s project dir = %v, want fs.ErrNotExist", testPrefix, err)
		}

		if stages, _ := filepath.Glob(filepath.Join(d, ".migrate-*")); len(stages) != 0 {
			t.Errorf("staging left behind: %v", stages)
		}
	})

	t.Run("refuses ids that are not normalised", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Fixture(t)

		lower, err := (&task.Task{ID: testOwnTrack2, Title: "Lower", Status: task.StatusTodo}).Bytes()
		if err != nil {
			t.Fatalf("Bytes(BIT-2) returned error: %v", err)
		}

		files[filepath.Join(testTasksDir, "bit-2.md")] = lower
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		_, err = run(t, migrateCmdUse)
		if !errors.Is(err, migrate.ErrIDs) {
			t.Fatalf("bp migrate error = %v, want migrate.ErrIDs", err)
		}

		if !strings.Contains(err.Error(), "tasks/bit-2.md") {
			t.Errorf("bp migrate error = %q, want it to name tasks/bit-2.md", err)
		}

		if projects := listProjects(t); len(projects) != 0 {
			t.Errorf("ListProjects() = %v, want none", projects)
		}

		d := dataDir(t)

		if _, err := os.Stat(filepath.Join(d, testPrefix)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat %s project dir = %v, want fs.ErrNotExist", testPrefix, err)
		}

		if stages, _ := filepath.Glob(filepath.Join(d, ".migrate-*")); len(stages) != 0 {
			t.Errorf("staging left behind: %v", stages)
		}
	})

	t.Run("refuses a code another project holds", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		a := t.TempDir()
		writeV1Store(t, a, v1Fixture(t))
		t.Chdir(a)
		mustRun(t, migrateCmdUse)

		aPath, err := project.CanonicalPath(a)
		if err != nil {
			t.Fatalf("project.CanonicalPath(%q) returned error: %v", a, err)
		}

		b := t.TempDir()
		writeV1Store(t, b, v1Fixture(t))
		t.Chdir(b)

		d := dataDir(t)
		before := snapshotData(t, d)

		_, err = run(t, migrateCmdUse)
		if !errors.Is(err, migrate.ErrCodeTaken) {
			t.Fatalf("bp migrate error = %v, want migrate.ErrCodeTaken", err)
		}

		if !strings.Contains(err.Error(), aPath) {
			t.Errorf("bp migrate error = %q, want it to name %s", err, aPath)
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Errorf("ListProjects() returned %d projects, want 1", len(projects))
		}

		if stages, _ := filepath.Glob(filepath.Join(d, ".migrate-*")); len(stages) != 0 {
			t.Errorf("staging left behind: %v", stages)
		}

		if after := snapshotData(t, d); !reflect.DeepEqual(after, before) {
			t.Errorf("data dir changed:\nbefore %v\nafter  %v", before, after)
		}
	})

	t.Run("refuses a removed project's code", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		a := t.TempDir()
		writeV1Store(t, a, v1Fixture(t))
		t.Chdir(a)
		mustRun(t, migrateCmdUse)
		markRemoved(t, testPrefix)

		b := t.TempDir()
		writeV1Store(t, b, v1Fixture(t))
		t.Chdir(b)

		d := dataDir(t)
		before := snapshotData(t, d)

		_, err := run(t, migrateCmdUse)
		if !errors.Is(err, project.ErrCodeRemoved) {
			t.Fatalf("bp migrate error = %v, want project.ErrCodeRemoved", err)
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Errorf("ListProjects() returned %d projects, want 1", len(projects))
		}

		if after := snapshotData(t, d); !reflect.DeepEqual(after, before) {
			t.Errorf("data dir changed:\nbefore %v\nafter  %v", before, after)
		}
	})

	t.Run("refuses an invalid or reserved code", func(t *testing.T) {
		cases := []struct {
			name, prefix string
			want         error
		}{
			{name: "reserved", prefix: "FEEDBACK", want: project.ErrReservedCode},
			{name: "invalid", prefix: "BIT-PRO", want: project.ErrInvalidCode},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				mcpSandbox(t)
				gittest.Isolate(t)

				dir := t.TempDir()
				files := v1Fixture(t)
				files["config.toml"] = []byte("prefix = \"" + tc.prefix + "\"\n")
				writeV1Store(t, dir, files)
				t.Chdir(dir)

				_, err := run(t, migrateCmdUse)
				if !errors.Is(err, tc.want) {
					t.Fatalf("bp migrate error = %v, want %v", err, tc.want)
				}

				if !strings.Contains(err.Error(), "config.toml") {
					t.Errorf("bp migrate error = %q, want it to name config.toml", err)
				}

				if projects := listProjects(t); len(projects) != 0 {
					t.Errorf("ListProjects() = %v, want none", projects)
				}

				entries, err := os.ReadDir(dataDir(t))
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("os.ReadDir(data dir) returned error: %v", err)
				}

				for _, e := range entries {
					if !strings.HasPrefix(e.Name(), "main.db") {
						t.Errorf("data dir holds %s, want nothing written", e.Name())
					}
				}
			})
		}
	})

	t.Run("a note on a missing track stops the migration", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Fixture(t)
		files[filepath.Join("feedback", "BIT-9-001.md")] = []byte("## What happened\n\nOrphan.\n")
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		_, err := run(t, migrateCmdUse)
		if err == nil {
			t.Fatal("bp migrate with an orphan note returned no error")
		}

		if !strings.Contains(err.Error(), "BIT-9") {
			t.Errorf("bp migrate error = %q, want it to name BIT-9", err)
		}
	})
	t.Run("walks up from a subfolder", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))

		sub := filepath.Join(dir, "src", "pkg")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) returned error: %v", sub, err)
		}

		t.Chdir(sub)

		want, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("project.CanonicalPath(%q) returned error: %v", dir, err)
		}

		out := mustRun(t, migrateCmdUse)

		wantLine := "migrated " + testPrefix + " " + want
		if first := strings.SplitN(out, "\n", 2)[0]; first != wantLine {
			t.Errorf("first output line = %q, want %q", first, wantLine)
		}

		projects := listProjects(t)
		if len(projects) != 1 || projects[0].Path != want {
			t.Errorf("ListProjects() = %+v, want one row at %s", projects, want)
		}
	})

	t.Run("prints the cleanup step for an untracked bit folder", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Fixture(t)
		writeV1Store(t, dir, files)
		before := hashV1Store(t, dir, files)
		t.Chdir(dir)

		want, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("project.CanonicalPath(%q) returned error: %v", dir, err)
		}

		out := mustRun(t, migrateCmdUse)

		if rm := "rm -rf " + filepath.Join(want, ".bit"); !strings.Contains(out, rm) {
			t.Errorf("output = %q, want it to contain %q", out, rm)
		}

		if strings.Contains(out, "git rm") {
			t.Errorf("output = %q, want no git rm", out)
		}

		if after := hashV1Store(t, dir, files); !slices.Equal(after, before) {
			t.Errorf("source hashes changed: before %x, after %x", before, after)
		}
	})

	t.Run("prints git rm for a tracked bit folder", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		gittest.Run(t, dir, "init", "-b", "main")
		gittest.Run(t, dir, "add", ".bit")
		gittest.Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-m", "v1")
		t.Chdir(dir)

		out := mustRun(t, migrateCmdUse)

		for _, step := range []string{"git rm -r .bit", "rm -rf .bit"} {
			if !strings.Contains(out, step) {
				t.Errorf("output = %q, want it to contain %q", out, step)
			}
		}

		if status := gittest.Run(t, dir, "status", "--porcelain"); status != "" {
			t.Errorf("git status --porcelain = %q, want clean", status)
		}
	})

	t.Run("first migration ensures the global wiring", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		t.Chdir(dir)

		var calls [][]string

		out, err := runWithRunner(t, recordCalls(&calls, -1), "", migrateCmdUse)
		if err != nil {
			t.Fatalf("bp migrate returned error: %v", err)
		}

		if !strings.HasPrefix(out, "migrated "+testPrefix) {
			t.Errorf("output = %q, want it to start with %q", out, "migrated "+testPrefix)
		}

		if want := claude.GlobalWiring(); !slices.EqualFunc(calls, want, slices.Equal) {
			t.Errorf("calls = %v, want %v", calls, want)
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Errorf("ListProjects() returned %d projects, want 1", len(projects))
		}
	})

	t.Run("a re-run does not wire again", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		t.Chdir(dir)

		mustRun(t, migrateCmdUse)

		var calls [][]string

		out, err := runWithRunner(t, recordCalls(&calls, -1), "", migrateCmdUse)
		if err != nil {
			t.Fatalf("second bp migrate returned error: %v", err)
		}

		if want := "already migrated\n"; out != want {
			t.Errorf("output = %q, want %q", out, want)
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
		}
	})

	t.Run("a wiring failure exits non-zero and keeps the project", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		writeV1Store(t, dir, v1Fixture(t))
		t.Chdir(dir)

		var calls [][]string

		if _, err := runWithRunner(t, recordCalls(&calls, 2), "", migrateCmdUse); err == nil {
			t.Fatal("bp migrate with a failing wiring step returned no error")
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Errorf("ListProjects() returned %d projects, want 1", len(projects))
		}

		root, err := store.ProjectDir(testPrefix)
		if err != nil {
			t.Fatalf("store.ProjectDir(%q) returned error: %v", testPrefix, err)
		}

		if _, err := os.Stat(root); err != nil {
			t.Errorf("stat project dir: %v", err)
		}
	})

	t.Run("a refused migration does not wire", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		dir := t.TempDir()
		files := v1Fixture(t)
		files["notes.txt"] = []byte("notes\n")
		writeV1Store(t, dir, files)
		t.Chdir(dir)

		var calls [][]string

		if _, err := runWithRunner(t, recordCalls(&calls, -1), "", migrateCmdUse); !errors.Is(err, migrate.ErrUnknownFiles) {
			t.Fatalf("bp migrate error = %v, want migrate.ErrUnknownFiles", err)
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
		}
	})

	t.Run("no bit folder here or above", func(t *testing.T) {
		mcpSandbox(t)
		gittest.Isolate(t)

		t.Chdir(t.TempDir())

		if _, err := run(t, migrateCmdUse); !errors.Is(err, migrate.ErrNoBitDir) {
			t.Fatalf("bp migrate error = %v, want migrate.ErrNoBitDir", err)
		}
	})
}

func recordCalls(calls *[][]string, failAt int) claude.Runner {
	return func(_ context.Context, name string, args ...string) error {
		*calls = append(*calls, append([]string{name}, args...))
		if len(*calls)-1 == failAt {
			return errors.New("wiring step failed")
		}

		return nil
	}
}

type migratedRecord struct {
	Status string   `json:"status"`
	Order  []string `json:"order"`
}

func readMigratedRecord(t *testing.T, path string) migratedRecord {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	var rec migratedRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("json.Unmarshal(%q) returned error: %v", path, err)
	}

	return rec
}

func v1Files(t *testing.T, places map[string][]*task.Task) map[string][]byte {
	t.Helper()

	files := map[string][]byte{"config.toml": []byte("prefix = \"BIT\"\n")}

	for dir, tasks := range places {
		for _, tk := range tasks {
			data, err := tk.Bytes()
			if err != nil {
				t.Fatalf("Bytes(%s) returned error: %v", tk.ID, err)
			}

			files[filepath.Join(dir, tk.ID+".md")] = data
		}
	}

	return files
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

	return v1Files(t, map[string][]*task.Task{testTasksDir: append([]*task.Task{v1Track()}, v1Bars()...)})
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

func snapshotData(t *testing.T, dir string) map[string]time.Time {
	t.Helper()

	snap := map[string]time.Time{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(d.Name(), "main.db") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		snap[path] = info.ModTime()

		return nil
	})
	if err != nil {
		t.Fatalf("filepath.WalkDir(%q) returned error: %v", dir, err)
	}

	return snap
}
