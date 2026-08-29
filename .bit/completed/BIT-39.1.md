---
id: BIT-39.1
title: Tick moves the counts write into daemon
status: done
approved: true
phase: 1
phase_label: Loop lives in daemon
---
## **Verse 1**

`writeCounts` moves out of `cmd/serve.go` into `daemon/` as `Tick` — one tick's worth of work,
named so the later verses add dispatch *inside* it rather than beside it. Forced by a test in the
`daemon` package that calls `daemon.Tick` directly, which cannot compile while the function lives
in `cmd`.

## Scope
- `daemon/loop.go` — new file: `Tick(ctx, queries, log)`, the body moved verbatim from `writeCounts`.
- `daemon/loop_test.go` — new file: the RED test.
- `cmd/serve.go` — `writeCounts` deleted (`cmd/serve.go:25`); the `ticker.C` case calls `daemon.Tick`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTick_WritesProjectCounts` in `daemon/loop_test.go`
     - **Behavior:** one call to `Tick` reads a registered project's `.bit/` store and writes its
       four counts onto that project's `projects` row — the counts refresh is reachable from
       `daemon`, not only from inside the command's loop body.
     - **Setup:** `t.Setenv("HOME", t.TempDir())` and `t.Setenv("XDG_DATA_HOME", "")` so `db.Open()`
       migrates a throwaway registry — the pattern `cmd/serve_test.go:24-25` uses. `dir := t.TempDir()`;
       `task.New(filepath.Join(dir, ".bit")).Save(&task.Task{ID: "ACME-1", Title: "a track", Status: task.StatusTodo})`
       — left unapproved, so it counts as backlog. Then `db.Open()` and
       `orm.New(sqlDB).CreateProject(ctx, orm.CreateProjectParams{Path: dir, Code: "ACME"})`.
       Logger: `slog.New(slog.NewJSONHandler(io.Discard, nil))`.
     - **Assertions:** after `daemon.Tick(t.Context(), orm.New(sqlDB), log)`,
       `SELECT backlog, todo, done, completed FROM projects WHERE code = 'ACME'` scans exactly
       `(1, 0, 0, 0)` — the same expectation `TestServeCmd_WritesProjectCounts` already asserts for
       this fixture.
     - **Boundary:** registered-project count == 1 — the lower non-empty bound of the `ListProjects`
       loop; proves one iteration actually writes, rather than that the loop was merely entered.
   - [ ] Confirm fails: `undefined: daemon.Tick` — a compile failure in `daemon/loop_test.go`, not a
     failed assertion. If it fails any other way the test is wired wrong.

2. **Implement (GREEN):**
   - [ ] Create `daemon/loop.go` with
     `func Tick(ctx context.Context, queries *orm.Queries, log *slog.Logger)` — the body of
     `cmd.writeCounts` verbatim. Imports: `context`, `log/slog`, `path/filepath`,
     `github.com/B4Dmonkey/bit-pro/db/orm`, `github.com/B4Dmonkey/bit-pro/task`. No import cycle:
     `db/orm` and `task` import nothing from `daemon`.
   - [ ] Delete `writeCounts` from `cmd/serve.go`, along with its now-unused `context`, `path/filepath` and
     `task` imports (nothing else in the file uses them).
   - [ ] In `newServeDaemonCmd`'s `case <-ticker.C:` arm, replace `writeCounts(ctx, queries, log)`
     with `daemon.Tick(ctx, queries, log)` and add the `daemon` import.

## Claude verifies
- [ ] `just test` — the whole suite, so the four counts/skip tests in `cmd/serve_test.go` still pass
  against the moved function
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. The counts behaviour is unchanged and `cmd/serve_test.go` already covers
  it end to end.

## Commit (user)
`refactor(daemon): move the counts write into Tick`