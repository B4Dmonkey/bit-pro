package task

import (
	"github.com/B4Dmonkey/bit-pro/bitdir"
	taskstore "github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newMoveCmd() *cobra.Command {
	var before, after string

	cmd := &cobra.Command{
		Use:   "move <bar>",
		Short: "Resequence a bar within its track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return taskstore.New(bitdir.Current()).Move(args[0], before, after)
		},
	}
	cmd.Flags().StringVar(&before, "before", "", "move the bar directly before this sibling")
	cmd.Flags().StringVar(&after, "after", "", "move the bar directly after this sibling")

	return cmd
}
