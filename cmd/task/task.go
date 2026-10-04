package task

import (
	"fmt"
	"os"

	"github.com/B4Dmonkey/bit-pro/project"
	taskstore "github.com/B4Dmonkey/bit-pro/task"
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

func openStore(cmd *cobra.Command) (*taskstore.Store, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting the working directory: %w", err)
	}

	return project.OpenStore(cmd.Context(), wd)
}
