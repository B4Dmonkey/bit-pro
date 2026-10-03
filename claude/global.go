package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func EnsureGlobal(ctx context.Context, run Runner, home string) error {
	steps := GlobalWiring()
	skipMCP := hasUserMCP(home)

	for i, argv := range steps {
		if i == len(steps)-1 && skipMCP {
			break
		}

		if err := run(ctx, argv[0], argv[1:]...); err != nil {
			return fmt.Errorf("claude wiring step %d of %d: %w", i+1, len(steps), err)
		}
	}

	return nil
}

func hasUserMCP(home string) bool {
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return false
	}

	var cfg struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}

	_, ok := cfg.MCPServers["bit"]

	return ok
}
