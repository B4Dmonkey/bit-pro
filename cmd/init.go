package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/B4Dmonkey/bit-pro/assets"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var prefix string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the .bit/ directory bit uses to track this project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if prefix == "" {
				var existing string
				if cfg, err := task.New(bitDir).Config(); err == nil {
					existing = cfg.Prefix
				}
				if existing != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Task ID prefix (%s): ", existing)
				} else {
					fmt.Fprint(cmd.OutOrStdout(), "Task ID prefix: ")
				}
				line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if err != nil && line == "" {
					return fmt.Errorf("reading task ID prefix: %w", err)
				}
				prefix = strings.TrimSpace(line)
				if prefix == "" {
					prefix = existing
				}
			}
			if prefix == "" {
				return errors.New("task ID prefix cannot be empty")
			}
			if err := task.New(bitDir).SaveConfig(&task.Config{Prefix: prefix}); err != nil {
				return err
			}
			return assets.Seed(claudeDir)
		},
	}
	cmd.Flags().StringVar(&prefix, "prefix", "", "task ID prefix for this project (e.g. BIT)")
	return cmd
}
