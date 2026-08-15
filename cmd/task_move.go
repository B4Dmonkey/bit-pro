package cmd

import (
	"errors"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskMoveCmd() *cobra.Command {
	var before, after string

	cmd := &cobra.Command{
		Use:   "move <bar>",
		Short: "Resequence a bar within its track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hasBefore := cmd.Flags().Changed("before")

			hasAfter := cmd.Flags().Changed("after")
			if hasBefore == hasAfter {
				return errors.New("specify exactly one of --before or --after")
			}

			anchor := after
			if hasBefore {
				anchor = before
			}

			return task.New(bitDir).Move(args[0], anchor, hasBefore)
		},
	}
	cmd.Flags().StringVar(&before, "before", "", "move the bar directly before this sibling")
	cmd.Flags().StringVar(&after, "after", "", "move the bar directly after this sibling")

	return cmd
}
