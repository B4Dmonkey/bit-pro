package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/spf13/cobra"
)

const addCmdUse = "add"

func newAddCmd(run claude.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "add <path>",
		Short: "Enroll a project in the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := project.CanonicalPath(args[0])
			if err != nil {
				return err
			}

			sqlDB, err := db.Open()
			if err != nil {
				return err
			}
			defer sqlDB.Close()

			queries := orm.New(sqlDB)

			projects, err := project.Load(cmd.Context(), queries)
			if err != nil {
				return err
			}

			if registered(projects, path) {
				fmt.Fprintln(cmd.OutOrStdout(), "already added")
				return nil
			}

			if _, err := os.Stat(filepath.Join(path, ".bit")); err == nil {
				return fmt.Errorf("%s: %w", path, project.ErrNeedsMigrate)
			}

			typed, err := readProjectCode(cmd)
			if err != nil {
				return err
			}

			code, err := project.ValidateCode(typed)
			if err != nil {
				return err
			}

			if err := writeClaudeWiring(cmd, run, path); err != nil {
				return err
			}

			params := orm.CreateProjectParams{Path: path, Code: code}
			if err := queries.CreateProject(cmd.Context(), params); err != nil {
				return fmt.Errorf("registering %s: %w", path, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "added %s %s\n", code, path)

			return nil
		},
	}
}

func registered(projects []project.Project, path string) bool {
	for _, p := range projects {
		if strings.EqualFold(p.Path, path) {
			return true
		}
	}

	return false
}

func readProjectCode(cmd *cobra.Command) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), "Project code: ")

	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("reading project code: %w", err)
	}

	return strings.TrimSpace(line), nil
}
