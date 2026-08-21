package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/B4Dmonkey/bit-pro/daemon"
	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/spf13/cobra"
)

const statusCmdUse = "status"

func newStatusCmd(lc daemon.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   statusCmdUse,
		Short: "Report whether the background daemon is running",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			state, pid, err := daemon.Status(cmd.Context(), lc)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()

			if state == daemon.StateRunning {
				fmt.Fprintf(out, "running (pid %d)\n", pid)
			} else {
				fmt.Fprintln(out, state)
			}

			return printProjectCounts(cmd.Context(), out)
		},
	}
}

func printProjectCounts(ctx context.Context, out io.Writer) error {
	sqlDB, err := db.Open()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	projects, err := orm.New(sqlDB).ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("listing projects: %w", err)
	}

	if len(projects) == 0 {
		return nil
	}

	fmt.Fprintln(out)

	for _, p := range projects {
		fmt.Fprintf(out, "  %s\tbacklog:%d\ttodo:%d\tdone:%d\n", p.Code, p.Backlog, p.Todo, p.Done)
	}

	return nil
}
