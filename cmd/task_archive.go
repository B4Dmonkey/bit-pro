package cmd

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskArchiveCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a task, relocating it out of the active list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return task.New(bitDir).Relocate(args[0], force)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "archive even when bars are unfinished")
	return cmd
}
