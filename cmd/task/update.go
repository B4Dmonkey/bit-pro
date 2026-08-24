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

			var p taskstore.Patch

			if cmd.Flags().Changed("title") {
				p.Title = &title
			}

			if cmd.Flags().Changed("description") {
				p.Body = &description
			}

			if cmd.Flags().Changed("status") {
				p.Status = &status
			}

			if cmd.Flags().Changed("phase") {
				p.Phase = &phase
			}

			if cmd.Flags().Changed("phase-label") {
				p.PhaseLabel = &phaseLabel
			}

			_, err := s.Update(args[0], p)

			return err
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "", "new task title")
	cmd.Flags().StringVarP(&description, "description", "d", "", "new task description")
	cmd.Flags().StringVarP(&status, "status", "s", "", "new task status")
	cmd.Flags().IntVar(&phase, "phase", 0, "new scope phase this step serves")
	cmd.Flags().StringVar(&phaseLabel, "phase-label", "", "new human-readable label for the phase")

	return cmd
}
