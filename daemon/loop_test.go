package daemon

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

	Tick(t.Context(), queries, log, idleRunner())

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

	if err := Loop(ctx, queries, log, 5*time.Millisecond, idleRunner()); err != nil {
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

	if err := Loop(ctx, queries, log, 5*time.Millisecond, idleRunner()); err != nil {
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
	bgFlag      = "--bg"
	queuedBarID = "ACME-1.1"
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

	bar := &task.Task{ID: queuedBarID, Title: "a bar", Status: task.StatusTodo, Approved: true}
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

	Tick(t.Context(), queries, slog.New(slog.NewJSONHandler(io.Discard, nil)), run)

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
		"-w", "ACME-1-a-track", "-n", "ACME-1-a-track",
		"/bit:do ACME-1.1",
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

func TestTick_DequeuesAConfirmedDispatch(t *testing.T) {
	queries, project := queuedBar(t)

	run := func(_ context.Context, _, _ string, args ...string) (string, int, error) {
		if len(args) > 0 && args[0] == bgFlag {
			return "", 0, nil
		}

		return `[{"name":"ACME-1-a-track","cwd":"/somewhere/else"}]`, 0, nil
	}

	Tick(t.Context(), queries, slog.New(slog.NewJSONHandler(io.Discard, nil)), run)

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

	Tick(t.Context(), queries, log, idleRunner())

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

	if !warnedAbout(t, buf.String(), queuedBarID) {
		t.Errorf("Tick() logged no warning naming ACME-1.1:\n%s", buf.String())
	}
}

func warnedAbout(t *testing.T, out, bar string) bool {
	t.Helper()

	for line := range strings.SplitSeq(out, "\n") {
		if line == "" {
			continue
		}

		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("line %q is not JSON: %v", line, err)
		}

		if record["level"] == "WARN" && record["bar"] == bar {
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
