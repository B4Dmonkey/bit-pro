package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/B4Dmonkey/bit-pro/tui"
	"github.com/B4Dmonkey/bit-pro/usage"
	"github.com/spf13/cobra"
)

const uiCmdUse = "ui"

func newUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   uiCmdUse,
		Short: "Browse tasks in a terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			stores, err := project.OpenAll(cmd.Context())
			if err != nil {
				return err
			}

			tasks, err := stores.List()
			if err != nil {
				return err
			}

			return tui.Run(tui.New(tasks).
				WithReload(stores.List).
				WithApprove(func(id string, approved bool) error {
					s, err := stores.For(id)
					if err != nil {
						return err
					}

					return s.SetApproved(id, approved)
				}).
				WithCost(costFor(cmd.Context(), stores)))
		},
	}
}

func costFor(ctx context.Context, stores *project.Stores) func(track string) (string, error) {
	return func(track string) (string, error) {
		sqlDB, err := db.Open()
		if err != nil {
			return "", err
		}
		defer sqlDB.Close()

		q := orm.New(sqlDB)

		projects, err := project.Load(ctx, q)
		if err != nil {
			return "", err
		}

		code, _, _ := strings.Cut(task.NormalizeID(track), "-")

		p, ok := project.ByCode(projects, code)
		if !ok {
			return "", fmt.Errorf("track %s: %w", track, project.ErrNotRegistered)
		}

		store, err := stores.For(track)
		if err != nil {
			return "", err
		}

		status, err := trackStatus(store, track)
		if err != nil {
			return "", err
		}

		bars, err := trackBars(store, track)
		if err != nil {
			return "", err
		}

		md, rows, err := usage.Build(ctx, q, p.ID, task.NormalizeID(track), status, bars)
		if err != nil {
			return "", err
		}

		if rows == 0 {
			return "", nil
		}

		return md, nil
	}
}
