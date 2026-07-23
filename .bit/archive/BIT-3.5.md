---
id: BIT-3.5
title: A dotted ID contradicts the flat sort
status: done
phase: 2
phase_label: A step belongs to a scope
---
## Step 5 (Phase 2 — A step belongs to a scope) — A dotted ID contradicts the flat sort

**Status:** ✅ Done — verified 2026-07-17

Step 2's comparator reads the number after the last `-`, so `BIT-1.2` parses as garbage
and sorts last. A list with parents and children mixed forces the hierarchy into the
comparison. This is the contradiction the scope predicted.

**Scope:**
- `task/store.go` — `compareIDs`: split the suffix on `.` into (track, bar)
- `task/store_test.go` — extend `TestCompareIDs`

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskListCmd_GroupsBarsUnderTheirTrack`
     - **Behavior:** the list reads as a tree — newest track first, its steps beneath it
       in plan order — without an index or a type field.
     - **Setup:** `initProject(t, "BIT")`; create track "One" (BIT-1) and track "Two"
       (BIT-2); under BIT-1 create two children; under BIT-2 create one.
     - **Assertions:** the ID column, in order, is exactly `BIT-2, BIT-2.1, BIT-1,
       BIT-1.1, BIT-1.2`.
     - **Boundary:** two tracks with differing child counts (1 and 2) — proves the sort
       groups by track rather than interleaving, which a single-track fixture cannot show.
   - [x] Confirm fails: dotted IDs fail `Atoi` and sort last — got `BIT-2, BIT-1,
         BIT-2.1, BIT-1.1, BIT-1.2` or similar

2. **Implement (GREEN):**
   - [x] `compareIDs`: take the substring after the last `-`, split on the first `.` into
         track and bar (`bar = 0` when absent), `Atoi` both. Compare **track descending**;
         on a tie, compare **bar ascending**. A track's bar of 0 makes it head its group
         with no special case.

3. **More tests (RED → GREEN):**
   - [x] Extend `TestCompareIDs` with: `("BIT-2","BIT-2.1")` → `<0` (track heads its
         bars); `("BIT-2.1","BIT-2.13")` → `<0` (bars ascend, and 13 is not "less than" 2);
         `("BIT-2.1","BIT-1.9")` → `<0` (track dominates bar).
     - **Behavior:** each of the three ordering rules is pinned independently, so a later
       change that breaks one fails loudly.
     - **Boundary:** bar 1 vs 13 — the digit-count boundary again, now on the right of
       the dot, where the same lexical bug would reappear.

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**User verifies:**
- [x] The grouped list is what you actually want to look at — this is the ordering
      decision made during planning, seen for the first time against real output

**Commit (user):** `feat(list): group child tasks under their parent`