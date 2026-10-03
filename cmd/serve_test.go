package cmd

import (
	"strings"
	"testing"
)

func TestServeCmd(t *testing.T) {
	t.Run("lists mcp in help", func(t *testing.T) {
		out := mustRun(t, "serve", "--help")

		if !strings.Contains(out, "mcp") {
			t.Errorf("bp serve --help output does not list mcp:\n%s", out)
		}
	})
}
