package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
)

func newTaskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			matches, err := filepath.Glob(filepath.Join(tasksDir, "*.md"))
			if err != nil {
				return fmt.Errorf("scanning %s for tasks: %w", tasksDir, err)
			}
			slices.Sort(matches)

			for _, path := range matches {
				data, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("reading %s: %w", path, err)
				}

				t, err := parseTask(data)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", path, err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", t.ID, t.Status, t.Title)
			}
			return nil
		},
	}
	return cmd
}
