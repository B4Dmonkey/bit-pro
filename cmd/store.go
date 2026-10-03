package cmd

import (
	"fmt"
	"os"

	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func openStore(cmd *cobra.Command) (*task.Store, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting the working directory: %w", err)
	}

	return project.OpenStore(cmd.Context(), wd)
}
