package daemon

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	msgStarted = "started"
	msgStopped = "stopped"
)

func Loop(
	ctx context.Context,
	queries *orm.Queries,
	log *slog.Logger,
	interval time.Duration,
	run claude.DirRunner,
) error {
	log.Info(msgStarted)
	defer log.Info(msgStopped)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			log.Debug("tick")
			Tick(ctx, queries, log, run)
		}
	}
}

func Tick(ctx context.Context, queries *orm.Queries, log *slog.Logger, run claude.DirRunner) {
	projects, err := queries.ListProjects(ctx)
	if err != nil {
		log.Error("listing projects", "err", err)

		return
	}

	for _, p := range projects {
		store := task.New(filepath.Join(p.Path, ".bit"))

		counts, err := store.Counts()
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

		dispatch(ctx, queries, log, run, p, store)
	}
}

func dispatch(
	ctx context.Context,
	queries *orm.Queries,
	log *slog.Logger,
	run claude.DirRunner,
	p orm.Project,
	store *task.Store,
) {
	rows, err := queries.ListQueueByProject(ctx, p.ID)
	if err != nil {
		log.Error("listing the queue", "project", p.Code, "err", err)

		return
	}

	if len(rows) == 0 {
		return
	}

	head := rows[0]

	bar, err := store.Load(head.TargetID)
	if err != nil {
		log.Warn("loading the queued bar", "project", p.Code, "bar", head.TargetID, "err", err)

		return
	}

	parent, ok := task.ParentID(bar.ID)
	if !ok {
		log.Warn("queued target is not a bar", "project", p.Code, "bar", bar.ID)

		return
	}

	track, err := store.Load(parent)
	if err != nil {
		log.Warn("loading the bar's track", "project", p.Code, "track", parent, "err", err)

		return
	}

	name := claude.WorktreeName(track.ID, track.Title)
	if err := claude.Spawn(ctx, run, p.Path, name, bar.ID); err != nil {
		log.Error("dispatching the queued bar", "project", p.Code, "bar", bar.ID, "err", err)

		return
	}

	log.Info("dispatched", "project", p.Code, "bar", bar.ID, "worktree", name)
}
