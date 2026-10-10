package gittest

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	t.Run("ignores an inherited index file", func(t *testing.T) {
		bogus := filepath.Join(t.TempDir(), "missing", "index")
		t.Setenv("GIT_INDEX_FILE", bogus)

		dir := t.TempDir()
		Run(t, dir, "init", "-b", "main")

		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		Run(t, dir, "add", "a.txt")
		Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-m", "x")

		if got := Run(t, dir, "ls-files"); got != "a.txt" {
			t.Errorf("ls-files = %q, want %q", got, "a.txt")
		}

		if _, err := os.Stat(bogus); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("Stat(%s) error = %v, want fs.ErrNotExist", bogus, err)
		}
	})
}

func TestIsolate(t *testing.T) {
	t.Run("unsets inherited git variables", func(t *testing.T) {
		bogus := filepath.Join(t.TempDir(), "missing", "index")
		t.Setenv("GIT_INDEX_FILE", bogus)
		t.Setenv("GIT_DIR", bogus)

		Isolate(t)

		for _, k := range []string{"GIT_INDEX_FILE", "GIT_DIR"} {
			if v, ok := os.LookupEnv(k); ok {
				t.Errorf("%s = %q, want unset", k, v)
			}
		}
	})
}
