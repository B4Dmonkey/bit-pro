package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <id>",
		Short: "Show a task's full content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := task.New(bitDir).Load(args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s\t%s\t%s\n\n", t.ID, t.Status, t.Title)
			fmt.Fprint(out, t.Body)
			return nil
		},
	}
}
