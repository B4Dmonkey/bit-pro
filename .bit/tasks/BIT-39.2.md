---
id: BIT-39.2
title: Loop owns the ticker; serve keeps only wiring
status: todo
phase: 1
phase_label: Loop lives in daemon
---
## **Verse 1**

The ticker, the started/stopped logging, and the cancellation check move into `daemon.Loop`,
leaving `cmd/serve.go` with flag parsing and wiring only. Forced by a test that drives the loop
directly from the `daemon` package — it cannot compile while the loop is a closure inside
`newServeDaemonCmd`'s `RunE`.

## Scope
- `daemon/loop.go` — add `Loop(ctx, queries, log, interval) error` above `Tick`.
- `daemon/loop_test.go` — add the two RED tests.
- `cmd/serve.go` — `RunE` reduces to level → handler → `db.Open` → `orm.New` →
  `return daemon.Loop(cmd.Context(), queries, log, serveTick)`. `serveTick` stays a `cmd` package
  var so the existing tests keep shortening it.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestLoop_LogsStartedAndStoppedAroundItsTicks` in `daemon/loop_test.go`
     - **Behavior:** `Loop` owns the whole daemon body — it brackets its work with the `started` and
       `stopped` info lines and ticks at the interval it is handed, so the command no longer has to.
     - **Setup:** `t.Setenv("HOME", t.TempDir())`, `t.Setenv("XDG_DATA_HOME", "")`; `db.Open()` with
       no projects registered; `var buf bytes.Buffer` behind
       `slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))`;
       `ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)`; interval `5*time.Millisecond`
       — the same 5ms/50ms pairing `cmd/serve_test.go` already relies on.
     - **Assertions:** `Loop` returns nil; decoding `buf` line by line as JSON, the `msg` sequence
       starts with `started`, ends with `stopped`, and carries at least 2 `tick` records in between.
     - **Boundary:** elapsed time == 10× the interval — the low end of the many-ticks range; two ticks
       is the smallest count that proves the ticker repeats rather than firing once.
   - [ ] `TestLoop_ReturnsWithoutTickingWhenTheContextIsAlreadyCancelled`
     - **Behavior:** a cancelled context wins over the ticker, so a stopped daemon does no work on
       its way out.
     - **Setup:** same buffer/logger; `ctx, cancel := context.WithCancel(t.Context())` then `cancel()`
       immediately; interval `5*time.Millisecond`.
     - **Assertions:** `Loop` returns nil and `buf` contains no `tick` record — only `started` and
       `stopped`.
     - **Boundary:** ticks elapsed == 0 — the lower bound; the only case where `ctx.Done()` and
       `ticker.C` are not both plausibly ready, so it isolates the select's cancellation arm.
   - [ ] Confirm fails: `undefined: daemon.Loop` — a compile failure, not a failed assertion.

2. **Implement (GREEN):**
   - [ ] Add to `daemon/loop.go`:
     `func Loop(ctx context.Context, queries *orm.Queries, log *slog.Logger, interval time.Duration) error`
     — `log.Info("started")`, `defer log.Info("stopped")`, `ticker := time.NewTicker(interval)`,
     `defer ticker.Stop()`, then the `for { select { case <-ctx.Done(): return nil; case <-ticker.C:
     log.Debug("tick"); Tick(ctx, queries, log) } }` body moved from `RunE`. Add the `time` import.
   - [ ] In `cmd/serve.go`, replace everything in `RunE` after `queries := orm.New(sqlDB)` with
     `return daemon.Loop(cmd.Context(), queries, log, serveTick)`. Keep the `time` import —
     `serveTick = 10 * time.Second` still needs it. `newHandler` stays in `cmd`; it shapes the
     logger, it is not part of the loop.

## Claude verifies
- [ ] `just test` — in particular `TestServeCmd_LogsStartAndStop`, `TestServeCmd_TicksOnlyWhenVerbose`,
  `TestServeCmd_ReturnsWhenContextCancelled` and `TestNewHandler_PicksEncodingFromTheWriter` all still
  pass unchanged: the command's observable behaviour is what this move must not alter
- [ ] `just lint`

## User verifies
- [ ] Whole slice: `just install`, then run `bp serve daemon -v` in one terminal and leave it ~30s.
  It logs `started` once and a `tick` line about every 10s. In a second terminal, `bp list` now shows
  non-zero `backlog:`/`todo:` counts for `BIT` that match what `bp task list` reports — the refresh
  still runs, and nothing about the operator's view changed when the loop moved packages.

## Commit (user)
`refactor(daemon): move the tick loop out of the serve command`