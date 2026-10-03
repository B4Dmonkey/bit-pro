package cmd

import (
	"github.com/spf13/cobra"
)

const serveCmdUse = "serve"

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   serveCmdUse,
		Short: "Run foreground servers",
	}

	cmd.AddCommand(newServeMCPCmd())

	return cmd
}
