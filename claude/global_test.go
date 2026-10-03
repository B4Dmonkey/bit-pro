package claude

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	claudeBin         = "claude"
	scopeFlag         = "--scope"
	pluginSubCmd      = "plugin"
	bitProPlugin      = "bit@bit-pro"
	updateSubCmd      = "update"
	mcpSubCmd         = "mcp"
	bitServer         = "bit"
	addSubCmd         = "add"
	bitProMarketplace = "bit-pro"
	marketplaceSubCmd = "marketplace"
)

type recorder struct {
	calls [][]string
	errs  map[int]error
}

func newRecorder(errs map[int]error) *recorder {
	return &recorder{errs: errs}
}

func (r *recorder) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, append([]string{name}, args...))
	return r.errs[len(r.calls)-1]
}

func TestEnsureGlobal(t *testing.T) {
	all := [][]string{
		{claudeBin, pluginSubCmd, marketplaceSubCmd, addSubCmd, "B4Dmonkey/bit-pro"},
		{claudeBin, pluginSubCmd, marketplaceSubCmd, updateSubCmd, bitProMarketplace},
		{claudeBin, pluginSubCmd, "install", bitProPlugin, scopeFlag, "user"},
		{claudeBin, mcpSubCmd, addSubCmd, "-s", "user", bitServer, "--", "bp", "serve", mcpSubCmd},
	}

	tests := []struct {
		name       string
		claudeJSON string
		want       [][]string
	}{
		{
			name: "skips mcp add when a user entry exists",
			claudeJSON: `{"mcpServers": {"bit": {"type": "stdio", "command": "bp", "args": ["serve", "mcp"], "env": {}}}, ` +
				`"projects": {}}`,
			want: all[:3],
		},
		{
			name: "runs the four commands in order",
			want: all,
		},
		{
			name:       "adds mcp for a local-only entry",
			claudeJSON: `{"projects": {"/p/a": {"mcpServers": {"bit": {"command": "bp"}}}}}`,
			want:       all,
		},
		{
			name:       "adds mcp when only another user server exists",
			claudeJSON: `{"mcpServers": {"other": {}}}`,
			want:       all,
		},
		{
			name:       "adds mcp when claude json is malformed",
			claudeJSON: `{`,
			want:       all,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if tt.claudeJSON != "" {
				writeFixture(t, filepath.Join(home, ".claude.json"), tt.claudeJSON)
			}

			rec := newRecorder(nil)

			if err := EnsureGlobal(t.Context(), rec.Run, home); err != nil {
				t.Fatalf("EnsureGlobal returned error: %v", err)
			}

			if !slices.EqualFunc(rec.calls, tt.want, slices.Equal) {
				t.Errorf("calls = %v, want %v", rec.calls, tt.want)
			}
		})
	}

	t.Run("GlobalWiring lists the four commands", func(t *testing.T) {
		if got := GlobalWiring(); !slices.EqualFunc(got, all, slices.Equal) {
			t.Errorf("GlobalWiring() = %v, want %v", got, all)
		}
	})

	t.Run("stops at the first failing step", func(t *testing.T) {
		boom := errors.New("plugin bit not found")
		rec := newRecorder(map[int]error{2: boom})

		err := EnsureGlobal(t.Context(), rec.Run, t.TempDir())
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want it to wrap %v", err, boom)
		}

		if !strings.Contains(err.Error(), "step 3 of 4") {
			t.Errorf("err = %v, want it to name step 3 of 4", err)
		}

		if len(rec.calls) != 3 {
			t.Errorf("calls = %v, want 3 calls", rec.calls)
		}
	})
}
