package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

const initCmdUse = "init"

func newInitCmd(run claude.Runner) *cobra.Command {
	var prefix string

	cmd := &cobra.Command{
		Use:   initCmdUse,
		Short: "Create the .bit/ directory bit uses to track this project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if prefix == "" {
				var err error

				prefix, err = readInteractivePrefix(cmd)
				if err != nil {
					return err
				}
			}

			if prefix == "" {
				return errors.New("task ID prefix cannot be empty")
			}

			if err := task.New(bitDir).SaveConfig(&task.Config{Prefix: prefix}); err != nil {
				return err
			}

			return writeClaudeWiring(cmd, run, ".")
		},
	}
	cmd.Flags().StringVar(&prefix, "prefix", "", "task ID prefix for this project (e.g. BIT)")

	return cmd
}

func writeClaudeWiring(cmd *cobra.Command, run claude.Runner, dir string) error {
	if err := claude.WriteSettings(filepath.Join(dir, claudeDir, "settings.json")); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Bringing the bit plugin current...")

	return claude.SyncPlugin(cmd.Context(), run)
}

func readInteractivePrefix(cmd *cobra.Command) (string, error) {
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
		return "", fmt.Errorf("reading task ID prefix: %w", err)
	}

	p := strings.TrimSpace(line)
	if p == "" {
		p = existing
	}

	return p, nil
}
