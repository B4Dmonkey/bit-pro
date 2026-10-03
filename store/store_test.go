package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
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

			want := filepath.Join(home, ".local", "share", "bit")

			if tt.xdg {
				data := t.TempDir()
				t.Setenv("XDG_DATA_HOME", data)
				want = filepath.Join(data, "bit")
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

func TestProjectDir(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)

	got, err := ProjectDir("BIT")
	if err != nil {
		t.Fatalf("ProjectDir() returned error: %v", err)
	}

	want := filepath.Join(data, "bit", "BIT")
	if got != want {
		t.Errorf("ProjectDir() = %q, want %q", got, want)
	}

	if _, err := os.Stat(got); !os.IsNotExist(err) {
		t.Errorf("os.Stat(%q) error = %v, want not exist", got, err)
	}
}
