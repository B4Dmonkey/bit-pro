---
id: BIT-4.4
title: Model resolves the selected task
status: done
phase: 2
phase_label: read a task
---
## Step 4 (Phase 2 — read a task) — Model resolves the selected task
**Status:** ✅ Done — verified 2026-07-18
The detail view needs the task under the cursor, whose `Body` is already loaded. Forced by
a first-vs-moved contradiction: a hardcoded `tasks[0]` fails once the selection moves.

**Scope:**
- `tui/model.go` — `selected() *task.Task` off `list.SelectedItem()`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestSelected_TracksCursor`
     - **Behavior:** `selected()` returns the task at the current cursor, so the detail view
       always opens what the user is looking at.
     - **Setup:** `New` with 3 tasks; assert default selection, then `list.Select(2)`.
     - **Assertions:** before moving, `selected().ID == tasks[0].ID`; after `Select(2)`,
       `selected().ID == tasks[2].ID`.
     - **Boundary:** index 0 (default) and index N-1 (last) — both ends of the cursor range.
   - [x] Confirm fails: `selected` undefined.

2. **Implement (GREEN):**
   - [x] `selected()` type-asserts `list.SelectedItem()` to `item`, returns its `*task.Task`.
     (Comma-ok assertion: an empty list's nil `SelectedItem()` yields `ok == false` → nil,
     no panic.)

3. **More tests (RED → GREEN):**
   - [x] `TestSelected_EmptyListNil`
     - **Behavior:** with no tasks there is nothing to open; `selected()` returns nil rather
       than panicking on a nil `SelectedItem`.
     - **Setup:** `New(nil)`.
     - **Assertions:** `selected() == nil`.
     - **Boundary:** count == 0 — forces the nil-guard the hardcoded assertion skips.

**Claude verifies:**
- [x] `just test` green
- [x] `just lint` clean

**Commit (user):** `feat(tui): resolve the selected task`