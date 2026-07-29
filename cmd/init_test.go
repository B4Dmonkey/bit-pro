package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/assets"
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

func TestInitCmd_SeedsClaudeTree(t *testing.T) {
	t.Chdir(t.TempDir())

	mustRun(t, "init", "--prefix", "BIT")

	seeded := []string{
		".claude/bit-cli.md",
		".claude/skills/bit_scope/SKILL.md",
		".claude/skills/bit_plan/SKILL.md",
		".claude/skills/bit_do/SKILL.md",
		".claude/skills/bit_check/SKILL.md",
	}
	for _, rel := range seeded {
		got, err := os.ReadFile(rel)
		if err != nil {
			t.Fatalf("os.ReadFile(%q) returned error: %v", rel, err)
		}
		want, err := assets.FS.ReadFile(filepath.ToSlash(rel[len(".claude/"):]))
		if err != nil {
			t.Fatalf("assets.FS.ReadFile for %q returned error: %v", rel, err)
		}
		if string(got) != string(want) {
			t.Errorf("%s bytes do not match the embedded copy", rel)
		}
	}

	if _, err := os.Stat(".claude/CLAUDE.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(.claude/CLAUDE.md) error = %v, want fs.ErrNotExist", err)
	}
}

func TestInitCmd_ReseedRefreshes(t *testing.T) {
	t.Chdir(t.TempDir())

	mustRun(t, "init", "--prefix", "BIT")

	stale := filepath.Join(".claude", "skills", "bit_do", "SKILL.md")
	if err := os.WriteFile(stale, []byte("stale\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", stale, err)
	}

	mustRun(t, "init", "--prefix", "BIT")

	got, err := os.ReadFile(stale)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", stale, err)
	}
	want, err := assets.FS.ReadFile("skills/bit_do/SKILL.md")
	if err != nil {
		t.Fatalf("assets.FS.ReadFile returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("%s was not refreshed to the embedded copy", stale)
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
