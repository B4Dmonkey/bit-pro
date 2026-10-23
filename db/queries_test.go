package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	alphaPath    = "/tmp/alpha"
	alphaCode    = "ALPHA"
	beforeChange = "2026-10-20T14:00:00Z"
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

func TestSetProjectRemoved(t *testing.T) {
	t.Run("keeps the row and toggles the flag", func(t *testing.T) {
		q := openQueries(t)

		if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: alphaPath, Code: alphaCode}); err != nil {
			t.Fatalf("CreateProject() returned error: %v", err)
		}

		created, err := q.ListProjects(t.Context())
		if err != nil {
			t.Fatalf("ListProjects() returned error: %v", err)
		}

		id := created[0].ID

		for _, want := range []int64{1, 0} {
			if err := q.SetProjectRemoved(t.Context(), orm.SetProjectRemovedParams{Removed: want, ID: id}); err != nil {
				t.Fatalf("SetProjectRemoved(%d) returned error: %v", want, err)
			}

			projects, err := q.ListProjects(t.Context())
			if err != nil {
				t.Fatalf("ListProjects() returned error: %v", err)
			}

			if len(projects) != 1 {
				t.Fatalf("ListProjects() returned %d projects, want 1", len(projects))
			}

			if projects[0].Removed != want {
				t.Errorf("Removed = %d, want %d", projects[0].Removed, want)
			}
		}
	})
}

func TestUsageQueries(t *testing.T) {
	t.Run("deleting a project removes its rows", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		sqlDB, err := Open()
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		q := orm.New(sqlDB)

		if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: alphaPath, Code: alphaCode}); err != nil {
			t.Fatalf("CreateProject() returned error: %v", err)
		}

		projects, err := q.ListProjects(t.Context())
		if err != nil {
			t.Fatalf("ListProjects() returned error: %v", err)
		}

		row := orm.UpsertUsageParams{
			ProjectID: projects[0].ID, SessionID: "s1", Track: "BIT-1",
			MessageID: "msg_01", RequestID: "req_01", Ts: beforeChange, Model: "claude-opus-5-5",
		}
		if err := q.UpsertUsage(t.Context(), row); err != nil {
			t.Fatalf("UpsertUsage() returned error: %v", err)
		}

		if _, err := sqlDB.Exec("DELETE FROM projects WHERE id = ?", projects[0].ID); err != nil {
			t.Fatalf("deleting project: %v", err)
		}

		var n int
		if err := sqlDB.QueryRow("SELECT count(*) FROM usage").Scan(&n); err != nil {
			t.Fatalf("counting usage: %v", err)
		}

		if n != 0 {
			t.Errorf("usage rows = %d, want 0", n)
		}
	})
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

func TestPriceAt(t *testing.T) {
	const opus = "claude-opus-5-5"

	t.Run("uses the price in effect at the time", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")

		sqlDB, err := Open()
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		if _, err := sqlDB.Exec(
			"INSERT INTO model_prices (model, effective_from, input) VALUES (?, ?, ?)", opus, "2026-11-01", 3,
		); err != nil {
			t.Fatalf("inserting a price change: %v", err)
		}

		q := orm.New(sqlDB)

		tests := []struct {
			at   string
			want float64
		}{
			{beforeChange, 4},
			{"2026-11-02T00:00:00Z", 3},
		}

		for _, tt := range tests {
			got, err := q.PriceAt(t.Context(), orm.PriceAtParams{Model: opus, EffectiveFrom: tt.at})
			if err != nil {
				t.Fatalf("PriceAt(%q) returned error: %v", tt.at, err)
			}

			if got.Input.Float64 != tt.want {
				t.Errorf("PriceAt(%q).Input = %v, want %v", tt.at, got.Input.Float64, tt.want)
			}
		}
	})

	t.Run("unknown model", func(t *testing.T) {
		q := openQueries(t)

		_, err := q.PriceAt(t.Context(), orm.PriceAtParams{Model: "claude-nextgen-9", EffectiveFrom: beforeChange})
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("PriceAt() error = %v, want sql.ErrNoRows", err)
		}
	})
}
