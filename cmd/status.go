package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const statusCmdUse = "status"

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   statusCmdUse,
		Short: "Report whether the background daemon is running",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "not running")

			return nil
		},
	}
}
