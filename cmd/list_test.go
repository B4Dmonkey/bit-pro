package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	aceCode = "ACE"
	midCode = "MID"
)

func TestListCmd(t *testing.T) {
	tests := []struct {
		name   string
		seed   []orm.CreateProjectParams
		remove []string
		want   string
	}{
		{
			name: "three projects",
			seed: []orm.CreateProjectParams{
				{Path: "/tmp/mid", Code: midCode},
				{Path: "/tmp/zed", Code: "ZED"},
				{Path: "/tmp/ace", Code: aceCode},
			},
			want: "ACE /tmp/ace MID /tmp/mid ZED /tmp/zed",
		},
		{
			name: "no database yet",
		},
		{
			name: "hides removed projects",
			seed: []orm.CreateProjectParams{
				{Path: "/tmp/ace", Code: aceCode},
				{Path: "/tmp/mid", Code: midCode},
			},
			remove: []string{midCode},
			want:   "ACE /tmp/ace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_DATA_HOME", "")

			for _, p := range tt.seed {
				seedProject(t, p)
			}

			for _, code := range tt.remove {
				markRemoved(t, code)
			}

			out, err := run(t, listCmdUse)
			if err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			if normalizeSpaces(out) != tt.want {
				t.Errorf("output = %q, want %q", normalizeSpaces(out), tt.want)
			}

			if tt.want != "" {
				if _, err := os.Stat(filepath.Join(home, ".local", "share", "bit", "main.db")); err != nil {
					t.Errorf("os.Stat(main.db) returned error: %v", err)
				}
			}
		})
	}
}

func seedProject(t *testing.T, params orm.CreateProjectParams) {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	if err := orm.New(sqlDB).CreateProject(t.Context(), params); err != nil {
		t.Fatalf("CreateProject(%+v) returned error: %v", params, err)
	}
}

func markRemoved(t *testing.T, code string) {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	q := orm.New(sqlDB)

	projects, err := q.ListProjects(t.Context())
	if err != nil {
		t.Fatalf("ListProjects() returned error: %v", err)
	}

	for _, p := range projects {
		if p.Code != code {
			continue
		}

		if err := q.SetProjectRemoved(t.Context(), orm.SetProjectRemovedParams{Removed: 1, ID: p.ID}); err != nil {
			t.Fatalf("SetProjectRemoved(%s) returned error: %v", code, err)
		}

		return
	}

	t.Fatalf("no project with code %s", code)
}
