package cmd

import (
	"github.com/spf13/cobra"
)

const serveCmdUse = "serve"

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   serveCmdUse,
		Short: "Run the automation loop in the foreground",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			<-cmd.Context().Done()

			return nil
		},
	}
}
