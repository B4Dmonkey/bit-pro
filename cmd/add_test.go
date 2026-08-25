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
)

func TestAddCmd_EnrollsUsingTheBitPrefix(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	initProject(t, testPrefix)

	want, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("filepath.Abs(.) returned error: %v", err)
	}

	out, err := runWithStdin(t, "\n", addCmdUse, ".")
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if wantOut := "Project code (BIT): added BIT " + want + "\n"; out != wantOut {
		t.Errorf("output = %q, want %q", out, wantOut)
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

	if projects[0].Path != want {
		t.Errorf("Path = %q, want %q", projects[0].Path, want)
	}
}

func TestAddCmd_InitialisesAProjectWithoutBit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Chdir(t.TempDir())

	var calls [][]string

	run := func(_ context.Context, name string, args ...string) error {
		calls = append(calls, append([]string{name}, args...))
		return nil
	}

	want, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("filepath.Abs(.) returned error: %v", err)
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
}

func TestAddCmd_UppercasesATypedCode(t *testing.T) {
	tests := []struct {
		name  string
		typed string
	}{
		{name: "lowercase", typed: "foo"},
		{name: "uppercase", typed: testCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_DATA_HOME", "")

			initProject(t, testPrefix)

			want, err := filepath.Abs(".")
			if err != nil {
				t.Fatalf("filepath.Abs(.) returned error: %v", err)
			}

			out, err := runWithStdin(t, tt.typed+"\n", addCmdUse, ".")
			if err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			if wantOut := "Project code (BIT): added " + testCode + " " + want + "\n"; out != wantOut {
				t.Errorf("output = %q, want %q", out, wantOut)
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

			if projects[0].Code != testCode {
				t.Errorf("Code = %q, want %q", projects[0].Code, testCode)
			}
		})
	}
}

func TestAddCmd_RejectsAnEmptyCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Chdir(t.TempDir())

	_, err := runWithStdin(t, "\n", addCmdUse, ".")
	if err == nil {
		t.Fatal("Execute() returned nil error, want non-nil")
	}

	if want := "project code cannot be empty"; err.Error() != want {
		t.Errorf("err = %q, want %q", err, want)
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
}

func TestAddCmd_SkipsAPathAlreadyEnrolled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	initProject(t, testPrefix)

	if _, err := runWithStdin(t, "\n", addCmdUse, "."); err != nil {
		t.Fatalf("first Execute() returned error: %v", err)
	}

	out, err := runWithStdin(t, "\n", addCmdUse, ".")
	if err != nil {
		t.Fatalf("second Execute() returned error: %v", err)
	}

	if wantOut := "already added\n"; out != wantOut {
		t.Errorf("output = %q, want %q", out, wantOut)
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
}
