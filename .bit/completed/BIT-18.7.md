---
id: BIT-18.7
title: The contract teaches the new verb
status: done
phase: 3
phase_label: The session says complete
---
## **Verse 3**

`bp instructions` is the contract every bit skill reads before it drives the CLI, and it
still says `task archive` moves things into `.bit/archive/`. A session that trusts it files
finished work in the soft-delete bin. There is no red-green cycle here — nothing asserts the
contract's *content* (`TestInstructionsCmd_PrintsContract` only checks that the command
prints the embedded bytes) — so the checks are greps against the built binary.

Note the build step: `assets/bit-cli.md` is embedded with `go:embed`, so editing the file
changes nothing about the installed `bp` until `just install` runs.

## Scope
- `assets/bit-cli.md` — the commands block, the ordering note, the feedback section, and the
  ID-reservation gotcha.

## Edits
- [ ] Commands block: add `complete` next to the existing writes —
      `bp task complete "$TRACK"` for filing a signed-off track and its bars under
      `.bit/completed/`. Say that it refuses a track with an unfinished bar and has no
      override — mark the bars `done` first. Do not add `delete`; it was deliberately left out
      of the block and stays out.
- [ ] The ordering paragraph (`create` **appends** … `delete` removes from it): `complete`
      also drops a bar from its parent's order, so name both.
- [ ] Feedback section: the track may be active, **completed**, or archived — all three are
      accepted. And the note-survival line stops naming `task archive`: completing or
      deleting a track moves files within `.bit/` and never touches `.bit/feedback/`.
- [ ] The ID-reservation gotcha: `task delete` relocates into `.bit/archive/tasks/` and
      `task complete` into `.bit/completed/`; `NextID`/`NextChildID` count `tasks/`,
      `completed/`, and `archive/tasks/`. Keep the point that a removed ID is never
      re-minted — that hasn't changed, only where the files go.

## Claude verifies
- [ ] `just test`
- [ ] `just install`
- [ ] `bp instructions | grep -c "task archive"` is `0`.
- [ ] `bp instructions | grep -q "task complete"`, and the same for `.bit/completed/` and
      `.bit/archive/tasks/`.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`docs(cli): name complete and the two destinations in the bp contract`