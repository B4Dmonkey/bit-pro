package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tasks, err := task.New(bitDir).List()
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			for _, t := range tasks {
				fmt.Fprintf(out, "%s\t%s\t%s\n", t.ID, t.Status, t.Title)
			}
			return nil
		},
	}
}
