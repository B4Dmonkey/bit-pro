package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

				parts := strings.SplitN(string(data), "---", 3)
				if len(parts) != 3 {
					return fmt.Errorf("parsing %s: expected frontmatter delimited by \"---\"", path)
				}

				var fm taskFrontmatter
				if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
					return fmt.Errorf("parsing frontmatter in %s: %w", path, err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", fm.ID, fm.Status, fm.Title)
			}
			return nil
		},
	}
	return cmd
}
