package cmd

import "github.com/spf13/cobra"

const taskCmdUse = "task"

func newTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   taskCmdUse,
		Short: "Manage tasks",
	}
	taskCmd.AddCommand(newTaskCreateCmd())
	taskCmd.AddCommand(newTaskListCmd())
	taskCmd.AddCommand(newTaskReadCmd())
	taskCmd.AddCommand(newTaskUpdateCmd())
	taskCmd.AddCommand(newTaskMoveCmd())
	taskCmd.AddCommand(newTaskCompleteCmd())
	taskCmd.AddCommand(newTaskDeleteCmd())

	return taskCmd
}
