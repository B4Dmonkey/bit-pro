package cmd

import (
	"context"
	"os"

	"github.com/B4Dmonkey/bit-pro/bitdir"
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
			s := task.New(bitdir.Current())

			tasks, err := s.List()
			if err != nil {
				return err
			}

			var (
				enqueue   func(targetIDs []string, targetTyp string) error
				listQueue func() ([]string, error)
			)

			if sqlDB, err := db.Open(); err == nil {
				defer sqlDB.Close()

				enqueue, listQueue = queueFuncs(cmd.Context(), orm.New(sqlDB))
			}

			return tui.Run(tui.New(tasks).
				WithReload(s.List).
				WithApprove(func(id string, approved bool) error { return s.SetApproved(id, approved) }).
				WithEnqueue(enqueue).
				WithListQueue(listQueue))
		},
	}
}

func queueFuncs(ctx context.Context, q *orm.Queries) (
	enqueue func(targetIDs []string, targetTyp string) error,
	listQueue func() ([]string, error),
) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, nil
	}

	project, err := q.GetProjectByPath(ctx, wd)
	if err != nil {
		return nil, nil
	}

	enqueue = func(targetIDs []string, targetTyp string) error {
		for _, id := range targetIDs {
			if err := q.EnqueueTask(ctx, orm.EnqueueTaskParams{
				ProjectID: project.ID,
				TargetID:  id,
				TargetTyp: targetTyp,
			}); err != nil {
				return err
			}
		}

		return nil
	}

	listQueue = func() ([]string, error) {
		rows, err := q.ListQueueByProject(ctx, project.ID)
		if err != nil {
			return nil, err
		}

		ids := make([]string, len(rows))
		for i, r := range rows {
			ids[i] = r.TargetID
		}

		return ids, nil
	}

	return enqueue, listQueue
}
