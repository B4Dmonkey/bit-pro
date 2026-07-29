---
id: BIT-3.1
title: Reversing the sort proves the list is orderable
status: done
phase: 1
phase_label: Newest first
---
## Step 1 (Phase 1 — Newest first) — Reversing the sort proves the list is orderable

**Status:** ✅ Done — verified 2026-07-17

`bit task list` returns tasks in ascending lexical order. This step flips it, with the
smallest change that can pass: reverse the sort. That is deliberately not the real
answer — Step 2 exists to prove it.

**Scope:**
- `task/store.go` — `List`: reverse the existing `slices.Sort(matches)`
- `cmd/task_list_test.go` — `TestTaskListCmd_ShowsAllTasksInIDOrder` currently asserts
  ascending; it is the test being changed, so rename it to say newest-first

**TDD cycle:**

1. **Write test (RED):**
   - [x] Rename `TestTaskListCmd_ShowsAllTasksInIDOrder` →
         `TestTaskListCmd_ShowsNewestFirst`
     - **Behavior:** the most recently created task is the first line of output, because
       recent work is what you're almost always looking for.
     - **Setup:** `initProject(t, "BIT")`, then `createTask(t, "First", "...")` and
       `createTask(t, "Second", "...")` — yielding BIT-1 then BIT-2.
     - **Assertions:** output is exactly `"BIT-2\ttodo\tSecond\nBIT-1\ttodo\tFirst\n"`.
     - **Boundary:** task count == 2 — the smallest count where order is observable at
       all (at 1 or 0, every ordering is identical).
   - [x] Confirm fails: got `BIT-1…BIT-2`, want `BIT-2…BIT-1`

2. **Implement (GREEN):**
   - [x] In `List`, add `slices.Reverse(matches)` after `slices.Sort(matches)`

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**Commit (user):** `feat(list): show newest tasks first`