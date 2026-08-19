package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/launchd"
	"github.com/spf13/cobra"
)

const statusCmdUse = "status"

func newStatusCmd(lc launchd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   statusCmdUse,
		Short: "Report whether the background daemon is running",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			state, pid, err := launchd.Status(cmd.Context(), lc)
			if err != nil {
				return err
			}

			if state == launchd.StateRunning {
				fmt.Fprintf(cmd.OutOrStdout(), "running (pid %d)\n", pid)

				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), state)

			return nil
		},
	}
}
