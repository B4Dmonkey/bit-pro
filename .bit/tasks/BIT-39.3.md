---
id: BIT-39.3
title: Tick dispatches the queued bar
status: done
approved: true
phase: 2
phase_label: Bar runs unattended
---
## **Verse 2**

`Tick` grows a dispatch step: for a project with a queued bar, it spawns a background Claude
session in that project's directory running `/bit:do <BAR>` as `bit:bot-dev`. Forced by a test
that hands `Tick` a fake runner and asserts the argv — the counts-only `Tick` from Verse 1
records no call at all.

## Preflight
- [ ] `./clear-queue.sh` — it prints `cleared N queue rows`. Then
  `sqlite3 "${XDG_DATA_HOME:-$HOME/.local/share}/bit-pro/bit.db" 'SELECT count(*) FROM queue;'`
  returns `0`.

  This bar is where the queue gets its first consumer, and the live table still holds rows enqueued
  during earlier tracks — bars that are long done and whose files `bp task complete` has since
  moved to `.bit/completed/`. `Store.Load` reads `.bit/tasks/` only, so such a row errors on load,
  never dispatches, and never gets deleted either: dequeue lands in BIT-39.5 and only on a
  confirmed spawn, and the ledger check in BIT-39.7 runs *after* the load. Sitting at the head of a
  FIFO, one of them wedges the project's queue against everything queued behind it. Note the script
  clears every row in the table, for all registered projects, not just this one.

## Scope
- `claude/dispatch.go` — new file: `DirRunner` (the second runner shape), `ExecDirRunner`,
  `WorktreeName`, `Spawn`.
- `daemon/loop.go` — `Tick` and `Loop` each take a `claude.DirRunner`; `Tick` gains the dispatch
  step after the counts write.
- `daemon/loop_test.go` — the new RED test; the Verse 1 tests get the extra argument.
- `cmd/serve.go` — passes `claude.ExecDirRunner` into `daemon.Loop`.
- `task/store.go` — export `ParentID`, wrapping the existing unexported `barParent`
  (`task/store.go:548`), so `daemon` can find a bar's track.

## References
- `probe-dispatch.sh` — the bash spelling of this spawn. The Go here is that `cd` + `claude --bg`
  with `exec.Cmd.Dir` doing the `cd`; production adds `--agent bit:bot-dev` and `-w`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTick_DispatchesTheQueuedBar` in `daemon/loop_test.go`
     - **Behavior:** a queue row for an approved, not-done bar causes exactly one `claude` spawn,
       started **in the registered project's directory** and carrying the bar in its prompt — the
       loop finally consumes the queue instead of only refreshing counts.
     - **Setup:** `t.Setenv("HOME", t.TempDir())`, `t.Setenv("XDG_DATA_HOME", "")`.
       `dir := t.TempDir()`; `store := task.New(filepath.Join(dir, ".bit"))`; save
       `&task.Task{ID: "ACME-1", Title: "a track", Status: task.StatusTodo, Approved: true}` and
       `&task.Task{ID: "ACME-1.1", Title: "a bar", Status: task.StatusTodo, Approved: true}`.
       `db.Open()`, `q := orm.New(sqlDB)`, `q.CreateProject(ctx, orm.CreateProjectParams{Path: dir, Code: "ACME"})`,
       read the id back with `q.GetProjectByPath(ctx, dir)`, then
       `q.EnqueueTask(ctx, orm.EnqueueTaskParams{ProjectID: id, TargetID: "ACME-1.1", TargetTyp: "bar"})`.
       Fake runner: a closure over a `[]struct{dir, name string; args []string}` slice that appends
       each call and returns `("[]", 0, nil)` — an empty `agents --json` array, so nothing else in
       `Tick` has a live session to react to. Logger to `io.Discard`.
     - **Assertions:** among the recorded calls, exactly one has
       `args[0] == "--bg"`, and that call reads
       `dir == dir`, `name == "claude"`, and
       `args == []string{"--bg", "--agent", "bit:bot-dev", "-w", "ACME-1-a-track", "-n", "ACME-1-a-track", "/bit:do ACME-1.1"}`.
     - **Boundary:** queued rows for this project == 1 — the lower non-empty bound; proves one row
       is actually read and turned into a spawn, rather than the queue merely being queried.
   - [ ] Confirm fails: the recorded-calls slice is empty — `Tick` makes no `--bg` call. A compile
     failure on the new `Tick` argument comes first; add the argument, then this is the real RED.

2. **Implement (GREEN):**
   - [ ] In `claude/dispatch.go`: `type DirRunner func(ctx context.Context, dir, name string, args ...string) (out string, code int, err error)`
     and `ExecDirRunner`, which is `ExecRunner`'s body plus `cmd.Dir = dir` and returning
     `(out, code, err)` the way `daemon.ExecRunner` (`daemon/daemon.go:21`) already does.
   - [ ] `func WorktreeName(trackID, title string) string` — `trackID + "-" + slug(title)`, where
     `slug` lowercases and collapses each run of non-alphanumeric runes to a single `-`, trimming
     leading and trailing `-`. `"ACME-1"` + `"a track"` → `"ACME-1-a-track"`.
   - [ ] `func Spawn(ctx context.Context, run DirRunner, dir, name, bar string) error` — calls
     `run(ctx, dir, "claude", "--bg", "--agent", "bit:bot-dev", "-w", name, "-n", name, "/bit:do "+bar)`
     and wraps a non-nil error or non-zero code.
   - [ ] In `task/store.go`, add `func ParentID(id string) (string, bool) { return barParent(id) }`.
   - [ ] In `daemon/loop.go`, change to
     `Tick(ctx context.Context, queries *orm.Queries, log *slog.Logger, run claude.DirRunner)` and
     `Loop(ctx context.Context, queries *orm.Queries, log *slog.Logger, interval time.Duration, run claude.DirRunner) error`;
     `Loop` forwards `run` to `Tick`. Inside the per-project loop, after the counts write: read
     `queries.ListQueueByProject(ctx, p.ID)` (already `ORDER BY id`), take element 0 if any, load
     the bar and — via `task.ParentID` — its track from
     `task.New(filepath.Join(p.Path, ".bit"))`, then
     `claude.Spawn(ctx, run, p.Path, claude.WorktreeName(track.ID, track.Title), bar.ID)`, logging
     and continuing on any error.
   - [ ] In `cmd/serve.go`, pass `claude.ExecDirRunner` as `daemon.Loop`'s fifth argument and add
     the `claude` import.
   - [ ] Add the new argument to the three existing `daemon/loop_test.go` call sites — a runner that
     records nothing and returns `("[]", 0, nil)`.

## Claude verifies
- [ ] `just test` — the Verse 1 tests and `cmd/serve_test.go`'s counts tests still pass with the
  threaded runner
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. Nothing real is spawned; the runner is faked.

## Commit (user)
`feat(daemon): dispatch the queued bar on each tick`