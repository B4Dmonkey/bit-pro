---
id: BIT-39.7
title: The ledger drops a row it must not run
status: done
approved: true
phase: 2
phase_label: Bar runs unattended
---
## **Verse 2**

The ledger, not the queue, decides whether a bar should run: a popped row whose bar is already
`done`, or is no longer approved, is deleted without a spawn. Contradicts the dispatch path — same
queue row, but no `--bg` call may be recorded.

## Scope
- `daemon/loop.go` — the ledger check between popping the row and spawning.
- `daemon/loop_test.go` — the RED test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTick_DropsARowItMustNotDispatch` in `daemon/loop_test.go`, table-driven
     - **Behavior:** a row whose bar is finished or unblessed is removed rather than run or left in
       place — both states are reachable (a replan revokes approval), and leaving the row would
       block that project's queue head forever.
     - **Setup:** the `TestTick_DispatchesTheQueuedBar` fixture, varying only the saved
       `ACME-1.1`: `{name: "already done", bar: {Status: task.StatusDone, Approved: true}}` and
       `{name: "approval revoked", bar: {Status: task.StatusTodo, Approved: false}}`. Runner
       returns `("[]", 0, nil)`. Logger to a `bytes.Buffer`.
     - **Assertions:** for both cases — no recorded call has `args[0] == "--bg"`;
       `ListQueueByProject` returns 0 rows; the buffer names `ACME-1.1`.
     - **Boundary:** these are the two values on the far side of the dispatch predicate's two
       inputs — `Status` at `done` (its terminal value, the dispatch path having exercised `todo`)
       and `Approved` at `false` (its other boolean state). Together they cover both conditions the
       gate reads.
   - [ ] Confirm fails: a `--bg` call is recorded in both cases — the loop dispatches a done bar.

2. **Implement (GREEN):**
   - [ ] In `daemon/loop.go`, after loading the bar and before deriving the name: if
     `bar.Status == task.StatusDone || !bar.Approved`, `queries.DeleteQueueRow(ctx, row.ID)`,
     `log.Info` with the bar id and the reason, and move to the next project.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(daemon): drop a queue row the ledger says must not run`