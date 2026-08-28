package daemon

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

func idleRunner() claude.DirRunner {
	return func(_ context.Context, _, _ string, _ ...string) (string, int, error) {
		return "[]", 0, nil
	}
}

func TestTick_WritesProjectCounts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{ID: "ACME-1", Title: "a track", Status: task.StatusTodo}); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	queries := orm.New(sqlDB)

	if err := queries.CreateProject(t.Context(), orm.CreateProjectParams{Path: dir, Code: "ACME"}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	Tick(t.Context(), queries, log, idleRunner(), "claude")

	var backlog, todo, done, completed int64
	if err := sqlDB.QueryRowContext(t.Context(),
		"SELECT backlog, todo, done, completed FROM projects WHERE code = ?", "ACME",
	).Scan(&backlog, &todo, &done, &completed); err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("project row not found")
		}

		t.Fatalf("scanning counts returned error: %v", err)
	}

	if backlog != 1 || todo != 0 || done != 0 || completed != 0 {
		t.Errorf("counts = (%d, %d, %d, %d), want (1, 0, 0, 0)", backlog, todo, done, completed)
	}
}

func TestLoop_LogsStartedAndStoppedAroundItsTicks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	queries := orm.New(sqlDB)

	var buf bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	if err := Loop(ctx, queries, log, 5*time.Millisecond, idleRunner(), "claude"); err != nil {
		t.Fatalf("Loop() returned error: %v", err)
	}

	msgs := logMessages(t, buf.String())

	if len(msgs) < 4 {
		t.Fatalf("Loop() logged %v, want started, at least 2 ticks, and stopped", msgs)
	}

	if msgs[0] != msgStarted {
		t.Errorf("Loop() logged %q first, want %s", msgs[0], msgStarted)
	}

	if msgs[len(msgs)-1] != msgStopped {
		t.Errorf("Loop() logged %q last, want %s", msgs[len(msgs)-1], msgStopped)
	}

	var ticks int

	for _, msg := range msgs[1 : len(msgs)-1] {
		if msg == "tick" {
			ticks++
		}
	}

	if ticks < 2 {
		t.Errorf("Loop() logged %d ticks, want at least 2:\n%s", ticks, buf.String())
	}
}

func TestLoop_ReturnsWithoutTickingWhenTheContextIsAlreadyCancelled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	queries := orm.New(sqlDB)

	var buf bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := Loop(ctx, queries, log, 5*time.Millisecond, idleRunner(), "claude"); err != nil {
		t.Fatalf("Loop() returned error: %v", err)
	}

	want := []string{msgStarted, msgStopped}
	if got := logMessages(t, buf.String()); !slices.Equal(got, want) {
		t.Errorf("Loop() logged %v, want %v:\n%s", got, want, buf.String())
	}
}

func logMessages(t *testing.T, out string) []string {
	t.Helper()

	var msgs []string

	for line := range strings.SplitSeq(out, "\n") {
		if line == "" {
			continue
		}

		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("line %q is not JSON: %v", line, err)
		}

		msgs = append(msgs, record["msg"].(string))
	}

	return msgs
}

const (
	bgFlag         = "--bg"
	queuedBarID    = "ACME-1.1"
	queuedBarTitle = "a bar"
	queuedBarTree  = "ACME-1-a-track"
)

func queuedBar(t *testing.T) (*orm.Queries, orm.GetProjectByPathRow) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{ID: "ACME-1", Title: "a track", Status: task.StatusTodo, Approved: true}); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	bar := &task.Task{ID: queuedBarID, Title: queuedBarTitle, Status: task.StatusTodo, Approved: true}
	if err := store.Save(bar); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	t.Cleanup(func() { sqlDB.Close() })

	queries := orm.New(sqlDB)

	if err := queries.CreateProject(t.Context(), orm.CreateProjectParams{Path: dir, Code: "ACME"}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	project, err := queries.GetProjectByPath(t.Context(), dir)
	if err != nil {
		t.Fatalf("GetProjectByPath() returned error: %v", err)
	}

	if err := queries.EnqueueTask(t.Context(), orm.EnqueueTaskParams{
		ProjectID: project.ID,
		TargetID:  queuedBarID,
		TargetTyp: "bar",
	}); err != nil {
		t.Fatalf("EnqueueTask() returned error: %v", err)
	}

	return queries, project
}

func TestTick_DispatchesTheQueuedBar(t *testing.T) {
	queries, project := queuedBar(t)

	dir := project.Path

	calls, run := recordingRunner()

	Tick(t.Context(), queries, slog.New(slog.NewJSONHandler(io.Discard, nil)), run, "claude")

	var spawns []call

	for _, c := range *calls {
		if len(c.args) > 0 && c.args[0] == bgFlag {
			spawns = append(spawns, c)
		}
	}

	if len(spawns) != 1 {
		t.Fatalf("Tick() made %d --bg calls, want 1: %+v", len(spawns), *calls)
	}

	wantArgs := []string{
		"--bg", "--agent", "bit:bot-dev",
		"-w", queuedBarTree, "-n", queuedBarTree,
		"/bit:do " + queuedBarID,
	}

	if spawns[0].dir != dir {
		t.Errorf("Tick() spawned in %q, want %q", spawns[0].dir, dir)
	}

	if spawns[0].name != "claude" {
		t.Errorf("Tick() spawned %q, want %q", spawns[0].name, "claude")
	}

	if !slices.Equal(spawns[0].args, wantArgs) {
		t.Errorf("Tick() spawned with args %q, want %q", spawns[0].args, wantArgs)
	}
}

