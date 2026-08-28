package daemon

import (
	"context"
	"log/slog"
	"path/filepath"
	"slices"
	"time"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	msgStarted        = "started"
	msgStopped        = "stopped"
	msgNotDispatching = "not dispatching"
)

func Loop(
	ctx context.Context,
	queries *orm.Queries,
	log *slog.Logger,
	interval time.Duration,
	run claude.DirRunner,
	bin string,
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
			Tick(ctx, queries, log, run, bin)
		}
	}
}

func Tick(ctx context.Context, queries *orm.Queries, log *slog.Logger, run claude.DirRunner, bin string) {
	projects, err := queries.ListProjects(ctx)
	if err != nil {
		log.Error("listing projects", "err", err)

		return
	}

	live, err := claude.Agents(ctx, run, bin)

	mayDispatch := err == nil
	if err != nil {
		log.Error("listing live sessions", "err", err)
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

		if !mayDispatch {
			continue
		}

		if i := slices.IndexFunc(live, func(a claude.Agent) bool { return a.Under(p.Path) }); i >= 0 {
			log.Info(msgNotDispatching, "project", p.Code, "session", live[i].Name, "cwd", live[i].Cwd)

			continue
		}

		dispatch(ctx, queries, log, run, bin, p, store)
	}
}

func dispatch(
	ctx context.Context,
	queries *orm.Queries,
	log *slog.Logger,
	run claude.DirRunner, bin string,
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

	if dropped(ctx, queries, log, p, head.ID, bar) {
		return
	}

	name, ok := worktreeFor(log, store, p, bar)
	if !ok {
		return
	}

	if err := claude.Spawn(ctx, run, bin, p.Path, name, bar.ID); err != nil {
		log.Error("dispatching the queued bar", "project", p.Code, "bar", bar.ID, "err", err)

		return
	}

	log.Info("dispatched", "project", p.Code, "bar", bar.ID, "worktree", name)

	agents, err := claude.Agents(ctx, run, bin)
	if err != nil {
		log.Error("confirming the dispatched session", "project", p.Code, "bar", bar.ID, "err", err)

		return
	}

	if !slices.ContainsFunc(agents, func(a claude.Agent) bool { return a.Name == name }) {
		log.Warn("dispatched session not visible yet", "project", p.Code, "bar", bar.ID, "worktree", name)

		return
	}

	if err := queries.DeleteQueueRow(ctx, head.ID); err != nil {
		log.Error("dequeueing the dispatched bar", "project", p.Code, "bar", bar.ID, "err", err)
	}
}

func dropped(
	ctx context.Context,
	queries *orm.Queries,
	log *slog.Logger,
	p orm.Project,
	rowID int64,
	bar *task.Task,
) bool {
	if bar.Status != task.StatusDone && bar.Approved {
		return false
	}

	log.Info("dropping a queued bar the ledger says must not run",
		"project", p.Code, "bar", bar.ID, "status", bar.Status, "approved", bar.Approved)

	if err := queries.DeleteQueueRow(ctx, rowID); err != nil {
		log.Error("dropping the queue row", "project", p.Code, "bar", bar.ID, "err", err)
	}

	return true
}

func worktreeFor(
	log *slog.Logger,
	store *task.Store,
	p orm.Project,
	bar *task.Task,
) (name string, ok bool) {
	parent, ok := task.ParentID(bar.ID)
	if !ok {
		log.Warn("queued target is not a bar", "project", p.Code, "bar", bar.ID)

		return "", false
	}

	track, err := store.Load(parent)
	if err != nil {
		log.Warn("loading the bar's track", "project", p.Code, "track", parent, "err", err)

		return "", false
	}

	return claude.WorktreeName(track.ID, track.Title), true
}
