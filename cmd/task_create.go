package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

const statusTodo = "todo"

func newTaskCreateCmd() *cobra.Command {
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
			return runTaskCreate(cmd, args, parent, after, description, phase, phaseLabel)
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "task description (body content)")
	cmd.Flags().StringVarP(&parent, "parent", "p", "", "parent task ID (mints a dotted child ID)")
	cmd.Flags().StringVar(&after, "after", "", "sibling bar ID to insert the new bar after in the parent's order")
	cmd.Flags().IntVar(&phase, "phase", 0, "scope phase this step serves")
	cmd.Flags().StringVar(&phaseLabel, "phase-label", "", "human-readable label for the phase")

	return cmd
}

func runTaskCreate(
	cmd *cobra.Command, args []string,
	parent, after, description string,
	phase int, phaseLabel string,
) error {
	s := task.New(bitDir)

	var (
		id  string
		err error
	)

	if parent != "" {
		id, err = s.NextChildID(parent)
	} else {
		var cfg *task.Config

		cfg, err = s.Config()
		if err != nil {
			return err
		}

		id, err = s.NextID(cfg.Prefix)
	}

	if err != nil {
		return err
	}

	if after != "" {
		if err := s.InsertAfter(parent, id, after); err != nil {
			return err
		}
	}

	if err := s.Save(&task.Task{
		ID:         id,
		Title:      args[0],
		Status:     statusTodo,
		Phase:      phase,
		PhaseLabel: phaseLabel,
		Body:       description,
	}); err != nil {
		return err
	}

	if parent != "" && after == "" {
		if err := s.AppendToOrder(parent, id); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), id)

	return err
}
