package db

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	alphaPath = "/tmp/alpha"
	alphaCode = "ALPHA"
)

func TestCreateProject(t *testing.T) {
	t.Run("round trips through list projects", func(t *testing.T) {
		q := openQueries(t)

		if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: alphaPath, Code: alphaCode}); err != nil {
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

		if p.Path != alphaPath {
			t.Errorf("Path = %q, want %q", p.Path, alphaPath)
		}

		if p.Code != alphaCode {
			t.Errorf("Code = %q, want %q", p.Code, alphaCode)
		}

		if p.ID == 0 {
			t.Error("ID = 0, want non-zero")
		}
	})

	refusals := []struct {
		name   string
		second orm.CreateProjectParams
	}{
		{name: "refuses a duplicate code", second: orm.CreateProjectParams{Path: "/tmp/beta", Code: alphaCode}},
		{name: "refuses a duplicate path", second: orm.CreateProjectParams{Path: alphaPath, Code: "BETA"}},
	}

	for _, tt := range refusals {
		t.Run(tt.name, func(t *testing.T) {
			q := openQueries(t)

			if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: alphaPath, Code: alphaCode}); err != nil {
				t.Fatalf("CreateProject() returned error: %v", err)
			}

			if err := q.CreateProject(t.Context(), tt.second); err == nil {
				t.Errorf("CreateProject(%+v) returned nil error, want a constraint error", tt.second)
			}

			projects, err := q.ListProjects(t.Context())
			if err != nil {
				t.Fatalf("ListProjects() returned error: %v", err)
			}

			if len(projects) != 1 {
				t.Errorf("ListProjects() returned %d projects, want 1", len(projects))
			}
		})
	}
}

func openQueries(t *testing.T) *orm.Queries {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := Open()
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}

	t.Cleanup(func() { sqlDB.Close() })

	return orm.New(sqlDB)
}
