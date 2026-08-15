package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskReadCmd() *cobra.Command {
	var bodyOnly bool

	cmd := &cobra.Command{
		Use:   "read <id>",
		Short: "Show a task's full content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := task.New(bitDir).Load(args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if bodyOnly {
				fmt.Fprint(out, t.Body)

				return nil
			}

			fmt.Fprintf(out, "%s\t%s\t%s", t.ID, t.Status, t.Title)

			if t.Phase != 0 {
				fmt.Fprintf(out, "\tphase %d — %s", t.Phase, t.PhaseLabel)
			}

			fmt.Fprint(out, "\n\n")
			fmt.Fprint(out, t.Body)

			return nil
		},
	}
	cmd.Flags().BoolVar(&bodyOnly, "body", false, "print only the task body, without the summary header")

	return cmd
}
