package db

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestOpen(t *testing.T) {
	t.Run("migrates a fresh database", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("XDG_DATA_HOME", "")

		sqlDB, err := Open()
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		path := filepath.Join(home, ".local", "share", "bit", "main.db")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("os.Stat(%q) returned error: %v", path, err)
		}

		v1Dir := filepath.Join(home, ".local", "share", "bit-pro")
		if _, err := os.Stat(v1Dir); !os.IsNotExist(err) {
			t.Errorf("os.Stat(%q) error = %v, want not exist", v1Dir, err)
		}

		rows, err := sqlDB.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
		if err != nil {
			t.Fatalf("listing tables: %v", err)
		}
		defer rows.Close()

		var tables []string

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatalf("scanning a table name: %v", err)
			}

			tables = append(tables, name)
		}

		if err := rows.Err(); err != nil {
			t.Fatalf("iterating tables: %v", err)
		}

		if want := []string{"projects", "schema_migrations"}; !slices.Equal(tables, want) {
			t.Errorf("tables = %v, want %v", tables, want)
		}

		var applied int
		if err := sqlDB.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&applied); err != nil {
			t.Fatalf("counting applied migrations: %v", err)
		}

		if applied != 1 {
			t.Errorf("applied migrations = %d, want 1", applied)
		}
	})
}
