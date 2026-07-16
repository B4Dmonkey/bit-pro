package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newTaskReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <id>",
		Short: "Show a task's full content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := filepath.Join(tasksDir, args[0]+".md")
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

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s\t%s\t%s\n\n", fm.ID, fm.Status, fm.Title)
			fmt.Fprint(out, parts[2])
			return nil
		},
	}
	return cmd
}
