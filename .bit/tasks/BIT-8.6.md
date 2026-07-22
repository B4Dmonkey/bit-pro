---
id: BIT-8.6
title: create --after inserts mid-plan
status: todo
phase: 2
phase_label: Insert mid-plan
---
Add `bit task create --after <anchor>` so a new step drops into the middle of a plan instead of only at the end — the "add a missing step" half of the flaw-fixing workflow. The minted ID is still the next free dotted number (identity unchanged); only its position in `order` is chosen. Contradicts Bar 5's append-only behavior: the new bar must land after the anchor, not last.

**Scope:**
- `cmd/task_create.go` — add `--after` (string, anchor bar ID). When set, `--parent` is implied by the anchor's parent (or require both and validate they agree). After minting + saving, insert the new ID into the parent `Order` immediately after the anchor (materializing `order` from the current sequence first if the parent has none, so the insert is well-defined).
- `task/store.go` — reuse the splice/materialize helper from `Move`/Bar 5 so insertion has one definition.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestTaskCreate_AfterInsertsMidPlan` (cmd pkg, table-driven)
     - **Behavior:** `--after` positions a brand-new bar mid-sequence while its ID stays the next free number — proving identity and position are fully decoupled.
     - **Setup:** init `BIT`; track `BIT-1`; bars `BIT-1.1`, `BIT-1.2`; run `task create "inserted" --parent BIT-1 --after BIT-1.1`.
     - **Assertions:** printed new ID is `BIT-1.3`; reload `BIT-1`, `Order == [BIT-1.1, BIT-1.3, BIT-1.2]`; `task list --parent BIT-1` shows `BIT-1.1, BIT-1.3, BIT-1.2`.
     - **Boundary:** insert after the first of two bars (mid-list, not head, not tail) — the position an append cannot reach.
   - [ ] `TestTaskCreate_AfterRejectsUnknownAnchor`
     - **Behavior:** an anchor that isn't an existing sibling bar errors instead of creating a dangling entry.
     - **Setup/Assertions:** `--after BIT-1.9` (nonexistent) → error; no new task file created.
     - **Boundary:** anchor outside the valid set (unknown ID).
   - [ ] Confirm fails: `unknown flag: --after`.

2. **Implement (GREEN):**
   - [ ] Add the `--after` flag and the insert-after-anchor path, reusing the Bar 5 / `Move` splice helper.

**Claude verifies:**
- [ ] `just test` passes.
- [ ] `just lint` clean.

**User verifies:**
- [ ] Reviewing a plan and inserting a missing step with `--after` feels like the natural bit_plan workflow; the minted ID vs. position split reads clearly in `task list`.

**Commit (user):** `feat(cli): task create --after inserts a bar mid-plan`