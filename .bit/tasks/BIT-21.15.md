---
id: BIT-21.15
title: Hand-edited order entries still rank bars
status: todo
phase: 2
phase_label: Recurrence
---
## **Verse 2**

The `order:` list is the fourth ID carrier and the one the previous bar did not cover: `Parse`
now normalizes `ID`, but a track's `Order` entries are a separate field, matched against bar IDs
by exact string equality in `orderPositions`/`compareByOrder`. A lowercase entry there does not
corrupt anything — it silently stops applying, and the track quietly reverts to ID order.

## Scope
- `task/task.go` — normalize each `Order` entry in `Parse`.
- `cmd/task_list_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskListCmd_HandEditedLowercaseOrderStillRanksBars`
     - **Behavior:** an explicit order applies on identity, not on the spelling stored in the
       list.
     - **Setup:** `initProject(t, "BIT")`; create `BIT-1` with bars `BIT-1.1` and `BIT-1.2`, then
       rewrite `.bit/tasks/BIT-1.md` by hand so its frontmatter carries

       ```
       order:
           - bit-1.2
           - bit-1.1
       ```

       — the reverse of ID order, in the wrong case. Run `task list --parent BIT-1`.
     - **Assertions:** `BIT-1.2` appears before `BIT-1.1` in the output.
     - **Boundary:** an order that *contradicts* ID order. If it agreed, the test would pass for
       the wrong reason — the fallback comparator would produce the same result and prove nothing.
   - [ ] Confirm fails: output is `BIT-1.1` then `BIT-1.2` — the order list is silently ignored
     because no entry matches any bar ID

2. **Implement (GREEN):**
   - [ ] In `Parse`, map `normalizeID` over `t.Order` alongside the `t.ID` normalization.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(task): normalize order list entries read from disk`