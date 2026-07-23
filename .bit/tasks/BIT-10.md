---
id: BIT-10
title: Archiving & soft deletes
status: doing
order:
    - BIT-10.1
    - BIT-10.2
    - BIT-10.3
    - BIT-10.8
    - BIT-10.4
    - BIT-10.5
    - BIT-10.6
    - BIT-10.7
    - BIT-10.9
---
## Why

A `done` track never leaves the list or the board — finished work piles up in the done
column and crowds out the live work someone actually needs to act on. That's the pain felt
today. The same clutter has a mirror image in deletion: `bit task delete` calls `os.Remove`
and destroys the markdown file outright, so a mistaken delete loses the work with no undo
short of git. Both share one fix: instead of leaving a file in the way or destroying it,
move it out of `tasks/` into a sibling folder the list and TUI don't scan. Finished work
gets out of the way, and a mistaken delete becomes recoverable — one mechanism, several
triggers.

## Summary

Introduce a "move the file aside" operation at the store layer, into a single sibling folder
(`.bit/archive/`) that `Store.List` never globs — so a relocated file drops out of the list,
the board, and `--parent` views for free. Relocating **reserves the ID**: `NextID`/
`NextChildID` count archived files, so archiving `BIT-12` still leaves `BIT-13` next and a
stored ID is never silently re-minted onto a different task. Triggers wire onto this primitive
in delivery order: an archive action (declutter done work — the felt pain), a non-destructive
`bit task delete` (recoverable), a fix so relocating keeps the parent track's bar ordering
consistent, and finally the bit_* skill lifecycle so a track gets archived when the human
explicitly signs it off as done.

## Visual aid

```
before                          after (archive BIT-9 / delete BIT-9)
.bit/                           .bit/
├── config.toml                 ├── config.toml
├── tasks/                      ├── tasks/
│   ├── BIT-8.md                │   └── BIT-8.md
│   └── BIT-9.md   ──relocate──▶└── archive/          (List / TUI never glob here;
                                    └── BIT-9.md       NextID still counts it, so its
                                                       ID stays reserved)
```

## Decisions

- **One shared folder.** Archived and soft-deleted files both land in `.bit/archive/`; with no
  restore/view command, nothing downstream needs to tell "finished" from "thrown away" apart.
- **Relocating reserves the ID.** Archiving `BIT-12` does not free `12` — `NextID`/
  `NextChildID` count `archive/`, so the next ID is `13`. A stored ID is stable identity and is
  never silently re-minted onto a different task.
- **Relocating a bar updates its parent track's `Order`.** The relocate primitive renames the
  file but never touched the parent's ordering list, so an archived or deleted bar left a
  phantom entry behind. Relocating must also drop the bar's id from its parent's `Order`,
  keeping the "every write goes through `bit`, so the list ⇄ files stay in sync" invariant true
  on the relocate path too. (A track relocate is covered for free — its bars' parent is the
  track being removed.)
- **Explicit human sign-off marks a track done — not auto-rollup.** Finishing the last bar
  rolls its verse up, but the track is *not* auto-flipped to `done`; the human makes that call
  after a final check, and it's that call that triggers the archive. So the lifecycle verse must
  change bit_do's rollup to stop short of auto-`done` and leave the final track-done to the person.
- **Relocating a track cascades to its bars.** Archiving or deleting a track moves the track
  *and* all its bars into `.bit/archive/` in one action, so a finished (or discarded) track
  never leaves its bars cluttering the list — the declutter goal holds on every path. The
  store primitive can stay a single-file move; the command loops over the children.
- **A track only relocates when all its bars are `done` — otherwise it fails.** Archiving or
  deleting a track with unfinished bars would tear down live work, so the command refuses and
  reports which bars aren't done. A `--force` flag overrides for the deliberate case
  (abandoning an in-progress track). A leaf bar has no children, so the guard is vacuous there.

## Verses

- [x] Verse 1 — Clearing finished work out of the way: an archive action relocates a track
  and its bars into `.bit/archive/`, so `bit task list`, the board, and the TUI show only live
  work — and the relocated IDs stay reserved, never re-minted. Refuses unless every bar is
  `done` (`--force` overrides). This is the pain felt today, and it proves the relocate
  mechanism end to end.
  Touches: a new `cmd/task_archive.go`, `task/store.go` (the relocate primitive; `NextID`/
  `NextChildID` scan `archive/` so IDs stay reserved).
- [x] Verse 2 — Deleting a task no longer loses it: `bit task delete <id>` reuses the relocate
  primitive instead of `os.Remove`, so a fat-fingered delete is survivable — the file is still
  on disk, its ID stays reserved, and deleting a track takes its bars with it (same all-done
  guard and `--force` as archive).
  Touches: `task/store.go` (`Delete` relocates instead of `os.Remove`), `cmd/task_delete.go`
  (confirmation wording).
- [ ] Verse 3 — Relocating keeps a track's bar ordering honest: after a bar is archived or
  deleted, its id is dropped from the parent track's `Order` list too, so the ordering never
  carries a phantom entry pointing at a bar that's no longer live — restore and anything that
  reads `Order` stay trustworthy. Discovered after Verses 1–2 shipped: the relocate primitive
  renamed the file but never updated `Order` (only `move`/`create` ever wrote it), so the
  documented "`delete` removes from the order list" contract was quietly false. Harmless today
  only because `List` globs `tasks/` — but Verse 4 introduces dropping a bar mid-plan by
  archiving it, which is exactly when a *live* track would be left holding a phantom entry, so
  the primitive needs fixing before that workflow ships on top of it.
  Touches: `task/store.go` (the relocate primitive also drops the id from the parent's `Order`),
  `assets/bit-cli.md` (the "`delete` removes from it" line becomes accurate).
- [ ] Verse 4 (last) — Archiving becomes part of the lifecycle: update the bit_* skills (via
  skill-creator) so a track is archived on the human's explicit sign-off (per the decision
  above — a deliberate final call, not an automatic status flip), and a step dropped mid-plan
  is archived rather than `delete`d so its ID isn't freed. Finished and abandoned work leaves
  the active view as a normal beat of the workflow, not a command someone has to remember.
  Gotcha: the skills `bit init` seeds live in the embedded `assets/skills/` copy, **not** the
  repo's `.claude/skills/` — and the copy is embedded at build time, so a sync is edit
  `assets/skills/` → `just install` (re-embed) → `bit init` (reseed `.claude/`). Editing only
  `.claude/skills/` would leave `init` shipping stale copies.
  Touches: `assets/skills/bit_do/` (the seeded source of truth, incl. the rollup change from the
  decision above; and `bit_plan`/`bit_scope` if their lifecycle docs mention archiving), the
  repo's own `.claude/skills/` copies kept in sync, and `assets/bit-cli.md` (rephrase the
  "status spelling matters" rollup example — all-bars-`done` no longer auto-flips a track).