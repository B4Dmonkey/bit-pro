package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/spf13/cobra"
)

func writeClaudeWiring(cmd *cobra.Command, run claude.Runner, dir string) error {
	if err := claude.WriteSettings(filepath.Join(dir, claudeDir, "settings.json")); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Bringing the bit plugin current...")

	if err := claude.SyncPlugin(cmd.Context(), run); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Registering bit MCP server...")

	if err := claude.RegisterMCP(cmd.Context(), run); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "bit MCP server registered (local scope).")

	return nil
}
