package cmd

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

const addCmdUse = "add"

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <path>",
		Short: "Enroll a project in the registry the daemon watches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolving %s: %w", args[0], err)
			}

			sqlDB, err := db.Open()
			if err != nil {
				return err
			}
			defer sqlDB.Close()

			queries := orm.New(sqlDB)

			enrolled, err := queries.ProjectExists(cmd.Context(), abs)
			if err != nil {
				return fmt.Errorf("looking up %s: %w", abs, err)
			}

			if enrolled {
				fmt.Fprintln(cmd.OutOrStdout(), "already added")
				return nil
			}

			var existing string
			if cfg, err := task.New(filepath.Join(abs, ".bit")).Config(); err == nil {
				existing = cfg.Prefix
			}

			code, err := readProjectCode(cmd, existing)
			if err != nil {
				return err
			}

			params := orm.CreateProjectParams{Path: abs, Code: code}
			if err := queries.CreateProject(cmd.Context(), params); err != nil {
				return fmt.Errorf("registering %s: %w", abs, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "added %s %s\n", code, abs)

			return nil
		},
	}
}

func readProjectCode(cmd *cobra.Command, existing string) (string, error) {
	if existing != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Project code (%s): ", existing)
	} else {
		fmt.Fprint(cmd.OutOrStdout(), "Project code: ")
	}

	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("reading project code: %w", err)
	}

	code := strings.TrimSpace(line)
	if code == "" {
		code = existing
	}

	return code, nil
}
