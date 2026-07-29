---
id: BIT-3.4
title: A second child contradicts the hardcode
status: done
phase: 2
phase_label: A step belongs to a scope
---
## Step 4 (Phase 2 — A step belongs to a scope) — A second child contradicts the hardcode

**Status:** ✅ Done — verified 2026-07-17

`.1` can't also be `.2`. This forces `NextChildID` to actually scan.

**Scope:**
- `task/store.go` — `NextChildID`: real implementation
- `task/store_test.go` — extend `TestStoreNextID` with a dotted-children row

**TDD cycle:**

1. **Write test (RED):**
   - [x] Extend `TestTaskCreateCmd_ParentMintsDottedID` into a table, or add
         `TestTaskCreateCmd_SecondChildIncrements`
     - **Behavior:** children number sequentially, so dotted IDs can match a plan's step
       numbers 1..N.
     - **Setup:** BIT-1 exists; create three children in a row.
     - **Assertions:** `task list` contains `BIT-1.1`, `BIT-1.2`, `BIT-1.3`.
     - **Boundary:** child count 1→N — the second child is where "always .1" first
       contradicts.
   - [x] Confirm fails: second create overwrites `BIT-1.1.md`; no `BIT-1.2` to read

2. **Implement (GREEN):**
   - [x] `NextChildID`: glob `filepath.Join(s.tasksDir(), parent+".*.md")`, match
         `^` + `regexp.QuoteMeta(parent)` + `\.(\d+)\.md$`, take `highest+1` — the same
         shape as `NextID`, whose `QuoteMeta` on a parent containing a dot is what keeps
         the pattern honest

3. **More tests (RED → GREEN):**
   - [x] Add a row to `TestStoreNextID`: `{name: "ignores dotted children", existing:
         []string{"BIT-1", "BIT-1.1", "BIT-1.13"}, want: "BIT-2"}`
     - **Behavior:** minting a new *track* skips over bars entirely — a plan's 13 steps
       must not push the next scope to BIT-14.
     - **Setup:** as above.
     - **Assertions:** `NextID("BIT")` == `"BIT-2"`.
     - **Boundary:** a dotted ID present in the same directory `NextID` globs — verified
       to already pass today (`^BIT-(\d+)\.md$` excludes it), so this is a regression
       guard pinning behavior that currently works by accident.

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**Commit (user):** `feat(create): number child tasks sequentially under their parent`