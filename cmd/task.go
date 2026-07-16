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
	taskCmd.AddCommand(newTaskReadCmd())
	taskCmd.AddCommand(newTaskUpdateCmd())
	return taskCmd
}
