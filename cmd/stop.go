package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/daemon"
	"github.com/spf13/cobra"
)

const stopCmdUse = "stop"

func newStopCmd(lc daemon.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   stopCmdUse,
		Short: "Stop the background daemon and keep it down across reboots",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := daemon.Stop(cmd.Context(), lc); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "stopped")

			return nil
		},
	}
}
