package cmd

import (
	"context"
	"os"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/B4Dmonkey/bit-pro/tui"
	"github.com/spf13/cobra"
)

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Browse tasks in a terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := task.New(bitDir)

			tasks, err := s.List()
			if err != nil {
				return err
			}

			var enqueue func(targetID, targetTyp string) error

			if sqlDB, err := db.Open(); err == nil {
				defer sqlDB.Close()

				enqueue = enqueueFunc(cmd.Context(), orm.New(sqlDB))
			}

			return tui.Run(tui.New(tasks).
				WithReload(s.List).
				WithApprove(func(id string, approved bool) error { return s.SetApproved(id, approved) }).
				WithEnqueue(enqueue))
		},
	}
}

func enqueueFunc(ctx context.Context, q *orm.Queries) func(targetID, targetTyp string) error {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}

	project, err := q.GetProjectByPath(ctx, wd)
	if err != nil {
		return nil
	}

	return func(targetID, targetTyp string) error {
		return q.EnqueueTask(ctx, orm.EnqueueTaskParams{
			ProjectID: project.ID,
			TargetID:  targetID,
			TargetTyp: targetTyp,
		})
	}
}
