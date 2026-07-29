package cmd

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Complete a task, filing it and its bars under .bit/completed/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return task.New(bitDir).Complete(args[0])
		},
	}
}
