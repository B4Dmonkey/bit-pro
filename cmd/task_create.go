package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type taskFrontmatter struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Status string `yaml:"status"`
}

func newTaskCreateCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			if err := os.MkdirAll(tasksDir, 0o755); err != nil {
				return err
			}

			id := cfg.Prefix + "-1"
			frontmatter, err := yaml.Marshal(taskFrontmatter{
				ID:     id,
				Title:  args[0],
				Status: "todo",
			})
			if err != nil {
				return fmt.Errorf("encoding task frontmatter: %w", err)
			}

			content := "---\n" + string(frontmatter) + "---\n" + description
			path := filepath.Join(tasksDir, id+".md")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "task description (body content)")
	return cmd
}
