package task

import (
	"github.com/B4Dmonkey/bit-pro/bitdir"
	taskstore "github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var title, description, status, phaseLabel string

	var phase int

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := taskstore.New(bitdir.Current())

			t, err := s.Load(args[0])
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("title") {
				t.Title = title
			}

			if cmd.Flags().Changed("description") {
				t.Body = description
			}

			if cmd.Flags().Changed("status") {
				t.Status = status
			}

			if cmd.Flags().Changed("phase") {
				t.Phase = phase
			}

			if cmd.Flags().Changed("phase-label") {
				t.PhaseLabel = phaseLabel
			}

			anyChanged := cmd.Flags().Changed("title") ||
				cmd.Flags().Changed("description") ||
				cmd.Flags().Changed("phase") ||
				cmd.Flags().Changed("phase-label")
			sentBack := cmd.Flags().Changed("status") && status == taskstore.StatusTodo

			if t.Approved && (anyChanged || sentBack) {
				t.Approved = false
			}

			return s.Save(t)
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "", "new task title")
	cmd.Flags().StringVarP(&description, "description", "d", "", "new task description")
	cmd.Flags().StringVarP(&status, "status", "s", "", "new task status")
	cmd.Flags().IntVar(&phase, "phase", 0, "new scope phase this step serves")
	cmd.Flags().StringVar(&phaseLabel, "phase-label", "", "new human-readable label for the phase")

	return cmd
}
