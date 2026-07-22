---
id: BIT-8
title: 'Reorderable plans: decouple bar order from ID'
status: doing
---
## Why

A bar's dotted ID doubles as its position in the plan: bit_plan creates bars in order and
bit_do runs them in that order, so `BIT-7.2` is both "this task" and "second step." That
coupling makes plans unable to correct themselves. When you review a plan and find a flaw —
a step in the wrong place, or a missing step that belongs in the middle — there's no way to
move or insert it, because IDs are append-only. Fixing the order means deleting every bar
after the flaw and recreating them, which churns IDs that commit messages and notes already
reference. A plan-review step that can't cheaply reorder is working against its own purpose.

## Summary

Separate identity from order. The dotted ID stays a stable, never-changing handle; a track
gains an explicit ordered list of its bars that the CLI owns — `create` appends to it, a new
reorder operation rewrites it, `delete` removes from it. Crucially there is one ordering
source: `List()` / child-ordering is the single chokepoint that the CLI (`task list`,
`--parent`), the TUI list, the kanban board, and bit_do's "next step" resume all consume. Fix
that source and every surface reflects the new order by construction; the work is to point it
at the list and verify each surface. Illegal orderings can't be expressed — it's a total order
in a list, so no cycles or forks. Reordering is a bit_plan-time capability; bit_do stays a
straight-line executor and hands control back on a flaw. Existing tracks are backfilled from
their current ID sequence so nothing regresses. ID minting is unchanged, and decoupling
ID-reuse/backfill-on-delete is deliberately a separate follow-up.

## Phases

- [ ] Phase 1 — Resequence a bar, reflected everywhere: during bit_plan you move an existing
  bar to a new position in its track, and the new order shows up consistently across every
  surface that displays it — `bit task list` and `--parent` (so bit_do resumes on the correct
  next step), the TUI list, and the kanban board. Existing tracks keep working because their
  order is backfilled from today's ID sequence. This is the walking skeleton: it forces order
  onto the parent, a rewrite operation, and the single ordering source onto the manifest — and
  since all four surfaces consume that one source, "correct order everywhere" is proven here.
  Touches: `task/store.go` (List / child ordering, `compareIDs`), a new reorder command in
  `cmd/`, `cmd/task_list.go`, the TUI (`cmd/tui.go`, `tui/model.go`, `tui/board.go`), and the
  shared contract `.claude/bit-cli.md` + the bit_plan skill.

- [ ] Phase 2 — Insert a new bar mid-plan: `bit task create --after <bar>` drops a new step
  into a chosen position instead of only appending to the end — the "add a missing step" half of
  the flaw-fixing workflow, on top of Phase 1's order foundation. It flows through the same
  ordering source, so it too appears correctly in the CLI, the TUI list, and the board.
  Touches: `cmd/task_create.go` and the order-append logic in `task/store.go`.

## Visual aid

```
Before — order IS the ID (append-only):
  order:  BIT-7.1  BIT-7.2  BIT-7.3
  insert between .1 and .2  ->  must delete & recreate .2, .3 (new IDs, broken refs)

After — track owns the order; IDs are fixed identity:
  BIT-7 order: [ BIT-7.1, BIT-7.3, BIT-7.2 ]
  move .2 after .3          ->  rewrite one list; IDs untouched
  insert new .4 after .1    ->  [ BIT-7.1, BIT-7.4, BIT-7.3, BIT-7.2 ]

One source feeds every surface:
  order list --> List() / child-ordering --> { CLI task list & --parent,
                                               TUI list, kanban board,
                                               bit_do next-step resume }
```

## Risks & unknowns

- **Unknown:** Is `List()` / `compareIDs` (plus the `--parent` filter) truly the *only* place
  order is computed, with the TUI list, the board, and bit_do all mere consumers of it? If some
  surface re-derives order independently, it becomes a second source of truth.
  **Resolve by:** trace every consumer of `List()` at the start of Phase 1 and confirm the board
  (`groupByStatus`) and TUI list preserve incoming order rather than re-sorting.
  **De-risk before planning?** No — it's Phase 1's opening audit and cheap; the plan just needs
  to name every surface so none is missed.

- **Resolved (bit_plan):** the order list lives in the track's YAML frontmatter as `order:` —
  an `Order []string` field holding **full dotted bar IDs** (`["BIT-1.2", "BIT-1.1"]`), not bare
  numbers. IDs stay the identity handle, so a manifest entry maps directly to a loaded task and
  its `<id>.md` file with no ID re-composition. It round-trips for free through the existing
  yaml marshal (`Parse`/`Bytes`), and a stale entry (deleted bar) is harmless — `List()` only
  positions files it loaded. See bar BIT-8.1.

- **Unknown:** Global `bit task list` (no `--parent`) interleaves tracks with their bars via
  `compareIDs`, and the board groups bars from different tracks into one status column. Within a
  track the manifest order must hold; across tracks a sensible default order must remain.
  **Resolve by:** handle alongside `--parent` in Phase 1 and verify on the board.
  **De-risk before planning?** No — contained within Phase 1.