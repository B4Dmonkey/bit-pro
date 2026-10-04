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

	"github.com/B4Dmonkey/bit-pro/claude"
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

		wantOut := "Project code: added BIT " + want + "\n" +
			"Setting up bit in Claude Code (user scope)...\n"
		if out != wantOut {
			t.Errorf("output = %q, want %q", out, wantOut)
		}

		prompt := strings.SplitN(out, "added", 2)[0]
		if strings.Contains(prompt, "(") {
			t.Errorf("prompt = %q, want no %q", prompt, "(")
		}

		if _, err := os.Stat(".bit"); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(.bit) error = %v, want fs.ErrNotExist", err)
		}

		wantCalls := claude.GlobalWiring()
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

				wantOut := "Project code: added " + testCode + " " + want + "\n" +
					"Setting up bit in Claude Code (user scope)...\n"
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

	t.Run("revives a removed project", func(t *testing.T) {
		tests := []struct {
			name  string
			typed func(path string) string
		}{
			{name: "same case", typed: func(path string) string { return path }},
			{name: "different case", typed: strings.ToUpper},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				dir := initProject(t, testPrefix)
				root := storeDir(t)

				path, err := project.CanonicalPath(dir)
				if err != nil {
					t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
				}

				createTask(t, "One", "")
				createTask(t, "Two", "")

				if _, err := runWithStdin(t, "y\n", removeCmdUse); err != nil {
					t.Fatalf("bp remove returned error: %v", err)
				}

				var calls [][]string

				runner := func(_ context.Context, name string, args ...string) error {
					calls = append(calls, append([]string{name}, args...))
					return nil
				}

				out, err := runWithRunner(t, runner, "", addCmdUse, tt.typed(path))
				if err != nil {
					t.Fatalf("bp add returned error: %v", err)
				}

				if wantOut := "revived " + testPrefix + " " + path + "\n"; out != wantOut {
					t.Errorf("output = %q, want %q", out, wantOut)
				}

				if len(calls) != 0 {
					t.Errorf("calls = %v, want none", calls)
				}

				if p := loadProject(t, path); p.Removed {
					t.Errorf("project %s Removed = true, want false", path)
				}

				if out := mustRun(t, "task", "list"); strings.Contains(out, "BIT-") {
					t.Errorf("bp task list = %q, want no tracks", out)
				}

				assertExists(t, filepath.Join(root, "archive", "tasks", "BIT-1.json"))

				if id := createTask(t, "Next", ""); id != "BIT-3" {
					t.Errorf("created ID = %q, want %q", id, "BIT-3")
				}
			})
		}
	})

	t.Run("refuses a removed project code", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		tmp := t.TempDir()
		old := filepath.Join(tmp, "old")

		if err := os.Mkdir(old, 0o755); err != nil {
			t.Fatalf("os.Mkdir(old) returned error: %v", err)
		}

		oldPath, err := project.CanonicalPath(old)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", old, err)
		}

		if _, err := runWithStdin(t, testCode+"\n", addCmdUse, old); err != nil {
			t.Fatalf("first bp add returned error: %v", err)
		}

		t.Chdir(old)

		if _, err := runWithStdin(t, "y\n", removeCmdUse); err != nil {
			t.Fatalf("bp remove returned error: %v", err)
		}

		var calls [][]string

		run := func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			return nil
		}

		_, err = runWithRunner(t, run, "foo\n", addCmdUse, filepath.Join(tmp, "new"))
		if !errors.Is(err, project.ErrCodeRemoved) {
			t.Fatalf("second bp add error = %v, want %v", err, project.ErrCodeRemoved)
		}

		if msg := err.Error(); !strings.Contains(msg, testCode) || !strings.Contains(msg, oldPath) {
			t.Errorf("error = %q, want it to contain %q and %q", msg, testCode, oldPath)
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
		}

		projects := listProjects(t)
		if len(projects) != 1 {
			t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
		}

		if projects[0].Path != oldPath || projects[0].Removed == 0 {
			t.Errorf("project = %+v, want %s removed", projects[0], oldPath)
		}
	})

	t.Run("refuses a code already taken", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		tmp := t.TempDir()

		if _, err := runWithStdin(t, testCode+"\n", addCmdUse, filepath.Join(tmp, "a")); err != nil {
			t.Fatalf("first Execute() returned error: %v", err)
		}

		var calls [][]string

		run := func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			return nil
		}

		if _, err := runWithRunner(t, run, testCode+"\n", addCmdUse, filepath.Join(tmp, "b")); err == nil {
			t.Fatal("second Execute() returned nil error, want one")
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
		}

		if projects := listProjects(t); len(projects) != 1 {
			t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
		}
	})

	t.Run("wiring failure keeps the project and prints the commands", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		dir := t.TempDir()

		want, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("project.CanonicalPath(%q) returned error: %v", dir, err)
		}

		boom := errors.New("plugin bit not found")

		var calls [][]string

		run := func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			if args[1] == "install" {
				return boom
			}

			return nil
		}

		out, err := runWithRunner(t, run, testCode+"\n", addCmdUse, dir)
		if !errors.Is(err, boom) {
			t.Fatalf("Execute() error = %v, want %v", err, boom)
		}

		if wantAdded := "added " + testCode + " " + want; !strings.Contains(out, wantAdded) {
			t.Errorf("output = %q, want it to contain %q", out, wantAdded)
		}

		var b strings.Builder

		b.WriteString("Run these to finish setting up bit in Claude Code:\n")

		for _, argv := range claude.GlobalWiring() {
			b.WriteString("  " + strings.Join(argv, " ") + "\n")
		}

		if block := b.String(); !strings.Contains(out, block) {
			t.Errorf("output = %q, want it to contain %q", out, block)
		}

		if len(calls) != 3 {
			t.Errorf("calls = %v, want 3", calls)
		}

		projects := listProjects(t)
		if len(projects) != 1 {
			t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
		}

		if projects[0].Code != testCode {
			t.Errorf("Code = %q, want %q", projects[0].Code, testCode)
		}

		calls = nil

		out, err = runWithRunner(t, run, testCode+"\n", addCmdUse, dir)
		if err != nil {
			t.Fatalf("second Execute() returned error: %v", err)
		}

		if wantOut := "already added\n"; out != wantOut {
			t.Errorf("output = %q, want %q", out, wantOut)
		}

		if len(calls) != 0 {
			t.Errorf("calls = %v, want none", calls)
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
