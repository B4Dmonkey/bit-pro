package project

import (
	"path/filepath"
	"testing"
)

func TestMainCheckout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		dir    string
		want   string
		wantOK bool
	}{
		{"worktree root", "/r/.claude/worktrees/hazy", "/r", true},
		{"deep inside a worktree", "/r/.claude/worktrees/hazy/src/pkg", "/r", true},
		{"nested worktree resolves to the outermost", "/r/.claude/worktrees/outer/.claude/worktrees/inner", "/r", true},
		{"outside a worktree", "/r/src", "", false},
		{"claude dir without worktrees", "/r/.claude/settings", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := MainCheckout(filepath.FromSlash(tt.dir))

			if got != filepath.FromSlash(tt.want) || ok != tt.wantOK {
				t.Errorf("MainCheckout(%q) = (%q, %v), want (%q, %v)", tt.dir, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
