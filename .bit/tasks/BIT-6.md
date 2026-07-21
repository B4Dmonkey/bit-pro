---
id: BIT-6
title: Drive the full task lifecycle through bin/bit
status: done
---
# Drive the full task lifecycle through bin/bit

## Why

The bit_scope / bit_plan / bit_do skills are meant to track their own work *inside*
`.bit/` through `bin/bit`, so a track and its bars get created and advanced as a side
effect of doing the work — real dogfooding. Today they can't: the skills author
root-level `<feature>-scope.md` / `-plan.md`, and the work only reaches `.bit/` when a
human hand-imports it with throwaway scripts (that is literally how BIT-1..BIT-5 were
recorded). The blocker is `bin/bit` itself — a few everyday lifecycle moves can't be done
cleanly as a single CLI call, so the skills can't be rewritten to drive it. This scope
closes those CLI gaps. The skill rewrite is a separate pass that depends on this one.

## Summary

Four small, independent additions to the `task` subcommands so a caller driving `bin/bit`
(a skill) can run the whole **author → plan → implement → roll-up** loop through the CLI
without reading task files off disk: learn the ID it just minted, round-trip a body
cleanly, read one track's bars, and be stopped from orphaning a bar under a missing parent.

I re-walked the full lifecycle as `bin/bit` invocations against the current code
(`cmd/task_*.go`, `task/store.go`). The four phases below are the complete set of blockers
for that loop — no additional one surfaced. The remaining risk is that the loop hasn't yet
been driven end-to-end *by a skill*, only in pieces (see Risks).

## Phases

- [x] Phase 1 — Learn the ID you just created: `task create` reports the ID it minted
  (a track, or a `--parent` bar), so a caller can immediately author into it or hang bars
  off it instead of guessing with a list-before/after diff. This is the blocker the rest
  of the loop assumes.
  Touches: `cmd/task_create.go`.

- [x] Phase 2 — Round-trip a body cleanly: `task read` can emit just the task body with no
  summary header, so a caller can read → edit → write back (refine a scope, tick a phase
  checkbox) without string-peeling the header off first.
  Touches: `cmd/task_read.go`.

- [x] Phase 3 — Read one track's bars: `task list` can narrow to a single parent's
  children, so the auto-roll-up can derive a track's status from its bars without dumping
  every task and grepping the ID prefix.
  Touches: `cmd/task_list.go` (and a `task.Store` helper if the filter belongs there).

- [x] Phase 4 — Refuse to orphan a bar: `task create --parent X` fails when `X` doesn't
  exist, instead of silently minting a stray `X.1`, so a typo'd parent surfaces at the
  call site.
  Touches: `cmd/task_create.go`, `task/store.go` (`NextChildID`).

## Visual aid

Where each gap sits in the loop, driven as CLI calls:

```
scope   bit task create "<scope title>"           → ID?          [P1]
        bit task read   <id> --body               → clean body   [P2]
        bit task update <id> -d "<edited body>"    ✓  byte-safe today

plan    bit task create "<step>" --parent <id> …  → child ID?    [P1]
        ( --parent must exist )                   → validated?   [P4]
        bit task update <id.n> -d "<step body>"    ✓

do      bit task update <id.n> -s doing | done     ✓
        bit task list   --parent <id>             → bars only?   [P3]
        ( derive track status, then set it )       ✓
```

## Risks & unknowns

- **Unknown:** Does anything consume `task read`'s current `ID\tstatus\ttitle … \n\n body`
  format, such that a body-only mode must be *additive* (a flag) rather than a changed
  default?
  **Resolve by:** grep for callers/tests of the read format. Note the TUI reads via
  `task.Store.Load`, not the `read` command, so it's likely unaffected either way.
  **De-risk before planning?** No — a quick grep during planning settles flag-vs-default.

- **Unknown:** The full loop hasn't been driven end-to-end through `bin/bit` by a skill
  yet — the auto-roll-up (read a track's bars → set the track's status) is only proven in
  pieces.
  **Resolve by:** after these four land, dry-run one scope → plan → do cycle against a
  scratch `.bit/` before rewriting the skills.
  **De-risk before planning?** No — this validates the *skill rewrite* (the next pass),
  not this CLI scope.

## Out of scope

- `--from-file` for `task create` / `update` — not needed: `-d "$(cat file)"` is a proven
  byte-safe whole-body write (backticks, `$`, code fences, and `---` lines all round-trip
  identical). Revisit only if shell-quoting ever bites.
- `bit init` interactivity — a separate LLM shortcut initializes projects; in practice the
  project is pre-init'd when the skills run. Robust init is future work.
- The bit_scope / bit_plan / bit_do rewrites (and reconciling bit_check) to drive the
  now-capable CLI — the following skill-creator pass.
- Any status *state machine* — status stays a plain field the caller sets directly.

## Context
See scope: [cli-authoring-scope.md](./cli-authoring-scope.md) — the WHY and phase order live there.
Recap: close the four `bin/bit` gaps that stop a skill from driving the author → plan → implement → roll-up loop through the CLI.

## How this plan works
Every gap is a single `task` subcommand behavior, so each step's highest-level test is a
CLI invocation through the same `run`/`mustRun` harness the existing command tests use
(`cmd/cmd_test.go`). The throughline is the loop itself: Phase 1 is the walking skeleton
(a caller can learn the ID it just minted — nothing downstream works without it), then
Phase 2 lets that caller round-trip a body, Phase 3 lets it read one track's bars, and
Phase 4 stops it minting an orphan. Steps run in the scope's delivery order.

Design fork already settled (scope's Phase 2 risk): `read` gains an **additive `--body`
flag**, not a changed default. Grep confirms `newTaskReadCmd` is only wired into the
command tree — nothing shells out to `task read` or parses its stdout except the tests —
so a flag preserves the human summary view and keeps every existing `read` test green.

No new dependencies. Task runner is `just` (`just test`, `just lint`, `just build`).
Per project convention, tests carry no comments (no AAA markers).