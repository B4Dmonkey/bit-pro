package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir_FollowsXDGDataHome(t *testing.T) {
	tests := []struct {
		name string
		xdg  bool
	}{
		{name: "XDG_DATA_HOME unset"},
		{name: "XDG_DATA_HOME set", xdg: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)

			want := filepath.Join(home, ".local", "share", "bit-pro")

			if tt.xdg {
				data := t.TempDir()
				t.Setenv("XDG_DATA_HOME", data)
				want = filepath.Join(data, "bit-pro")
			} else {
				t.Setenv("XDG_DATA_HOME", "")
			}

			got, err := Dir()
			if err != nil {
				t.Fatalf("Dir() returned error: %v", err)
			}

			if got != want {
				t.Errorf("Dir() = %q, want %q", got, want)
			}

			info, err := os.Stat(got)
			if err != nil {
				t.Fatalf("os.Stat(%q) returned error: %v", got, err)
			}

			if !info.IsDir() {
				t.Errorf("%s is not a directory", got)
			}
		})
	}
}
