package cmd

import (
	"github.com/spf13/cobra"
)

func newTaskUpdateCmd() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := loadTask(args[0])
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("title") {
				t.Title = title
			}

			return t.save()
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "", "new task title")
	return cmd
}
