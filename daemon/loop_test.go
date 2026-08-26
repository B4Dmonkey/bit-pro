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

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

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

	Tick(t.Context(), queries, log)

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

	if err := Loop(ctx, queries, log, 5*time.Millisecond); err != nil {
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

	if err := Loop(ctx, queries, log, 5*time.Millisecond); err != nil {
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
