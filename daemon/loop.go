package daemon

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

func Tick(ctx context.Context, queries *orm.Queries, log *slog.Logger) {
	projects, err := queries.ListProjects(ctx)
	if err != nil {
		log.Error("listing projects", "err", err)

		return
	}

	for _, p := range projects {
		counts, err := task.New(filepath.Join(p.Path, ".bit")).Counts()
		if err != nil {
			log.Warn("reading counts", "project", p.Code, "path", p.Path, "err", err)

			continue
		}

		if err := queries.UpdateProjectCounts(ctx, orm.UpdateProjectCountsParams{
			Backlog:   int64(counts.Backlog),
			Todo:      int64(counts.Todo),
			Done:      int64(counts.Done),
			Completed: int64(counts.Completed),
			ID:        p.ID,
		}); err != nil {
			log.Error("updating project counts", "project", p.Code, "err", err)
		}
	}
}
