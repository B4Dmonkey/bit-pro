package daemon

import (
	"database/sql"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

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
