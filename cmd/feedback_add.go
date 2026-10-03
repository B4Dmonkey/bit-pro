package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/B4Dmonkey/bit-pro/task"
)

func newFeedbackAddCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "add <track>",
		Short: "Record a feedback note against a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore(cmd)
			if err != nil {
				return err
			}

			path, err := s.AddNote(args[0], description, task.Commit{})
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
