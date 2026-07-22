---
id: BIT-8.5
title: create appends to a reordered track
status: doing
phase: 2
phase_label: Insert mid-plan
---
Once a track has an explicit `order`, plain `bit task create` must append the new bar's ID to it — otherwise the manifest silently drifts out of sync with the files. This forces `create` to own the invariant "create appends," which Phase 2's `--after` then builds on. Contradicts a reordered track: after Bar 3 rewrote `order`, a new bar must join the manifest, not just appear via fallback.

**Scope:**
- `task/store.go` — after minting + saving a child, if the parent track has a non-empty `Order`, append the new ID and `Save` the track. Expose this as the path `create` uses (e.g. a small helper `appendToOrder(parent, id)` or fold into a `CreateChild` seam). A parent with no `Order` (fully legacy) is left untouched — the fallback still places the new bar last.
- `cmd/task_create.go` — the `--parent` branch calls that path so the manifest stays complete.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestTaskCreate_AppendsToReorderedTrack` (cmd pkg)
     - **Behavior:** creating a bar under a track that has already been reordered keeps the manifest complete and correct — the new bar lands last and `order` contains every bar.
     - **Setup:** init `BIT`; track `BIT-1`; bars `BIT-1.1`, `BIT-1.2`; `task move BIT-1.2 --before BIT-1.1` (order now `[BIT-1.2, BIT-1.1]`); then `task create "third" --parent BIT-1`.
     - **Assertions:** reload `BIT-1`, `Order == [BIT-1.2, BIT-1.1, BIT-1.3]`; `task list --parent BIT-1` ends with `BIT-1.3`.
     - **Boundary:** parent with a present, reordered `order` (the append-into-existing case) — distinct from the legacy no-`order` parent, which stays untouched.
   - [ ] Confirm fails: `order` reloads as `[BIT-1.2, BIT-1.1]` — `BIT-1.3` missing from the manifest.

2. **Implement (GREEN):**
   - [ ] Append the new child ID to a present parent `Order` on create; leave a legacy (no-`order`) parent alone.

**Claude verifies:**
- [ ] `just test` passes — existing `task_create` tests (legacy parents, no `order`) still green.
- [ ] `just lint` clean.

**User verifies:**
- [ ] Creating under a not-yet-reordered track still behaves exactly as before (no `order` field written until a move happens).

**Commit (user):** `feat(order): task create appends to a reordered track's order`