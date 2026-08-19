package cmd

import (
	"context"
	"strings"
	"testing"
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