func TestTick_SpawnsWithTheBinaryItIsGiven(t *testing.T) {
	const bin = "/opt/homebrew/bin/claude"

	queries, _ := queuedBar(t)

	calls, run := recordingRunner()

	Tick(t.Context(), queries, slog.New(slog.NewJSONHandler(io.Discard, nil)), run, bin)

	var spawns []call

	for _, c := range *calls {
		if len(c.args) > 0 && c.args[0] == bgFlag {
			spawns = append(spawns, c)
		}
	}

	if len(spawns) != 1 {
		t.Fatalf("Tick() made %d %s calls, want 1: %+v", len(spawns), bgFlag, *calls)
	}

	wantArgs := []string{
		"--bg", "--agent", "bit:bot-dev",
		"-w", queuedBarTree, "-n", queuedBarTree,
		"/bit:do " + queuedBarID,
	}

	if spawns[0].name != bin {
		t.Errorf("Tick() spawned %q, want %q", spawns[0].name, bin)
	}

	if !slices.Equal(spawns[0].args, wantArgs) {
		t.Errorf("Tick() spawned with args %q, want %q", spawns[0].args, wantArgs)
	}
}

func TestTick_DequeuesAConfirmedDispatch(t *testing.T) {
	queries, project := queuedBar(t)

	run := func(_ context.Context, _, _ string, args ...string) (string, int, error) {
		if len(args) > 0 && args[0] == bgFlag {
			return "", 0, nil
		}

		return `[{"name":"ACME-1-a-track","cwd":"/somewhere/else"}]`, 0, nil
	}

	Tick(t.Context(), queries, slog.New(slog.NewJSONHandler(io.Discard, nil)), run, "claude")

	rows, err := queries.ListQueueByProject(t.Context(), project.ID)
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("Tick() left %d queue rows, want 0: %+v", len(rows), rows)
	}
}

func TestTick_KeepsTheRowWhenTheSessionCannotBeConfirmed(t *testing.T) {
	queries, project := queuedBar(t)

	var buf bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	Tick(t.Context(), queries, log, idleRunner(), "claude")

	rows, err := queries.ListQueueByProject(t.Context(), project.ID)
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Tick() left %d queue rows, want 1: %+v", len(rows), rows)
	}

	if rows[0].TargetID != queuedBarID {
		t.Errorf("Tick() left a row for %q, want %q", rows[0].TargetID, queuedBarID)
	}

	if !loggedAbout(t, buf.String(), "WARN", queuedBarID) {
		t.Errorf("Tick() logged no warning naming ACME-1.1:\n%s", buf.String())
	}
}

func TestTick_DropsARowItMustNotDispatch(t *testing.T) {
	tests := []struct {
		name string
		bar  task.Task
	}{
		{
			name: "already done",
			bar:  task.Task{ID: queuedBarID, Title: queuedBarTitle, Status: task.StatusDone, Approved: true},
		},
		{
			name: "approval revoked",
			bar:  task.Task{ID: queuedBarID, Title: queuedBarTitle, Status: task.StatusTodo, Approved: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries, project := queuedBar(t)

			store := task.New(filepath.Join(project.Path, ".bit"))
			if err := store.Save(&tt.bar); err != nil {
				t.Fatalf("Save() returned error: %v", err)
			}

			var buf bytes.Buffer

			log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

			calls, run := recordingRunner()

			Tick(t.Context(), queries, log, run, "claude")

			for _, c := range *calls {
				if len(c.args) > 0 && c.args[0] == bgFlag {
					t.Errorf("Tick() spawned with args %q, want no %s call", c.args, bgFlag)
				}
			}

			rows, err := queries.ListQueueByProject(t.Context(), project.ID)
			if err != nil {
				t.Fatalf("ListQueueByProject() returned error: %v", err)
			}

			if len(rows) != 0 {
				t.Errorf("Tick() left %d queue rows, want 0: %+v", len(rows), rows)
			}

			if !loggedAbout(t, buf.String(), "INFO", queuedBarID) {
				t.Errorf("Tick() logged nothing naming %s:\n%s", queuedBarID, buf.String())
			}
		})
	}
}

func TestTick_HoldsAProjectThatHasALiveSession(t *testing.T) {
	queries, project := queuedBar(t)

	live := `[{"name":"6a4a7973","cwd":"` + filepath.Join(project.Path, "cmd") + `"}]`

	var calls []call

	run := func(_ context.Context, dir, name string, args ...string) (string, int, error) {
		calls = append(calls, call{dir: dir, name: name, args: args})

		return live, 0, nil
	}

	Tick(t.Context(), queries, slog.New(slog.NewJSONHandler(io.Discard, nil)), run, "claude")

	for _, c := range calls {
		if len(c.args) > 0 && c.args[0] == bgFlag {
			t.Errorf("Tick() spawned with args %q, want no %s call", c.args, bgFlag)
		}
	}

	rows, err := queries.ListQueueByProject(t.Context(), project.ID)
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Tick() left %d queue rows, want 1: %+v", len(rows), rows)
	}
}

