package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskListCmd() *cobra.Command {
	var parent string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store := task.New(bitDir)

			var tasks []*task.Task
			var err error
			if parent == "" {
				tasks, err = store.List()
			} else {
				tasks, err = store.Children(parent)
			}
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			for _, t := range tasks {
				phase := ""
				if t.Phase != 0 {
					phase = fmt.Sprintf("phase %d — %s", t.Phase, t.PhaseLabel)
				}
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Title, phase)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&parent, "parent", "p", "", "list only this task's direct bars")
	return cmd
}
