package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
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
				{Path: "/tmp/mid", Code: "MID"},
				{Path: "/tmp/zed", Code: "ZED"},
				{Path: "/tmp/ace", Code: "ACE"},
			},
			want: "ACE\t/tmp/ace\nMID\t/tmp/mid\nZED\t/tmp/zed\n",
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

			if out != tt.want {
				t.Errorf("output = %q, want %q", out, tt.want)
			}

			if _, err := os.Stat(filepath.Join(home, ".local", "share", "bit-pro", "bit.db")); err != nil {
				t.Errorf("os.Stat(bit.db) returned error: %v", err)
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
