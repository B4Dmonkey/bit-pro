package task

import (
	"github.com/B4Dmonkey/bit-pro/bitdir"
	taskstore "github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Complete a task, filing it and its bars under .bit/completed/",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return taskstore.New(bitdir.Current()).Complete(args[0])
		},
	}
}
