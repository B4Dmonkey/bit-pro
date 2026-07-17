package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestInitCmd_CapturesPrefix(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		stdin string
	}{
		{name: "from --prefix flag", args: []string{"init", "--prefix", "BIT"}},
		{name: "from interactive prompt", args: []string{"init"}, stdin: "BIT\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

			if _, err := runWithStdin(t, tt.stdin, tt.args...); err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}

			cfg, err := task.New(".bit").Config()
			if err != nil {
				t.Fatalf("reading config: %v", err)
			}
			if cfg.Prefix != "BIT" {
				t.Errorf("cfg.Prefix = %q, want %q", cfg.Prefix, "BIT")
			}
		})
	}
}

func TestInitCmd_RejectsBadInvocations(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		stdin string
	}{
		{name: "empty prefix at the prompt", args: []string{"init"}, stdin: "\n"},
		{name: "extra positional args", args: []string{"init", "garbage", "--prefix", "BIT"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

			if _, err := runWithStdin(t, tt.stdin, tt.args...); err == nil {
				t.Fatal("Execute() returned nil error, want non-nil")
			}

			if _, err := os.Stat(".bit/config.toml"); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("os.Stat(.bit/config.toml) error = %v, want fs.ErrNotExist", err)
			}
		})
	}
}

func TestInitCmd_IsIdempotent(t *testing.T) {
	tests := []struct {
		name string
		runs int
	}{
		{name: "fresh directory", runs: 1},
		{name: "already initialized", runs: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)

			for range tt.runs {
				mustRun(t, "init", "--prefix", "BIT")
			}

			info, err := os.Stat(filepath.Join(dir, ".bit"))
			if err != nil {
				t.Fatalf("os.Stat(.bit) returned error: %v", err)
			}
			if !info.IsDir() {
				t.Error(".bit exists but is not a directory")
			}
		})
	}
}
