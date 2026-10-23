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

		if want := []string{"model_prices", "projects", "schema_migrations", "usage"}; !slices.Equal(tables, want) {
			t.Errorf("tables = %v, want %v", tables, want)
		}

		var applied int
		if err := sqlDB.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&applied); err != nil {
			t.Fatalf("counting applied migrations: %v", err)
		}

		if applied != 3 {
			t.Errorf("applied migrations = %d, want 3", applied)
		}
	})

	t.Run("connection waits on locks and enforces foreign keys", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		sqlDB, err := Open()
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		tests := []struct {
			pragma string
			want   string
		}{
			{"busy_timeout", "5000"},
			{"journal_mode", "wal"},
			{"foreign_keys", "1"},
		}

		for _, tt := range tests {
			var got string
			if err := sqlDB.QueryRow("PRAGMA " + tt.pragma).Scan(&got); err != nil {
				t.Fatalf("PRAGMA %s returned error: %v", tt.pragma, err)
			}

			if got != tt.want {
				t.Errorf("PRAGMA %s = %q, want %q", tt.pragma, got, tt.want)
			}
		}
	})
}
