package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/assets"
	"github.com/spf13/cobra"
)

func newInstructionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "instructions",
		Short: "Print the bp command contract the bit skills drive",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			contract, err := assets.FS.ReadFile("bit-cli.md")
			if err != nil {
				return fmt.Errorf("reading embedded bit-cli.md: %w", err)
			}

			fmt.Fprint(cmd.OutOrStdout(), string(contract))

			return nil
		},
	}
}
