package task

import (
	"github.com/spf13/cobra"
)

func newCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "File a track and its bars as completed, with no landing check (manual override)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore(cmd)
			if err != nil {
				return err
			}

			return s.Complete(args[0])
		},
	}
}
