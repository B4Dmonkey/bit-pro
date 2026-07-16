package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
)

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

			id, err := nextTaskID(cfg.Prefix)
			if err != nil {
				return err
			}
			t := &Task{
				ID:     id,
				Title:  args[0],
				Status: "todo",
				Body:   description,
			}
			return t.save()
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "task description (body content)")
	return cmd
}

func nextTaskID(prefix string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(tasksDir, prefix+"-*.md"))
	if err != nil {
		return "", fmt.Errorf("scanning %s for existing task IDs: %w", tasksDir, err)
	}
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `-(\d+)\.md$`)
	max := 0
	for _, m := range matches {
		if sub := re.FindStringSubmatch(filepath.Base(m)); sub != nil {
			if n, _ := strconv.Atoi(sub[1]); n > max {
				max = n
			}
		}
	}
	return fmt.Sprintf("%s-%d", prefix, max+1), nil
}
