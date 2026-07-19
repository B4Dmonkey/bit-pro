---
id: BIT-4.1
title: Model maps tasks to list items in order
status: done
phase: 1
phase_label: open & navigate
---
## Step 1 (Phase 1 — open & navigate) — Model maps tasks to list items in order
**Status:** ✅ Done — verified 2026-07-17
Constructing the model from `Store.List()` output must preserve the store's order and keep
each task reachable. Forced by an empty-vs-populated contradiction: a hardcoded item slice
can't satisfy both.

**Scope:**
- `tui/model.go` — new: `item` type (wraps `*task.Task`, implements `list.DefaultItem`),
  `model` struct (embeds `list.Model`), `New(tasks []*task.Task) model`.
- `tui/model_test.go` — new.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestNew_PreservesStoreOrder`
     - **Behavior:** the model exposes exactly the tasks it was given, in the same order,
       each recoverable by ID — so the view can never silently reorder or drop the store's
       track→bar ordering.
     - **Setup:** build `[]*task.Task` with IDs `{"BIT-2", "BIT-2.1", "BIT-1"}` (the shape
       `List` returns — a track, its bar, an older track).
     - **Assertions:** reading the model's items back yields IDs `["BIT-2","BIT-2.1","BIT-1"]`
       in that order.
     - **Boundary:** count == 3 (N) — the many-item path where order is observable.
   - [x] Confirm fails: `tui.New` / `model` undefined (compile error).

2. **Implement (GREEN):**
   - [x] `item` wrapping `*task.Task`; `FilterValue()`, `Title()`, `Description()` off the
     task so a default delegate can render Phase 1.
   - [x] `New` maps tasks → `[]list.Item` in order, seeds a `list.Model` with
     `list.NewDefaultDelegate()`. Just enough to make the ordering test pass.

3. **More tests (RED → GREEN):**
   - [x] `TestNew_EmptyList`
     - **Behavior:** an empty project builds a valid, empty model rather than panicking —
       the store legitimately returns zero tasks before anything is created.
     - **Setup:** `New(nil)`.
     - **Assertions:** the model has 0 items; construction does not panic.
     - **Boundary:** count == 0 — the lower bound; contradicts any hardcoded non-empty item
       slice.

**Claude verifies:**
- [x] `just test` green
- [x] `just lint` clean

**Commit (user):** `feat(tui): model maps store tasks to list items`