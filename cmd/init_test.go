package cmd

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

func TestInitCmd_UppercasesTheStoredPrefix(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		stdin string
	}{
		{name: "from --prefix flag", args: []string{"init", "--prefix", "foo"}},
		{name: "from interactive prompt", args: []string{"init"}, stdin: "foo\n"},
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
			if cfg.Prefix != "FOO" {
				t.Errorf("cfg.Prefix = %q, want %q", cfg.Prefix, "FOO")
			}

			data, err := os.ReadFile(filepath.Join(".bit", "config.toml"))
			if err != nil {
				t.Fatalf("os.ReadFile(.bit/config.toml) returned error: %v", err)
			}
			if !strings.Contains(string(data), `"FOO"`) {
				t.Errorf("config.toml = %s, want it to contain %q", data, `"FOO"`)
			}

			out := mustRun(t, "task", "create", "first", "-d", "...")
			if out != "FOO-1\n" {
				t.Errorf("task create stdout = %q, want %q", out, "FOO-1\n")
			}
			if _, err := os.Stat(filepath.Join(".bit", "tasks", "FOO-1.md")); err != nil {
				t.Errorf("os.Stat(.bit/tasks/FOO-1.md) returned error: %v", err)
			}
		})
	}
}

func TestInitCmd_ReuseExistingPrefixOnEnter(t *testing.T) {
	t.Chdir(t.TempDir())

	mustRun(t, "init", "--prefix", "BIT")

	if _, err := runWithStdin(t, "\n", "init"); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	cfg, err := task.New(".bit").Config()
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if cfg.Prefix != "BIT" {
		t.Errorf("cfg.Prefix = %q, want %q", cfg.Prefix, "BIT")
	}
}

func TestInitCmd_PromptShowsExistingPrefix(t *testing.T) {
	tests := []struct {
		name      string
		seed      bool
		wantShown string
		wantParen bool
	}{
		{name: "existing config", seed: true, wantShown: "Task ID prefix (BIT): ", wantParen: true},
		{name: "fresh project", seed: false, wantShown: "Task ID prefix: ", wantParen: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

			if tt.seed {
				mustRun(t, "init", "--prefix", "BIT")
			}

			out, _ := runWithStdin(t, "BIT\n", "init")

			if !strings.Contains(out, tt.wantShown) {
				t.Errorf("prompt = %q, want it to contain %q", out, tt.wantShown)
			}
			if !tt.wantParen && strings.Contains(out, "(") {
				t.Errorf("prompt = %q, want no %q", out, "(")
			}
		})
	}
}

func TestInitCmd_WritesNoSkills(t *testing.T) {
	t.Chdir(t.TempDir())

	mustRun(t, "init", "--prefix", "BIT")

	for _, rel := range []string{".claude/skills", ".claude/bit-cli.md"} {
		if _, err := os.Stat(rel); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Stat(%q) error = %v, want fs.ErrNotExist", rel, err)
		}
	}

	if _, err := os.Stat(filepath.Join(".claude", "settings.json")); err != nil {
		t.Errorf("os.Stat(.claude/settings.json) returned error: %v", err)
	}
}

func TestInitCmd_WritesPluginWiring(t *testing.T) {
	t.Chdir(t.TempDir())

	mustRun(t, "init", "--prefix", "BIT")

	data, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("os.ReadFile(.claude/settings.json) returned error: %v", err)
	}
	if !strings.Contains(string(data), "bit@bit-pro") {
		t.Errorf("settings.json = %s, want it to contain %q", data, "bit@bit-pro")
	}
}

func TestInitCmd_SyncsThePlugin(t *testing.T) {
	t.Chdir(t.TempDir())

	var calls [][]string
	run := func(_ context.Context, name string, args ...string) error {
		calls = append(calls, append([]string{name}, args...))
		return nil
	}

	if _, err := runWithRunner(t, run, "", "init", "--prefix", "BIT"); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := [][]string{
		{"claude", "plugin", "marketplace", "update", "bit-pro"},
		{"claude", "plugin", "update", "bit@bit-pro", "--scope", "project"},
	}
	if !slices.EqualFunc(calls, want, slices.Equal) {
		t.Errorf("calls = %v, want %v", calls, want)
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
