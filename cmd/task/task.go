package task

import (
	"github.com/spf13/cobra"
)

const CmdUse = "task"

func NewCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   CmdUse,
		Short: "Manage tasks",
	}
	taskCmd.AddCommand(newListCmd())
	taskCmd.AddCommand(newReadCmd())

	return taskCmd
}
