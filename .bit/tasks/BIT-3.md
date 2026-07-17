---
id: BIT-3
title: Plans live in the project, under their scope
status: todo
---
# Plans live in the project, under their scope

## Why

`bit` holds half the picture. The two scopes are in `.bit/` (BIT-1, BIT-2), but the
plans that implement them are still loose markdown at the project root, invisible to the
tool. So `bit` can tell you what work exists in the large, but not what's outstanding —
the 13 steps that actually delivered BIT-2 aren't in it, and nothing expresses that they
belonged to BIT-2 at all. That missing parent link is the top open question in the
README, and it blocks everything downstream: the TUI has almost nothing to render, and
`bit_plan` has nowhere to write.

It's also already unreadable at the size it's about to become. `bit task list` sorts
lexically, so ten tasks come back `BIT-1, BIT-10, BIT-11, BIT-2 …`. Importing the
existing plans roughly triples the record count, which makes a broken list the first
thing anyone sees.

## Summary

A step becomes a task parented to its scope, addressed by a dotted ID — `BIT-2.5` is the
5th step of `BIT-2`. The parent is readable straight from the ID, with no lookup and no
index. There is no plan record: "the plan for BIT-2" is just BIT-2's children. The
phases a plan groups its steps into survive as a label on each step, not as a node,
because a phase is something you want to *see on* a step, never something you open.

The vocabulary for all of this — album, track, verse, bar — is already mapped in
[hierarchy.md](./hierarchy.md), which this scope is written against. That naming is
presentational: the code keeps saying `task`, `scope`, and `step`.

The finish line is importing this project's own two plans, the same way the scopes were
imported — proving the model against real, awkward, backtick-heavy content rather than a
fixture.

## Phases

- [x] Phase 1 — a user reading `bit task list` sees the newest work first, in the right
  order: the list comes back `BIT-10, BIT-9 … BIT-1` instead of the current lexical
  `BIT-1, BIT-10, BIT-11, BIT-2`. Recent work is what you're almost always looking for,
  and it's the one thing the list can't currently show.
  Touches: `task/store.go` (`List`), `cmd/task_list.go`

- [x] Phase 2 — a user can record that a step belongs to a scope: creating a task under
  `BIT-2` gives it the next free dotted ID (`BIT-2.1`, `BIT-2.2` …), and the list stays
  coherent with parents and children mixed together. This is the parent link the README
  has had open since CRUD landed; after this phase, a plan is expressible in the tool.
  Touches: `cmd/task_create.go`, `task/store.go` (ID assignment, ordering)

- [x] Phase 3 — a user can see which phase of the scope a step serves without opening
  the scope: a step carries its phase, and `bit` shows it. This is the indicator the TUI
  needs to let you review bars one at a time without losing the thread of which slice of
  work they're building. It has to be visible in output, or it isn't worth storing.
  Touches: `task/task.go` (frontmatter), `cmd/task_create.go`, `cmd/task_update.go`,
  and whichever of `cmd/task_read.go` / `cmd/task_list.go` surfaces it

- [ ] Phase 4 — the project's own plans are in the project: `cli-bootstrap-plan.md` and
  `task-crud-plan.md` are imported under BIT-1 and BIT-2, and this scope's own work lands
  as BIT-3 with its plan's steps beneath it — steps parented and phase-labelled, driven by
  the prompt below. This is the acceptance test — the same one the scope import passed —
  and it's what finally seeds `.bit/` with enough real records for the TUI to be worth
  looking at.
  Touches: `.bit/tasks/` (no code)

## Visual aid

```
.bit/tasks/
├── BIT-1.md         track   CLI Bootstrap                    <- scope, no dot
├── BIT-1.1.md       bar     phase: 1                         <- step, dot = child
├── BIT-1.2.md       bar     phase: 1
├── BIT-2.md         track   Task Management (CRUD)
├── BIT-2.1.md       bar     phase: 1 — init wizard + create
├── BIT-2.6.md       bar     phase: 2 — list & read
└── BIT-2.13.md      bar     phase: 4 — delete

verse (phase) is a label on a bar, not a file.
"the plan for BIT-2" = grep 'BIT-2\.'
```

## The import prompt (Phase 4)

`bit_plan` will eventually do this itself; updating it is out of scope. Until then this
is the manual stand-in — the same shape as the prompt that imported the scopes:

> Import the existing plans into `.bit/` using `./bin/bit`, in delivery order —
> `cli-bootstrap-plan.md` under BIT-1, then `task-crud-plan.md` under BIT-2, then
> `plan-hierarchy-scope.md` as a new track (BIT-3) with `plan-hierarchy-plan.md`'s steps
> beneath it. For each plan: fold its preamble (Context, How this plan works) into the
> parent track's body, then create one child task per `## Step N`, in order, so the dotted
> IDs match the step numbers. Title it after the step's name, label it with the scope
> phase the step is tagged to, and set its status from the step's `**Status:**` line. The
> body is the step's section verbatim. Then verify by reading each one back and diffing
> against the source — don't eyeball it, and remember `$(cat file)` strips trailing
> newlines.