func loggedAbout(t *testing.T, out, level, bar string) bool {
	t.Helper()

	for line := range strings.SplitSeq(out, "\n") {
		if line == "" {
			continue
		}

		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("line %q is not JSON: %v", line, err)
		}

		if record["level"] == level && record["bar"] == bar {
			return true
		}
	}

	return false
}

type call struct {
	dir  string
	name string
	args []string
}

func recordingRunner() (*[]call, claude.DirRunner) {
	var calls []call

	return &calls, func(_ context.Context, dir, name string, args ...string) (string, int, error) {
		calls = append(calls, call{dir: dir, name: name, args: args})

		return "[]", 0, nil
	}
}

type spawned struct {
	dir  string
	tree string
	bar  string
}

type fakeSessions struct {
	live   map[string]string
	spawns []spawned
}

func (f *fakeSessions) run(_ context.Context, dir, _ string, args ...string) (string, int, error) {
	if len(args) == 0 || args[0] != bgFlag {
		agents := make([]claude.Agent, 0, len(f.live))
		for name, cwd := range f.live {
			agents = append(agents, claude.Agent{Name: name, Cwd: cwd})
		}

		out, err := json.Marshal(agents)
		if err != nil {
			return "", 0, fmt.Errorf("marshalling live sessions: %w", err)
		}

		return string(out), 0, nil
	}

	tree := flagValue(args, "-n")
	f.live[tree] = dir
	f.spawns = append(f.spawns, spawned{
		dir:  dir,
		tree: tree,
		bar:  strings.TrimPrefix(args[len(args)-1], "/bit:do "),
	})

	return "", 0, nil
}

func (f *fakeSessions) take() []spawned {
	out := f.spawns
	f.spawns = nil

	return out
}

func flagValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}

	return ""
}

func enrolled(t *testing.T, queries *orm.Queries, code, title string, bars int) orm.GetProjectByPathRow {
	t.Helper()

	dir := t.TempDir()
	store := task.New(filepath.Join(dir, ".bit"))
	track := code + "-1"

	if err := store.Save(&task.Task{ID: track, Title: title, Status: task.StatusTodo, Approved: true}); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	if err := queries.CreateProject(t.Context(), orm.CreateProjectParams{Path: dir, Code: code}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	project, err := queries.GetProjectByPath(t.Context(), dir)
	if err != nil {
		t.Fatalf("GetProjectByPath() returned error: %v", err)
	}

	for i := 1; i <= bars; i++ {
		bar := fmt.Sprintf("%s.%d", track, i)

		if err := store.Save(&task.Task{
			ID:       bar,
			Title:    fmt.Sprintf("bar %d", i),
			Status:   task.StatusTodo,
			Approved: true,
		}); err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		if err := queries.EnqueueTask(t.Context(), orm.EnqueueTaskParams{
			ProjectID: project.ID,
			TargetID:  bar,
			TargetTyp: "bar",
		}); err != nil {
			t.Fatalf("EnqueueTask() returned error: %v", err)
		}
	}

	return project
}

func TestTick_DrainsATrackOneBarPerTick(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	queries := orm.New(sqlDB)

	const treeA, treeB = "ACME-1-a-track", "ZULU-1-other-track"

	projectA := enrolled(t, queries, "ACME", "a track", 3)
	projectB := enrolled(t, queries, "ZULU", "other track", 1)

	sessions := &fakeSessions{live: map[string]string{}}
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	ticks := []struct {
		name string
		want []spawned
	}{
		{
			name: "tick 1",
			want: []spawned{
				{dir: projectA.Path, tree: treeA, bar: "ACME-1.1"},
				{dir: projectB.Path, tree: treeB, bar: "ZULU-1.1"},
			},
		},
		{name: "tick 2", want: []spawned{{dir: projectA.Path, tree: treeA, bar: "ACME-1.2"}}},
		{name: "tick 3", want: []spawned{{dir: projectA.Path, tree: treeA, bar: "ACME-1.3"}}},
	}

	for _, tt := range ticks {
		Tick(t.Context(), queries, log, sessions.run, "claude")

		if got := sessions.take(); !slices.Equal(got, tt.want) {
			t.Fatalf("%s spawned %+v, want %+v", tt.name, got, tt.want)
		}

		delete(sessions.live, treeA)
	}

	for _, project := range []orm.GetProjectByPathRow{projectA, projectB} {
		rows, err := queries.ListQueueByProject(t.Context(), project.ID)
		if err != nil {
			t.Fatalf("ListQueueByProject() returned error: %v", err)
		}

		if len(rows) != 0 {
			t.Errorf("%s kept %d queue rows, want 0: %+v", project.Code, len(rows), rows)
		}
	}
}
