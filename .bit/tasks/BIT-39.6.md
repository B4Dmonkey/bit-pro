---
id: BIT-39.6
title: An unconfirmed spawn keeps its row
status: todo
phase: 2
phase_label: Bar runs unattended
---
## **Verse 2**

A spawn whose session never shows up keeps its row, so the bar is retried next tick rather than
dropped. Contradicts the previous bar's unconditional delete: same fixture, empty `agents --json`,
opposite expectation.

## Scope
- `daemon/loop.go` — the delete becomes conditional; the miss logs at warn.
- `daemon/loop_test.go` — the RED test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTick_KeepsTheRowWhenTheSessionCannotBeConfirmed` in `daemon/loop_test.go`
     - **Behavior:** the queue row is the record that a bar still needs doing, so it survives a
       spawn the loop cannot see — losing it would silently drop the bar with nothing to say so.
     - **Setup:** the `TestTick_DispatchesTheQueuedBar` fixture, with the fake runner returning
       `("[]", 0, nil)` for the `agents --json` call — the spawn appears to succeed and no session
       is visible. Logger writes to a `bytes.Buffer` at debug level.
     - **Assertions:** `ListQueueByProject` still returns exactly 1 row, and its `TargetID` is
       `"ACME-1.1"`. The buffer contains a record naming the bar `ACME-1.1` — the miss is not
       silent.
     - **Boundary:** confirmed sessions matching the dispatched name == 0 — the lower bound of the
       `agents --json` match, and the exact case the previous bar's single-row payload sat above.
   - [ ] Confirm fails: `ListQueueByProject` returns 0 rows — the unconditional delete from the
     previous bar fires.

2. **Implement (GREEN):**
   - [ ] In `daemon/loop.go`, guard the `DeleteQueueRow` call on the name being found, and
     `log.Warn` with the bar id and the derived name when it is not.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(daemon): keep the queue row when a spawn cannot be confirmed`