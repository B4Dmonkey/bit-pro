package claude

import (
	"context"
	"fmt"
)

const (
	claudeCLI    = "claude"
	pluginCLICmd = "plugin"
)

func GlobalWiring() [][]string {
	return [][]string{
		{claudeCLI, pluginCLICmd, "marketplace", "add", "B4Dmonkey/bit-pro"},
		{claudeCLI, pluginCLICmd, "marketplace", "update", marketplaceName},
		{claudeCLI, pluginCLICmd, "install", pluginKey, "--scope", "user"},
		{claudeCLI, "mcp", "add", "-s", "user", "bit", "--", "bp", "serve", "mcp"},
	}
}

func EnsureGlobal(ctx context.Context, run Runner, _ string) error {
	steps := GlobalWiring()
	for i, argv := range steps {
		if err := run(ctx, argv[0], argv[1:]...); err != nil {
			return fmt.Errorf("claude wiring step %d of %d: %w", i+1, len(steps), err)
		}
	}

	return nil
}
