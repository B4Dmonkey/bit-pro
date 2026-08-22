package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/bitdir"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newFeedbackAddCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "add <track>",
		Short: "Record a feedback note against a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := task.New(bitdir.Current()).AddNote(args[0], description)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)

			return err
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "note body content")

	return cmd
}
