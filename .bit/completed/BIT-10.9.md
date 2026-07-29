---
id: BIT-10.9
title: Relocating a bar drops it from its parent's order
status: done
phase: 3
phase_label: Order consistency
---
Relocating a bar drops its id from its parent track's `Order`, closing the gap where `Relocate` renamed the file but left a phantom entry behind. Forced by a test that reorders three bars, relocates the middle one, and reads the parent's `Order` back — an assertion that can't pass while `Relocate` ignores `Order`.

**Scope:**
- `task/store.go` — after `Relocate` moves the file, when the relocated `id` is a bar (`barParent` returns ok), drop `id` from the parent track's `Order` via a small helper (`removeFromOrder`) that mirrors `AppendToOrder`/`insertInOrder`. A whole-track relocate is untouched: `barParent(track)` is false, so the helper never runs, and the track's own `Order` archives with it. Relocate the file first, then update `Order`, so a `Save` failure leaves today's harmless-phantom state rather than a dropped-but-still-present bar.
- `task/store_test.go` — new `TestStoreRelocate_*` cases alongside the existing family.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreRelocate_DropsBarFromParentOrder`
     - **Behavior:** relocating a bar removes exactly that bar's id from its parent track's explicit `Order`, and the surviving bars keep their relative order.
     - **Setup:** `s := New(t.TempDir())`; save `BIT-1` (`Status: "done"`, `Order: []string{"BIT-1.1", "BIT-1.2", "BIT-1.3"}`) and the three bars (`Status: "done"`); `s.Relocate("BIT-1.2", false)`.
     - **Assertions:** `s.Load("BIT-1").Order` equals `[]string{"BIT-1.1", "BIT-1.3"}` (use `slices.Equal`).
     - **Boundary:** relocate the *middle* of three — proves it drops only the target and preserves the order of the two that remain, not just "shrinks the list."
   - [ ] Confirm fails: `Order = [BIT-1.1 BIT-1.2 BIT-1.3], want [BIT-1.1 BIT-1.3]` — `Relocate` never touches `Order` today.

2. **Implement (GREEN):**
   - [ ] In `Relocate`, after `s.relocate(id)` succeeds, if `parent, ok := barParent(id); ok` then `return s.removeFromOrder(parent, id)`, else `return nil`.
   - [ ] Add `removeFromOrder(parent, id string) error`: `Load` the parent; if `len(track.Order) == 0` return nil (never synthesize an `Order` on a never-reordered track — same guard as `AppendToOrder`); else `track.Order = slices.DeleteFunc(track.Order, func(x string) bool { return x == id })` and `Save`.

3. **More tests (RED → GREEN):**
   - [ ] `TestStoreRelocate_LeavesLegacyOrderUnmaterialized`
     - **Behavior:** relocating a bar under a track that was never reordered neither errors nor writes an `Order` — the track keeps falling back to id-sequence ordering.
     - **Setup:** save `BIT-1` (`Status: "done"`, `Order: nil`) and `BIT-1.1`, `BIT-1.2` (`Status: "done"`); `s.Relocate("BIT-1.1", false)`.
     - **Assertions:** no error; `s.Load("BIT-1").Order` is empty (`len == 0`).
     - **Boundary:** `Order` length == 0 — the lower bound; pins that the fix guards the empty case instead of materializing a manifest on delete. (Guards the `len == 0` early return; passes as soon as GREEN lands.)

**Claude verifies:**
- [ ] `just test` passes (existing `TestStoreRelocate_CascadesToBars` and the delete-cmd tests stay green — the track-cascade path is unchanged).
- [ ] `just lint` passes.

**User verifies:**
- [ ] The diff only teaches the single-bar path to maintain `Order` and leaves the whole-track cascade alone — matching the scope's "a track relocate is covered for free" decision.

**Commit (user):** `fix(task): drop relocated bar from its parent track's order`