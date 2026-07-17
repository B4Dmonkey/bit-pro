package cmd

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskCreateCmd() *cobra.Command {
	var description string
	var parent string
	var phase int
	var phaseLabel string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s := task.New(bitDir)

			var id string
			var err error
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

			return s.Save(&task.Task{
				ID:         id,
				Title:      args[0],
				Status:     "todo",
				Phase:      phase,
				PhaseLabel: phaseLabel,
				Body:       description,
			})
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "task description (body content)")
	cmd.Flags().StringVarP(&parent, "parent", "p", "", "parent task ID (mints a dotted child ID)")
	cmd.Flags().IntVar(&phase, "phase", 0, "scope phase this step serves")
	cmd.Flags().StringVar(&phaseLabel, "phase-label", "", "human-readable label for the phase")
	return cmd
}
