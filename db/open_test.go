package db

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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

	t.Run("concurrent first opens migrate once", func(t *testing.T) {
		dataHome := t.TempDir()
		t.Setenv("XDG_DATA_HOME", dataHome)

		const openers = 8

		testBinary, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable() returned error: %v", err)
		}

		cmds := make([]*exec.Cmd, openers)
		outputs := make([]*bytes.Buffer, openers)

		for i := range openers {
			cmd := exec.Command(testBinary, "-test.run=^TestHelperProcessOpen$")

			cmd.Env = append(os.Environ(), "BIT_OPEN_HELPER=1", "XDG_DATA_HOME="+dataHome)
			outputs[i] = &bytes.Buffer{}
			cmd.Stdout = outputs[i]
			cmd.Stderr = outputs[i]
			cmds[i] = cmd
		}

		for i, cmd := range cmds {
			if err := cmd.Start(); err != nil {
				t.Fatalf("starting opener %d: %v", i, err)
			}
		}

		for i, cmd := range cmds {
			if err := cmd.Wait(); err != nil {
				t.Errorf("opener %d: %v\n%s", i, err, outputs[i].String())
			}
		}

		sqlDB, err := Open()
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		var applied int
		if err := sqlDB.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&applied); err != nil {
			t.Fatalf("counting applied migrations: %v", err)
		}

		if applied != 1 {
			t.Errorf("applied migrations = %d, want 1", applied)
		}
	})
}

func TestHelperProcessOpen(t *testing.T) {
	if os.Getenv("BIT_OPEN_HELPER") != "1" {
		return
	}

	sqlDB, err := Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open() returned error: %v\n", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	var count int
	if err := sqlDB.QueryRow("SELECT count(*) FROM projects").Scan(&count); err != nil {
		fmt.Fprintf(os.Stderr, "counting projects: %v\n", err)
		os.Exit(1)
	}
}
