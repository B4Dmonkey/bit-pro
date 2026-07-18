package cmd

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/B4Dmonkey/bit-pro/tui"
	"github.com/spf13/cobra"
)

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Browse tasks in a terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			tasks, err := task.New(bitDir).List()
			if err != nil {
				return err
			}
			return tui.Run(tui.New(tasks))
		},
	}
}
