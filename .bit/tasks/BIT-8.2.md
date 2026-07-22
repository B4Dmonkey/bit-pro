---
id: BIT-8.2
title: Store.Move resequences a bar
status: done
phase: 1
phase_label: Resequence
---
Add `Store.Move` — the write side that rewrites a track's `order` to resequence one bar relative to a sibling. A legacy track (no `order` yet) is materialized from its current ID sequence first, so the un-moved bars keep their positions. Bar 1 gave us the read side; this contradicts it by producing a track whose stored `order` differs from its ID sequence.

**Scope:**
- `task/store.go` — new method `Move(id, anchor string, before bool) error`: load the moved bar and its parent track (`parent := id[:strings.LastIndex(id, ".")]`); if the track has no `Order`, materialize it from the current `List()`-order of that parent's bars; remove `id`, find `anchor`'s index, re-insert `id` before/after it; `Save` the track. Errors: bar or anchor not found, anchor not a sibling of the moved bar (different parent), moving a bar relative to itself.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreMove_Resequences` (table-driven, task pkg)
     - **Behavior:** `Move` rewrites the parent's `order` so a later `List()` returns the new sequence — including on a track that had no `order` at all (materialize-then-splice).
     - **Setup:** track `BIT-1` + bars `BIT-1.1`, `BIT-1.2`, `BIT-1.3`. Cases: (a) no prior `order`, `Move("BIT-1.3", "BIT-1.1", before=true)`; (b) prior `order=[.1,.2,.3]`, `Move("BIT-1.1", "BIT-1.3", before=false)`.
     - **Assertions:** (a) reload `BIT-1`, `Order == [BIT-1.3, BIT-1.1, BIT-1.2]`; (b) `Order == [BIT-1.2, BIT-1.3, BIT-1.1]`.
     - **Boundary:** the materialize path (order absent — count 0 → full N) vs the splice path (order already present); move to front (`before` first element) and move to back (`after` last element).
   - [ ] `TestStoreMove_Rejects` (table-driven)
     - **Behavior:** illegal moves error instead of corrupting the manifest.
     - **Setup/Assertions:** anchor under a different parent (`BIT-1.1` anchored to `BIT-2.1`) → error; unknown bar / unknown anchor → error wrapping `fs.ErrNotExist`; `id == anchor` → error.
     - **Boundary:** the sibling constraint (same parent required) and the self-move degenerate case.
   - [ ] Confirm fails: `Move` doesn't exist (compile error).

2. **Implement (GREEN):**
   - [ ] Implement `Move` per Scope. Reuse `List()` (Bar 1) to derive the materialization sequence so there is one definition of "current order".

**Claude verifies:**
- [ ] `just test` passes.
- [ ] `just lint` clean.

**User verifies:**
- [ ] The sibling-only constraint (can't move a bar under a different track) matches how you'd resequence during bit_plan.

**Commit (user):** `feat(order): Store.Move resequences a bar within its track`