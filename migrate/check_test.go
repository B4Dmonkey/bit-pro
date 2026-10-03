package migrate

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestUnknownFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		dirs  []string
		want  []string
	}{
		{
			name: "every known shape",
			files: []string{
				"config.toml",
				"tasks/BIT-1.md",
				"completed/BIT-2.md",
				"archive/tasks/BIT-3.md",
				"feedback/BIT-1-001.md",
				"research/BIT-1/index.md",
				"retro/album-proposals.md",
			},
		},
		{
			name:  "research too deep",
			files: []string{"research/BIT-1/sub/x.md"},
			want:  []string{"research/BIT-1/sub/x.md"},
		},
		{
			name:  "research file at the top",
			files: []string{"research/index.md"},
			want:  []string{"research/index.md"},
		},
		{
			name:  "feedback with no seq",
			files: []string{"feedback/notes.md"},
			want:  []string{"feedback/notes.md"},
		},
		{
			name:  "retro without proposals",
			files: []string{"retro/summary.md"},
			want:  []string{"retro/summary.md"},
		},
		{
			name:  "flat archive",
			files: []string{"archive/BIT-3.md"},
			want:  []string{"archive/BIT-3.md"},
		},
		{
			name:  "non-md in a task dir",
			files: []string{"tasks/BIT-1.txt"},
			want:  []string{"tasks/BIT-1.txt"},
		},
		{
			name:  "dotfiles",
			files: []string{".DS_Store", "tasks/.DS_Store"},
			want:  []string{".DS_Store", "tasks/.DS_Store"},
		},
		{
			name: "empty completed dir",
			dirs: []string{"completed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := t.TempDir()

			for _, d := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(src, d), 0o755); err != nil {
					t.Fatalf("os.MkdirAll(%q) returned error: %v", d, err)
				}
			}

			for _, f := range tt.files {
				path := filepath.Join(src, filepath.FromSlash(f))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
				}

				if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
					t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
				}
			}

			got, err := unknownFiles(src)
			if err != nil {
				t.Fatalf("unknownFiles() returned error: %v", err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("unknownFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
