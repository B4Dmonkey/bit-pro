---
id: BIT-8.1
title: List is the single ordering source
status: done
phase: 1
phase_label: Resequence
---
Make `List()` the single ordering source: a track's bars sort by an explicit `order` list on the track, falling back to today's ID sequence when there is no `order`. This is the read side every surface consumes, so it lands first.

**Representation:** the manifest is `Order []string` — a YAML array of **full dotted bar IDs** (`["BIT-1.2", "BIT-1.1"]`), not bare numbers (`[2, 1]`). The ID is the identity handle, so the list references bars by that handle: `order[i]` maps directly to a loaded `*task.Task.ID` and to a `<id>.md` file, with no ID re-composition anywhere. A stale entry (deleted bar) is harmless — `List()` only positions files it actually loaded.

**Scope:**
- `task/task.go` — add `Order []string` to `Task` with `yaml:"order,omitempty"`, holding child bar IDs (full IDs) in display order. Only tracks populate it. Round-trips through the existing `Parse`/`Bytes` yaml marshal for free.
- `task/store.go` — in `List()`, after loading, build a position map `map[string]map[string]int` (track ID → bar ID → index) from each track's `Order`. Replace the `slices.SortFunc(..., compareIDs)` call with a comparator that: reuses `compareIDs` for track grouping, track-heads-its-bars, and the legacy fallback; and overrides only the bar-vs-bar case where both bars are the same track's children AND both appear in that track's order map — then compare by index. `compareIDs` and `idParts` stay as-is (their test stays green).

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreList_OrdersBarsByExplicitOrder` (table-driven, task pkg, using `Save` + `List`)
     - **Behavior:** when a track carries an explicit `order`, `List()` returns that track's bars in that order, not in ID-number order — proving order is decoupled from the ID.
     - **Setup:** `Save` track `BIT-1` with `Order: []string{"BIT-1.2", "BIT-1.1"}`, plus bars `BIT-1.1` and `BIT-1.2` (status todo). Second table case: same three tasks but track saved with no `Order`.
     - **Assertions:** with `Order` set → returned IDs are `[BIT-1, BIT-1.2, BIT-1.1]`. With no `Order` → `[BIT-1, BIT-1.1, BIT-1.2]` (legacy ID sequence, unchanged).
     - **Boundary:** the two states of the `order` field — populated (N=2, reversed from ID sequence) vs absent (the legacy default). The reversed case cannot pass under the old ID sort.
   - [ ] Confirm fails: `Order` field doesn't exist yet (compile error), then once added, the reversed case returns `[BIT-1.1, BIT-1.2]` — wrong order.

2. **Implement (GREEN):**
   - [ ] Add the `Order` field to `Task`.
   - [ ] Build the position map in `List()` and swap in the order-aware comparator described in Scope.

**Claude verifies:**
- [ ] `just test` passes — new test green, and every existing `task`/`cmd` test still green (legacy tracks with no `order` are unaffected).
- [ ] `just lint` clean.

**User verifies:**
- [ ] The frontmatter `order:` array is a reasonable home for the manifest (vs. body/sidecar) before it becomes load-bearing.

**Commit (user):** `feat(order): List honors a track's explicit bar order`