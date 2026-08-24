---
id: BIT-38.10
title: Store.Move owns the one-anchor rule
status: todo
approved: true
phase: 2
phase_label: Plan writes
---
## **Verse 2**

`cmd/task/move.go:22` owns "specify exactly one of `--before` or `--after`" because `Store.Move(id, anchor string, before bool)` cannot see whether the caller supplied both anchors or neither. The track's Decisions settle this: the rule moves down and both callers stay thin. Forced by a store-level test calling `Move` with the new anchor pair — it will not compile against the current signature.

This bar changes an existing signature, so the CLI caller and both existing store tests move with it in the same commit. Splitting them would leave the module un-buildable between bars.

## Scope
- `task/store.go` — `Move` becomes `Move(id, before, after string) error` and owns the one-anchor check.
- `task/store_test.go` — the two `Move` call sites (around lines 577 and 625) and their table fields.
- `cmd/task/move.go` — `RunE` drops the `Changed`/`hasBefore` dance and passes both flag values through.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreMove_RejectsAnchorPair` (table-driven, new)
     - **Behavior:** the store refuses a call that names both anchors or neither, so the rule holds for every caller rather than only the one that happens to check.
     - **Setup:** `s := New(t.TempDir())`; seed track `BIT-1` and bars `BIT-1.1`, `BIT-1.2` the way `TestStoreMove_Rejects` already does. Cases: (a) `Move("BIT-1.2", "BIT-1.1", "BIT-1.1")` — both; (b) `Move("BIT-1.2", "", "")` — neither.
     - **Assertions:** each returns a non-nil error. `s.Load("BIT-1").Order` is unchanged from the seed in both cases — the refusal happens before any write.
     - **Boundary:** the anchor pair at both invalid points of its four-state space (both set, neither set), with the two valid states covered by the existing `TestStoreMove` cases converted to the new signature.
   - [ ] Confirm fails: compile error — `too many arguments in call to s.Move` / `cannot use "BIT-1.1" (untyped string constant) as bool value`. That is the correct first failure: the signature is the thing under change.
   - [ ] Convert the existing tables: `TestStoreMove`'s `before bool` field becomes the two anchor strings (case "materializes then moves to front" passes `before: tid1_1, after: ""`; "splices an existing order to the back" passes `before: "", after: tid1_3`), and `TestStoreMove_Rejects` passes its `anchor` as `after` with `before` empty. These conversions are mechanical — no assertion changes.

2. **Implement (GREEN):**
   - [ ] Change the signature to `func (s *Store) Move(id, before, after string) error`.
   - [ ] First statement: `if (before == "") == (after == "") { return errors.New("specify exactly one of before or after") }`. Then `anchor, isBefore := after, false`; `if before != "" { anchor, isBefore = before, true }`. The rest of the method — the self-move check, the sibling check, the two `Load` calls, `insertInOrder` — is unchanged.
   - [ ] `task/store.go` needs the `errors` import.
   - [ ] Rewrite `newMoveCmd`'s `RunE` to a single line: `return taskstore.New(bitdir.Current()).Move(args[0], before, after)`. Delete the `hasBefore`/`hasAfter`/`anchor` block and the `errors` import from `cmd/task/move.go`.

## Claude verifies
- [ ] `just test` — the converted store tests pass, and `cmd/task/move_test.go` passes **unchanged**, including `TestTaskMoveCmd_RejectsBadFlags`; its both/neither cases now fail in the store instead of the command, which is the point
- [ ] `just lint`
- [ ] `grep -n "errors\|hasBefore" cmd/task/move.go` returns nothing

## User verifies
- [ ] none — deterministic

## Commit (user)
`refactor(task): move the one-anchor rule into Store.Move`