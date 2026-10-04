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

			if p, ok := project.ByPath(projects, path); ok {
				return addExisting(cmd, queries, p)
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

			if p, ok := project.ByCode(projects, code); ok && p.Removed {
				return fmt.Errorf("%s at %s: %w; run `bp add` there to revive it, or pick another code",
					p.Code, p.Path, project.ErrCodeRemoved)
			}

			params := orm.CreateProjectParams{Path: path, Code: code}
			if err := queries.CreateProject(cmd.Context(), params); err != nil {
				return fmt.Errorf("registering %s: %w", path, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "added %s %s\n", code, path)

			return setUpClaude(cmd, run)
		},
	}
}

func addExisting(cmd *cobra.Command, queries *orm.Queries, p project.Project) error {
	if !p.Removed {
		fmt.Fprintln(cmd.OutOrStdout(), "already added")
		return nil
	}

	params := orm.SetProjectRemovedParams{Removed: 0, ID: p.ID}
	if err := queries.SetProjectRemoved(cmd.Context(), params); err != nil {
		return fmt.Errorf("reviving %s: %w", p.Path, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "revived %s %s\n", p.Code, p.Path)

	return nil
}

func readProjectCode(cmd *cobra.Command) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), "Project code: ")

	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("reading project code: %w", err)
	}

	return strings.TrimSpace(line), nil
}