## Risks & unknowns

- **Unknown:** How a step names its phase. `phase: 1` is a positional reference into the
  parent track's body, and nothing validates it — renumber the scope's phases and every
  child silently points at the wrong one. Storing the label (`phase: "2 — List & read"`)
  is self-describing but duplicates text that can drift.
  **Resolve by:** Decide while planning Phase 3 — both options are a frontmatter field,
  and the import in Phase 4 is what will actually show which one reads better.
  **De-risk before planning?** No — cheap to change, and this scope's whole job is to
  find out.

- **Unknown:** Where a parent sits relative to its children under reverse sort. Verified
  today: with `BIT-2.1` and `BIT-2.13` on disk, `List` returns `BIT-1, BIT-2.1,
  BIT-2.13, BIT-2` — the parent sorts *after* its own children, because `"BIT-2.13.md" <
  "BIT-2.md"` lexically. Phase 1 fixes ordering before dotted IDs exist, so Phase 2 will
  have to extend it rather than inherit it.
  **Resolve by:** Let Phase 2's tests contradict Phase 1's ordering — that's the forcing
  function, and it's exactly why Phase 1 shouldn't try to guess the hierarchy early.
  **De-risk before planning?** No — already reproduced, and it's a plan-level detail.

- **Unknown:** Whether folding a plan's preamble into the track's body muddies the one
  view you most want clean — the full-scope approve/disapprove. The track body would
  then hold both scope prose and plan prose.
  **Resolve by:** Read `BIT-2` back after the Phase 4 import and judge it. If it reads
  badly, the preamble is droppable — it's the least load-bearing text in either plan.
  **De-risk before planning?** No — the import is the experiment.

- **Known, not unknown:** dotted IDs make reparenting a step a file rename, and only
  append-only numbering is cheap. Accepted: plans grow at the end, and a step that moves
  to another scope is rare enough to pay for by hand.

## Out of scope

- Updating `bit_scope`, `bit_plan`, or `bit_do` to write into `.bit/` — the import prompt
  above is the deliberate stand-in. Settling this relationship is the prerequisite for
  that work, not part of it.
- A phase-level command or view (`bit phase …`), which you'd only want if you ever
  reviewed a whole phase at once.
- Filtering (`bit task list --parent BIT-2`) — natural next, but `grep` covers the import
  verification, so nothing here demands it yet.
- **Cascading delete. Orphans are fine.** `bit task delete BIT-2` removes one file, so
  deleting a track leaves its bars behind pointing at a parent that's gone, with no
  warning. That's accepted, not overlooked — nothing here deletes a track, and cascade
  semantics aren't worth designing before there's real data to delete.
- Renaming anything in the code to the music vocabulary.

## Context

See scope: [plan-hierarchy-scope.md](./plan-hierarchy-scope.md) — the WHY and the phase
order live there. Vocabulary: [hierarchy.md](./hierarchy.md).

Recap: a step becomes a task parented to its scope by a dotted ID (`BIT-2.5`), so `bit`
can finally express the plan that delivered a scope — and the TUI has something to render.

## How this plan works

The entry point is `bit task list`, and the whole plan is driven by what it prints.

Step 1 makes it newest-first with the cheapest possible change — reversing the existing
lexical sort. Step 2 contradicts that with `BIT-10` vs `BIT-9`, which reverse-lexical
gets wrong, forcing a real numeric sort. Step 5 contradicts *that* with a dotted ID,
which flat numeric ordering gets wrong, forcing the hierarchy into the comparison. No
step knows about parents until a test makes it impossible not to.

The `--parent` flag arrives the same way: Step 3 can satisfy its test by hardcoding
`.1`, Step 4's second child can't.

The ordering being driven toward, settled with the user:

```
BIT-2       track   <- tracks descending, newest first
BIT-2.1     bar     <- its bars ascending, in plan order, directly under it
BIT-2.13    bar
BIT-1       track
BIT-1.2     bar
```

Sort key per task: `(track number DESC, bar number ASC)`, where a track's own bar number
is 0 so it always heads its group.

**Assumptions worth contradicting now rather than in review:**

- `--parent BIT-99` is not validated — a typo silently creates a bar under a track that
  doesn't exist. Consistent with the scope's "orphans are fine"; no test demands it.
- Only two levels exist. `--parent BIT-2.1` would mint `BIT-2.1.1`, which the sort would
  mishandle. Not rejected, not tested — [hierarchy.md](./hierarchy.md) says two levels,
  and nothing here creates a third.
- `phase` (int) and `phase_label` (string) are both plain frontmatter, unvalidated
  against the parent's body. The number is what code sorts and groups on; the label is
  for reading.