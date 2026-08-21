package cmd

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const serveCmdUse = "serve"

var serveTick = 10 * time.Second

func writeCounts(ctx context.Context, queries *orm.Queries, log *slog.Logger) {
	projects, err := queries.ListProjects(ctx)
	if err != nil {
		log.Error("listing projects", "err", err)

		return
	}

	for _, p := range projects {
		if err := queries.UpdateProjectCounts(ctx, orm.UpdateProjectCountsParams{
			Backlog:   1,
			Todo:      0,
			Done:      0,
			Completed: 0,
			ID:        p.ID,
		}); err != nil {
			log.Error("updating project counts", "project", p.Code, "err", err)
		}
	}
}

func newHandler(w io.Writer, level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}

	if f, ok := w.(*os.File); ok {
		if info, err := f.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
			return slog.NewTextHandler(w, opts)
		}
	}

	return slog.NewJSONHandler(w, opts)
}

func newServeCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   serveCmdUse,
		Short: "Run the automation loop in the foreground",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}

			log := slog.New(newHandler(cmd.OutOrStdout(), level))

			sqlDB, err := db.Open()
			if err != nil {
				return err
			}

			defer sqlDB.Close()

			queries := orm.New(sqlDB)

			log.Info("started")
			defer log.Info("stopped")

			ticker := time.NewTicker(serveTick)
			defer ticker.Stop()

			ctx := cmd.Context()

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					log.Debug("tick")
					writeCounts(ctx, queries, log)
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Log each tick of the loop")

	return cmd
}
