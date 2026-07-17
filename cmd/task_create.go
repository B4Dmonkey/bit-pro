package cmd

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newTaskCreateCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s := task.New(bitDir)

			cfg, err := s.Config()
			if err != nil {
				return err
			}

			id, err := s.NextID(cfg.Prefix)
			if err != nil {
				return err
			}

			return s.Save(&task.Task{
				ID:     id,
				Title:  args[0],
				Status: "todo",
				Body:   description,
			})
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "task description (body content)")
	return cmd
}
