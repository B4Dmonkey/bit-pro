---
id: BIT-10.2
title: Archiving a track takes its bars
status: todo
phase: 1
phase_label: Archive
---
Make archiving a track carry its bars. Contradicts bar 1.1's single-file move, which would leave the bars behind in `tasks/`.

**Scope:**
- `task/store.go` — add `children(parent string) ([]*Task, error)` (via `List()` + `barParent`); `Relocate` relocates each child, then the track.
- `task/store_test.go` — new test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreRelocate_CascadesToBars`
     - **Behavior:** relocating a track moves the track and every one of its bars in one call.
     - **Setup:** save track `BIT-1` (done) and bars `BIT-1.1`, `BIT-1.2` (both done); `s.Relocate("BIT-1", false)`.
     - **Assertions:** all three at `archive/…`; none in `tasks/`; `List()` is empty.
     - **Boundary:** children count > 1 — the multi-bar cascade (1.1 covered zero children).
   - [ ] Confirm fails: only `BIT-1` moves; `BIT-1.1`/`BIT-1.2` still in `tasks/`.

2. **Implement (GREEN):**
   - [ ] `children(parent)` returns the parent's bars; `Relocate` loops `relocate` over children first, then the track.

Non-transactional: a mid-cascade failure can leave a partial move. Out of scope — the all-done guard (bar 1.3) makes a mismatch rare and this is a local single-user store; don't build rollback.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**Commit (user):** `feat(task): cascade relocate to a track's bars`