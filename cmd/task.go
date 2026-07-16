package cmd

import "github.com/spf13/cobra"

const tasksDir = ".bit/tasks"

func newTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}
	taskCmd.AddCommand(newTaskCreateCmd())
	taskCmd.AddCommand(newTaskListCmd())
	return taskCmd
}
