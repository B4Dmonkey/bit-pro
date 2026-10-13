package task

import (
	"context"
	"fmt"

	"github.com/B4Dmonkey/bit-pro/project"
	taskstore "github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var parent string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tasks, err := listTasks(cmd.Context(), parent)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()

			for _, t := range tasks {
				approved := ""
				if t.Approved {
					approved = "approved"
				}

				phase := ""
				if t.Phase != 0 {
					phase = fmt.Sprintf("phase %d — %s", t.Phase, t.PhaseLabel)
				}

				fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Title, approved, phase)
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&parent, "parent", "p", "", "list only this task's direct bars")

	return cmd
}

func listTasks(ctx context.Context, parent string) ([]*taskstore.Task, error) {
	stores, err := project.OpenAll(ctx)
	if err != nil {
		return nil, err
	}

	if parent == "" {
		return stores.List()
	}

	store, err := stores.For(parent)
	if err != nil {
		return nil, err
	}

	return store.Children(parent)
}
