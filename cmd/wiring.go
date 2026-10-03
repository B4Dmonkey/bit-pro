package cmd

import (
	"fmt"
	"os"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/spf13/cobra"
)

func ensureGlobalWiring(cmd *cobra.Command, run claude.Runner) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Setting up bit in Claude Code (user scope)...")

	return claude.EnsureGlobal(cmd.Context(), run, home)
}
