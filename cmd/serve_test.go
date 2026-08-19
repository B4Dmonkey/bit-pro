package cmd

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestServeCmd_ReturnsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	out, err := runWithContext(t, ctx, serveCmdUse)
	if err != nil {
		t.Fatalf("bp serve returned error: %v", err)
	}

	if out != "" {
		t.Errorf("bp serve wrote output %q, want none", out)
	}
}

func TestServeCmd_IsListedInHelp(t *testing.T) {
	out := mustRun(t, "--help")

	if !strings.Contains(out, "serve") {
		t.Errorf("bp --help output does not list serve:\n%s", out)
	}
}

func TestServeCmd_TicksOnlyWhenVerbose(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantTicks bool
	}{
		{name: "verbose logs ticks at debug", args: []string{serveCmdUse, "-v"}, wantTicks: true},
		{name: "default is silent", args: []string{serveCmdUse}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := serveTick
			serveTick = 5 * time.Millisecond

			t.Cleanup(func() { serveTick = original })

			ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
			defer cancel()

			out, err := runWithContext(t, ctx, tt.args...)
			if err != nil {
				t.Fatalf("bp %s returned error: %v", strings.Join(tt.args, " "), err)
			}

			if !tt.wantTicks {
				if out != "" {
					t.Errorf("bp serve wrote output %q, want none", out)
				}

				return
			}

			if got := strings.Count(out, "tick"); got < 2 {
				t.Errorf("bp serve -v logged %d ticks, want at least 2:\n%s", got, out)
			}

			if !strings.Contains(out, "DEBUG") {
				t.Errorf("bp serve -v did not log at debug level:\n%s", out)
			}
		})
	}
}
