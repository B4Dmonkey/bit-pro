package cmd

import (
	"path/filepath"
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
