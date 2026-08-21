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

func TestListCmd_PrintsProjectsByCode(t *testing.T) {
	tests := []struct {
		name string
		seed []orm.CreateProjectParams
		want string
	}{
		{
			name: "three projects",
			seed: []orm.CreateProjectParams{
				{Path: "/tmp/mid", Code: midCode},
				{Path: "/tmp/zed", Code: "ZED"},
				{Path: "/tmp/ace", Code: aceCode},
			},
			want: "ACE /tmp/ace backlog:0 todo:0 done:0 completed:0" +
				" MID /tmp/mid backlog:0 todo:0 done:0 completed:0" +
				" ZED /tmp/zed backlog:0 todo:0 done:0 completed:0",
		},
		{
			name: "no database yet",
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

			out, err := run(t, listCmdUse)
			if err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			if normalizeSpaces(out) != tt.want {
				t.Errorf("output = %q, want %q", normalizeSpaces(out), tt.want)
			}

			if tt.want != "" {
				if _, err := os.Stat(filepath.Join(home, ".local", "share", "bit-pro", "bit.db")); err != nil {
					t.Errorf("os.Stat(bit.db) returned error: %v", err)
				}
			}
		})
	}
}

func TestListCmd_ShowsProjectCounts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	seedProject(t, orm.CreateProjectParams{Path: "/tmp/ace", Code: aceCode})
	seedProject(t, orm.CreateProjectParams{Path: "/tmp/mid", Code: midCode})

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
		counts := orm.UpdateProjectCountsParams{ID: p.ID}

		switch p.Code {
		case aceCode:
			counts.Backlog, counts.Todo, counts.Done, counts.Completed = 2, 1, 4, 7
		case midCode:
			counts.Backlog, counts.Todo, counts.Done, counts.Completed = 0, 3, 12, 2
		}

		if err := q.UpdateProjectCounts(t.Context(), counts); err != nil {
			t.Fatalf("UpdateProjectCounts(%s) returned error: %v", p.Code, err)
		}
	}

	out, err := run(t, listCmdUse)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := "ACE /tmp/ace backlog:2 todo:1 done:4 completed:7" +
		" MID /tmp/mid backlog:0 todo:3 done:12 completed:2"

	if got := normalizeSpaces(out); got != want {
		t.Errorf("output = %q, want %q", got, want)
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
