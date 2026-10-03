package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

const removeCmdUse = "remove"

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   removeCmdUse,
		Short: "Archive this project's open work and remove it from the registry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting the working directory: %w", err)
			}

			p, err := project.Find(cmd.Context(), wd)
			if err != nil {
				return err
			}

			s, err := project.StoreFor(p)
			if err != nil {
				return err
			}

			tasks, err := s.List()
			if err != nil {
				return err
			}

			printOutstanding(cmd, tasks)

			if !confirmRemove(cmd, p) {
				fmt.Fprintln(cmd.OutOrStdout(), "not removed")
				return nil
			}

			for _, t := range tasks {
				if strings.Contains(t.ID, ".") {
					continue
				}

				if err := s.Relocate(t.ID, true); err != nil {
					return err
				}
			}

			sqlDB, err := db.Open()
			if err != nil {
				return err
			}
			defer sqlDB.Close()

			params := orm.SetProjectRemovedParams{Removed: 1, ID: p.ID}
			if err := orm.New(sqlDB).SetProjectRemoved(cmd.Context(), params); err != nil {
				return fmt.Errorf("removing %s: %w", p.Path, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "removed %s %s\n", p.Code, p.Path)

			return nil
		},
	}
}

func printOutstanding(cmd *cobra.Command, tasks []*task.Task) {
	out := cmd.OutOrStdout()
	header := false

	for _, t := range tasks {
		if t.Status == task.StatusDone {
			continue
		}

		if !header {
			fmt.Fprintln(out, "Outstanding work:")

			header = true
		}

		fmt.Fprintf(out, "  %s\t%s\t%s\n", t.ID, t.Status, t.Title)
	}
}

func confirmRemove(cmd *cobra.Command, p project.Project) bool {
	fmt.Fprintf(cmd.OutOrStdout(), "Remove %s (%s)? [y/N]: ", p.Code, p.Path)

	line, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
