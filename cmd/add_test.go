package cmd

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
)

func TestAddCmd(t *testing.T) {
	t.Run("refuses an unregistered folder with bit", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")
		t.Chdir(t.TempDir())

		if err := os.MkdirAll(".bit", 0o755); err != nil {
			t.Fatalf("os.MkdirAll(.bit) returned error: %v", err)
		}

		if err := os.WriteFile(filepath.Join(".bit", "config.toml"), []byte("prefix = \"BIT\"\n"), 0o600); err != nil {
			t.Fatalf("os.WriteFile(.bit/config.toml) returned error: %v", err)
		}

		var calls [][]string

		run := func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			return nil
		}

		out, err := runWithRunner(t, run, "BIT\n", addCmdUse, ".")
		if !errors.Is(err, project.ErrNeedsMigrate) {
			t.Fatalf("Execute() error = %v, want %v", err, project.ErrNeedsMigrate)
		}

		if strings.Contains(out, "Project code") {
			t.Errorf("output = %q, want no %q", out, "Project code")
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
		}

		if projects := listProjects(t); len(projects) != 0 {
			t.Errorf("ListProjects() returned %d projects, want 0", len(projects))
		}
	})

	t.Run("initialises a project without bit", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("XDG_DATA_HOME", "")
		t.Chdir(t.TempDir())

		var calls [][]string

		run := func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			return nil
		}

		want, err := project.CanonicalPath(".")
		if err != nil {
			t.Fatalf("project.CanonicalPath(.) returned error: %v", err)
		}

		out, err := runWithRunner(t, run, "BIT\n", addCmdUse, ".")
		if err != nil {
			t.Fatalf("Execute() returned error: %v", err)
		}

		wantOut := "Project code: Bringing the bit plugin current...\n" +
			"Registering bit MCP server...\n" +
			"bit MCP server registered (local scope).\n" +
			"added BIT " + want + "\n"
		if out != wantOut {
			t.Errorf("output = %q, want %q", out, wantOut)
		}

		prompt := strings.SplitN(out, "Bringing", 2)[0]
		if strings.Contains(prompt, "(") {
			t.Errorf("prompt = %q, want no %q", prompt, "(")
		}

		data, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
		if err != nil {
			t.Fatalf("os.ReadFile(.claude/settings.json) returned error: %v", err)
		}

		if !strings.Contains(string(data), "bit@bit-pro") {
			t.Errorf("settings.json = %s, want it to contain %q", data, "bit@bit-pro")
		}

		if _, err := os.Stat(filepath.Join(".bit", "config.toml")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.bit/config.toml) error = %v, want fs.ErrNotExist", err)
		}

		wantCalls := pluginSyncCalls()
		if !slices.EqualFunc(calls, wantCalls, slices.Equal) {
			t.Errorf("calls = %v, want %v", calls, wantCalls)
		}

		sqlDB, err := db.Open()
		if err != nil {
			t.Fatalf("db.Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		projects, err := orm.New(sqlDB).ListProjects(t.Context())
		if err != nil {
			t.Fatalf("ListProjects() returned error: %v", err)
		}

		if len(projects) != 1 {
			t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
		}

		if projects[0].Code != testPrefix {
			t.Errorf("Code = %q, want %q", projects[0].Code, testPrefix)
		}
	})

	t.Run("uppercases a typed code", func(t *testing.T) {
		tests := []struct {
			name  string
			typed string
		}{
			{name: "lowercase", typed: "foo"},
			{name: "uppercase", typed: testCode},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				t.Setenv("XDG_DATA_HOME", "")
				t.Chdir(t.TempDir())

				want, err := project.CanonicalPath(".")
				if err != nil {
					t.Fatalf("project.CanonicalPath(.) returned error: %v", err)
				}

				out, err := runWithStdin(t, tt.typed+"\n", addCmdUse, ".")
				if err != nil {
					t.Fatalf("Execute() returned error: %v", err)
				}

				wantOut := "Project code: Bringing the bit plugin current...\n" +
					"Registering bit MCP server...\n" +
					"bit MCP server registered (local scope).\n" +
					"added " + testCode + " " + want + "\n"
				if out != wantOut {
					t.Errorf("output = %q, want %q", out, wantOut)
				}

				projects := listProjects(t)
				if len(projects) != 1 {
					t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
				}

				if projects[0].Code != testCode {
					t.Errorf("Code = %q, want %q", projects[0].Code, testCode)
				}

				if projects[0].Path != want {
					t.Errorf("Path = %q, want %q", projects[0].Path, want)
				}
			})
		}
	})

	t.Run("rejects an empty code", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("XDG_DATA_HOME", "")
		t.Chdir(t.TempDir())

		_, err := runWithStdin(t, "\n", addCmdUse, ".")
		if !errors.Is(err, project.ErrInvalidCode) {
			t.Fatalf("Execute() error = %v, want %v", err, project.ErrInvalidCode)
		}

		if _, err := os.Stat(filepath.Join(".claude", "settings.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.claude/settings.json) error = %v, want fs.ErrNotExist", err)
		}

		sqlDB, err := db.Open()
		if err != nil {
			t.Fatalf("db.Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		projects, err := orm.New(sqlDB).ListProjects(t.Context())
		if err != nil {
			t.Fatalf("ListProjects() returned error: %v", err)
		}

		if len(projects) != 0 {
			t.Errorf("ListProjects() returned %d projects, want 0", len(projects))
		}
	})

	t.Run("refuses an invalid code", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("XDG_DATA_HOME", "")
		t.Chdir(t.TempDir())

		var calls [][]string

		run := func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			return nil
		}

		_, err := runWithRunner(t, run, "bit-pro\n", addCmdUse, ".")
		if !errors.Is(err, project.ErrInvalidCode) {
			t.Fatalf("Execute() error = %v, want %v", err, project.ErrInvalidCode)
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
		}

		if _, err := os.Stat(filepath.Join(".claude", "settings.json")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.claude/settings.json) error = %v, want fs.ErrNotExist", err)
		}

		sqlDB, err := db.Open()
		if err != nil {
			t.Fatalf("db.Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		projects, err := orm.New(sqlDB).ListProjects(t.Context())
		if err != nil {
			t.Fatalf("ListProjects() returned error: %v", err)
		}

		if len(projects) != 0 {
			t.Errorf("ListProjects() returned %d projects, want 0", len(projects))
		}
	})

	t.Run("skips a path already enrolled", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		tmp := t.TempDir()
		if err := os.Mkdir(filepath.Join(tmp, "Repo"), 0o755); err != nil {
			t.Fatalf("os.Mkdir(Repo) returned error: %v", err)
		}

		if _, err := runWithStdin(t, testCode+"\n", addCmdUse, filepath.Join(tmp, "Repo")); err != nil {
			t.Fatalf("first Execute() returned error: %v", err)
		}

		var calls [][]string

		run := func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			return nil
		}

		out, err := runWithRunner(t, run, testCode+"\n", addCmdUse, filepath.Join(tmp, "repo"))
		if err != nil {
			t.Fatalf("second Execute() returned error: %v", err)
		}

		if wantOut := "already added\n"; out != wantOut {
			t.Errorf("output = %q, want %q", out, wantOut)
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
		}
	})

	t.Run("allows a path inside a registered path", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		tmp := t.TempDir()

		if _, err := runWithStdin(t, "ACME\n", addCmdUse, tmp); err != nil {
			t.Fatalf("first Execute() returned error: %v", err)
		}

		if _, err := runWithStdin(t, "API\n", addCmdUse, filepath.Join(tmp, "api")); err != nil {
			t.Fatalf("second Execute() returned error: %v", err)
		}

		if projects := listProjects(t); len(projects) != 2 {
			t.Fatalf("ListProjects() returned %d projects, want 2", len(projects))
		}
	})

	t.Run("refuses a code already taken", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		tmp := t.TempDir()

		if _, err := runWithStdin(t, testCode+"\n", addCmdUse, filepath.Join(tmp, "a")); err != nil {
			t.Fatalf("first Execute() returned error: %v", err)
		}

		if _, err := runWithStdin(t, testCode+"\n", addCmdUse, filepath.Join(tmp, "b")); err == nil {
			t.Fatal("second Execute() returned nil error, want one")
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
		}
	})
}

func listProjects(t *testing.T) []orm.Project {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}
	defer sqlDB.Close()

	projects, err := orm.New(sqlDB).ListProjects(t.Context())
	if err != nil {
		t.Fatalf("ListProjects() returned error: %v", err)
	}

	return projects
}
