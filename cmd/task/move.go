package task

import (
	"github.com/spf13/cobra"
)

func newMoveCmd() *cobra.Command {
	var before, after string

	cmd := &cobra.Command{
		Use:   "move <bar>",
		Short: "Resequence a bar within its track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore(cmd)
			if err != nil {
				return err
			}

			return s.Move(args[0], before, after)
		},
	}
	cmd.Flags().StringVar(&before, "before", "", "move the bar directly before this sibling")
	cmd.Flags().StringVar(&after, "after", "", "move the bar directly after this sibling")

	return cmd
}
