package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create the .bit/ directory bit uses to track this project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return os.MkdirAll(".bit", 0o755)
		},
	}
}
