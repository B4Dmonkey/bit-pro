---
id: BIT-10.4
title: Archived IDs stay reserved
status: done
phase: 1
phase_label: Archive
---
Count archived files when minting IDs, so a relocated ID is never re-issued onto a different task. Contradicts the tasks-only glob in `NextID`/`NextChildID`.

**Scope:**
- `task/store.go` — `NextID` and `NextChildID` also scan `archiveDir()`; the highest is the max across both dirs. Consider extracting `highestSuffix(dir string, re *regexp.Regexp) int` to avoid duplicating the glob+scan.
- `task/store_test.go` — new tests.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreNextID_ReservesArchivedIDs`
     - **Behavior:** an archived task's number still counts toward the next id.
     - **Setup:** `s.Save(BIT-1)`; `s.Save(BIT-2, Status:"done")` then `s.Relocate("BIT-2", false)`; `s.NextID("BIT")`.
     - **Assertions:** returns `"BIT-3"`, not `"BIT-2"`.
     - **Boundary:** the highest id lives only in `archive/` — the reserved-after-relocate case.
   - [ ] `TestStoreNextChildID_ReservesArchivedChildren`
     - **Behavior:** an archived bar's number still counts toward the next child id.
     - **Setup:** track `BIT-1` present in `tasks/`; child `BIT-1.1` placed in `archive/` (save then relocate); `s.NextChildID("BIT-1")`.
     - **Assertions:** returns `"BIT-1.2"`.
     - **Boundary:** the highest child only in `archive/`.
   - [ ] Confirm fails: both re-mint the archived number (`BIT-2` / `BIT-1.1`).

2. **Implement (GREEN):**
   - [ ] Glob both `tasksDir()` and `archiveDir()` in each; take the max highest across the two.

`NextChildID` still stats the parent in `tasksDir()` — a live parent is the normal case; leave that check as-is.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**Commit (user):** `feat(task): reserve archived IDs when minting`