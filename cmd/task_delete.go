package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newTaskDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("confirmation required; pass --yes or confirm the prompt")
			}
			return os.Remove(taskPath(args[0]))
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	return cmd
}
