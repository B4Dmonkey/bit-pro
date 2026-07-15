package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var prefix string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the .bit/ directory bit uses to track this project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(".bit", 0o755); err != nil {
				return err
			}
			if prefix == "" {
				fmt.Fprint(cmd.OutOrStdout(), "Task ID prefix: ")
				line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if err != nil && line == "" {
					return fmt.Errorf("reading task ID prefix: %w", err)
				}
				prefix = strings.TrimSpace(line)
			}
			return saveConfig(&Config{Prefix: prefix})
		},
	}
	cmd.Flags().StringVar(&prefix, "prefix", "", "task ID prefix for this project (e.g. BIT)")
	return cmd
}
