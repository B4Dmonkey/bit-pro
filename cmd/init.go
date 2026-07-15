package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var prefix string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the .bit/ directory bit uses to track this project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(".bit", 0o755); err != nil {
				return err
			}
			if prefix != "" {
				return saveConfig(&Config{Prefix: prefix})
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&prefix, "prefix", "", "task ID prefix for this project (e.g. BIT)")
	return cmd
}
