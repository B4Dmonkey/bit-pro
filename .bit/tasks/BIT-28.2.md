---
id: BIT-28.2
title: Ticks appear only under -v
status: todo
approved: true
phase: 1
phase_label: serve
---
## **Verse 1**

`bp serve -v` logs a tick every interval; plain `bp serve` stays silent. Contradicts BIT-28.1's
body, which only waits on the context and can never produce output no matter how long it runs.

## Scope
- `cmd/serve.go` — a `--verbose`/`-v` flag, a `time.Ticker` loop, and an `slog` logger written to
  `cmd.OutOrStdout()`
- `cmd/serve_test.go` — the new tests

The tick interval needs a test seam: a 30s ticker cannot be observed in a unit test. Use an
unexported package-level `var serveTick = 30 * time.Second` in `cmd/serve.go`, which the test
lowers and restores via `t.Cleanup`. No new user-facing flag — the interval is not operator
surface, and the package's tests already run sequentially (they use `t.Chdir`, which forbids
`t.Parallel`), so a package var is safe here.

`-v` is free on this subcommand: cobra only attaches the `--version` shorthand to the command
whose `Version` field is set (root), and it does so on `c.Flags()`, which is local and does not
propagate to children. Verified against `cobra@v1.10.2` `command.go:1238` before planning this.

## TDD cycle

1. **Write test (RED):** `cmd/serve_test.go`
   - [ ] `TestServeCmd_TicksOnlyWhenVerbose` (table-driven)
     - **Behavior:** the tick is the operator's liveness signal while watching the loop, and noise
       in a log kept for real events — so it is `debug`, and plain `bp serve` prints nothing until
       something actually happens.
     - **Setup:** in both subtests, save `serveTick`, set it to `5 * time.Millisecond`, and restore
       it with `t.Cleanup`; build `ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)`,
       `defer cancel()`; run `runWithContext(t, ctx, serveCmdUse)` — with `"-v"` appended in the
       verbose subtest and not in the default one. Wait for the command to return before reading
       the buffer, so there is no race on it.
     - **Assertions:** verbose subtest — `strings.Count(out, "tick") >= 2` and
       `strings.Contains(out, "DEBUG")`. Default subtest — `out == ""`. Assert on substrings, not
       on a whole formatted line: BIT-28.3 changes the handler's encoding, and these assertions
       must survive that.
     - **Boundary:** the `--verbose` boolean in both of its states — the only two levels this
       command can run at.
   - [ ] Confirm fails: verbose subtest fails with `strings.Count(out, "tick") == 0` (BIT-28.1's
         body writes nothing at all); the default subtest passes for the wrong reason, which is
         expected and is what the verbose row is there to catch.

2. **Implement (GREEN):**
   - [ ] `cmd/serve.go`: `var serveTick = 30 * time.Second`.
   - [ ] `cmd/serve.go`: bind `verbose` with `cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, …)`.
   - [ ] `cmd/serve.go`: in `RunE`, pick `slog.LevelDebug` when `verbose` and `slog.LevelInfo`
         otherwise, and build `slog.New(slog.NewTextHandler(cmd.OutOrStdout(), &slog.HandlerOptions{Level: level}))`.
         A hardcoded text handler is correct for this bar — BIT-28.3 contradicts it.
   - [ ] `cmd/serve.go`: replace the bare `<-cmd.Context().Done()` with a
         `ticker := time.NewTicker(serveTick)` / `defer ticker.Stop()` and a `for { select { … } }`
         over `<-ctx.Done()` (return `nil`) and `<-ticker.C` (`log.Debug("tick")`).

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. The live-terminal check belongs on BIT-28.3, which decides the format
      the operator actually reads.

## Commit (user)
`feat(serve): tick the stub loop, logged at debug behind -v`