package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/spf13/cobra"
)

func ensureGlobalWiring(cmd *cobra.Command, run claude.Runner) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Setting up bit in Claude Code (user scope)...")

	err = claude.EnsureGlobal(cmd.Context(), run, home)
	if err != nil {
		w := cmd.ErrOrStderr()
		fmt.Fprintln(w, "Run these to finish setting up bit in Claude Code:")

		for _, argv := range claude.GlobalWiring() {
			fmt.Fprintln(w, "  "+strings.Join(argv, " "))
		}
	}

	return err
}
