package bitdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoot(t *testing.T) {
	plain := t.TempDir()
	single := t.TempDir()
	nested := t.TempDir()

	tests := []struct {
		name string
		wd   string
		want string
	}{
		{"outside a worktree", plain, plain},
		{"inside a worktree", filepath.Join(single, claudeDir, worktreesDir, "hazy"), single},
		{
			"inside a nested worktree",
			filepath.Join(nested, claudeDir, worktreesDir, "outer", claudeDir, worktreesDir, "inner"),
			nested,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.MkdirAll(tt.wd, 0o755); err != nil {
				t.Fatalf("os.MkdirAll(%q) returned error: %v", tt.wd, err)
			}

			t.Chdir(tt.wd)

			got := Root()

			if resolve(t, got) != resolve(t, tt.want) {
				t.Errorf("Root() = %q, want %q", got, tt.want)
			}
		})
	}
}

func resolve(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks(%q) returned error: %v", path, err)
	}

	return resolved
}
