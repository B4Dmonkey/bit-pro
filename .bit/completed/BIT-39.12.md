---
id: BIT-39.12
title: A spawn that produced no session is not a dispatch
status: done
approved: true
phase: 5
phase_label: Cleanup
---
## **Verse 5**

`claude.Spawn` captures the combined output of the spawn and throws it away on success, and success
means exit 0. Measured 2026-08-27 on 2.1.250: a spawn whose `--agent` cannot be resolved prints a
warning, produces no working session, and exits 0 — so the loop logged `INFO dispatched` for a bar
that never ran, and nothing in the log distinguished it from a good dispatch.

That captured output is the only place a failure diagnostic can reach the log. Neither `--all` nor
parsing the `backgrounded ·` line is available, so the loop cannot classify the spawn — it can only
report what `claude` said and let the operator read it.

The record is renamed with the same change: `dispatching`, not `dispatched`, because the next bar
moves the confirmation to a later tick and one record cannot honestly mean both.

## Scope
- `claude/dispatch.go` — `Spawn` returns `(out string, err error)` instead of discarding `out`.
- `daemon/loop.go` — the `dispatched` record at `:136` becomes `dispatching` and carries `out`.
- `daemon/loop_test.go` — the new assertion.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTick_LogsWhatTheSpawnPrinted`
     - **Behavior:** whatever `claude` printed while spawning reaches the log, so a spawn that
       exited 0 and produced nothing is legible instead of silent.
     - **Setup:** `queries, _ := queuedBar(t)`. A runner that, for the call whose `args[0]` is
       `"--bg"`, returns
       `"warning: no agent named 'bit:bot-dev' — spawning with default template", 0, nil`, and for
       any other call returns `"[]", 0, nil`. JSON handler into a `bytes.Buffer` at
       `slog.LevelDebug`.
     - **Assertions:** a record with `msg == "dispatching"`, `bar == "ACME-1.1"`,
       `worktree == "ACME-1-a-track"`, and an `out` attribute containing
       `no agent named 'bit:bot-dev'`. No record has `msg == "dispatched"`.
     - **Boundary:** exit code 0 with non-empty output — the exact shape that reads as success
       today and is the failure this bar exists to expose. The `out` attribute at its interesting
       value rather than its empty one.
   - [ ] Confirm fails: `Spawn` returns only `error`, so there is no `out` to log; the record is
     `dispatched` and carries no output.

2. **Implement (GREEN):**
   - [ ] `claude/dispatch.go`: `Spawn` returns `(string, error)` — the captured `out` on the happy
     path, and `("", err)` on the exec-error and non-zero-exit paths, which already embed the
     output in their error strings.
   - [ ] `daemon/loop.go`: capture the returned output and log
     `log.Info("dispatching", "project", p.Code, "bar", bar.ID, "worktree", name, "out", out)`.

3. **More tests (RED → GREEN):**
   - [ ] `TestTick_DispatchesTheQueuedBar` (existing) keeps passing — it asserts the argv, which
     this bar does not touch. Run it rather than assuming.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. The record's new content is asserted above; whether it is *enough*
  is observed on the last bar of this verse, once the timing change lands with it.

## Commit (user)
`feat(daemon): log what a spawn printed`