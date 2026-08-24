package task

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/bitdir"
	taskstore "github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var (
		description string
		parent      string
		after       string
		phase       int
		phaseLabel  string
	)

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, parent, after, description, phase, phaseLabel)
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "task description (body content)")
	cmd.Flags().StringVarP(&parent, "parent", "p", "", "parent task ID (mints a dotted child ID)")
	cmd.Flags().StringVar(&after, "after", "", "sibling bar ID to insert the new bar after in the parent's order")
	cmd.Flags().IntVar(&phase, "phase", 0, "scope phase this step serves")
	cmd.Flags().StringVar(&phaseLabel, "phase-label", "", "human-readable label for the phase")

	return cmd
}

func runCreate(
	cmd *cobra.Command, args []string,
	parent, after, description string,
	phase int, phaseLabel string,
) error {
	t, err := taskstore.New(bitdir.Current()).Create(taskstore.CreateParams{
		Title:      args[0],
		Body:       description,
		Parent:     parent,
		After:      after,
		Phase:      phase,
		PhaseLabel: phaseLabel,
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), t.ID)

	return err
}
