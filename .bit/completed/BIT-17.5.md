---
id: BIT-17.5
title: The contract teaches the write path
status: done
phase: 2
phase_label: The skill composes it
---
## **Verse 2**

Every bit skill learns the CLI from `bp instructions` and nothing else, so a write path missing from
that contract is a write path no skill can drive. This is a doc edit, not code — there is nothing to
test-drive, which is why its checks are greps rather than a red-green cycle. It comes before the
skill because the skill's first act is to read this.

## Scope
- `assets/bit-cli.md` — add a `## Capturing feedback` section between `## Writing a body from the
  shell` and `## Gotchas`. A note body is multi-line prose authored exactly like a task body, so it
  sits directly after the section that explains how to do that.
- `assets/bit-cli.md`, opening paragraph — the sentence naming the skills that drive `bp` currently
  lists scope, plan, do, and check; add feedback.

The new section covers, and nothing more:

```bash
# Record a feedback note against a track. Prints the note's path.
bp feedback add "$TRACK" -d "$(cat note.md)"
```

- Notes live in `.bit/feedback/`, one file per note, named `<TRACK>-NNN.md`.
- The write is **create-only**: there is no `feedback read`, `update`, or `delete`. A new note can
  never damage one already recorded, which is the point — capture fires at the least reliable
  moment in a cycle, right after a run went wrong.
- A note keys to a **track**, and cites its bar in the prose as data ("happened at BIT-11.4").
  Replanning renumbers bars and replanning is frequently the fix itself, so a note keyed to a bar
  would be orphaned by the very event it describes.
- The track may be active or archived; both are accepted, because retro reads finished tracks.
- Archiving a track leaves its notes in place — `task archive` moves files within `.bit/tasks/` →
  `.bit/archive/` and never touches `.bit/feedback/`.

Nothing about *evaluating* notes goes in here. `bit_retro` becomes their consumer in a later scope,
and the contract describes commands, not what a future skill will make of the files.

No test changes: `TestInstructionsCmd_PrintsContract` compares stdout against the embedded bytes, so
it follows any edit to the doc automatically.

## Claude verifies
- [ ] `just test` — the instructions test still round-trips the embedded contract
- [ ] `just lint`
- [ ] `go run . instructions | grep -c 'feedback add'` is at least 1
- [ ] `go run . instructions | grep -q 'feedback read'` finds nothing — the doc must not imply a
      read path that does not exist
- [ ] `just install`, so the `bp instructions` a skill runs is the edited one

## User verifies
- [ ] none — deterministic.

## Commit (user)
`docs(assets): document the feedback write path in the bp contract`