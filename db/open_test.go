package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_MigratesAFreshDatabase(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := Open()
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	defer sqlDB.Close()

	path := filepath.Join(home, ".local", "share", "bit-pro", "bit.db")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("os.Stat(%q) returned error: %v", path, err)
	}

	var table string

	tableQuery := "SELECT name FROM sqlite_master WHERE type='table' AND name='projects'"
	if err := sqlDB.QueryRow(tableQuery).Scan(&table); err != nil {
		t.Fatalf("querying for the projects table: %v", err)
	}

	if table != "projects" {
		t.Errorf("table = %q, want %q", table, "projects")
	}

	var applied int
	if err := sqlDB.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&applied); err != nil {
		t.Fatalf("counting applied migrations: %v", err)
	}

	if applied != 1 {
		t.Errorf("applied migrations = %d, want 1", applied)
	}
}
