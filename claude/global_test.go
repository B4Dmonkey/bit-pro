package claude

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

const (
	addSubCmd         = "add"
	bitProMarketplace = "bit-pro"
	marketplaceSubCmd = "marketplace"
)

func TestEnsureGlobal(t *testing.T) {
	t.Run("runs the four commands in order", func(t *testing.T) {
		home := t.TempDir()
		rec := newRecorder(nil)

		if err := EnsureGlobal(t.Context(), rec.Run, home); err != nil {
			t.Fatalf("EnsureGlobal returned error: %v", err)
		}

		want := [][]string{
			{claudeBin, pluginSubCmd, marketplaceSubCmd, addSubCmd, "B4Dmonkey/bit-pro"},
			{claudeBin, pluginSubCmd, marketplaceSubCmd, updateSubCmd, bitProMarketplace},
			{claudeBin, pluginSubCmd, "install", bitProPlugin, scopeFlag, "user"},
			{claudeBin, mcpSubCmd, addSubCmd, "-s", "user", bitServer, "--", "bp", "serve", mcpSubCmd},
		}
		if !slices.EqualFunc(rec.calls, want, slices.Equal) {
			t.Errorf("calls = %v, want %v", rec.calls, want)
		}

		if got := GlobalWiring(); !slices.EqualFunc(got, want, slices.Equal) {
			t.Errorf("GlobalWiring() = %v, want %v", got, want)
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
