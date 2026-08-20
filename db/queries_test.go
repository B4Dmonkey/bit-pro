package db

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

func TestProjects_RoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := Open()
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	defer sqlDB.Close()

	q := orm.New(sqlDB)

	if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: "/tmp/alpha", Code: "ALPHA"}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	projects, err := q.ListProjects(t.Context())
	if err != nil {
		t.Fatalf("ListProjects() returned error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
	}

	p := projects[0]

	if p.Path != "/tmp/alpha" {
		t.Errorf("Path = %q, want %q", p.Path, "/tmp/alpha")
	}

	if p.Code != "ALPHA" {
		t.Errorf("Code = %q, want %q", p.Code, "ALPHA")
	}

	if p.ID == 0 {
		t.Error("ID = 0, want non-zero")
	}
}
