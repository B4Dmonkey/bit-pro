---
id: BIT-27
title: bp writes to the canonical .bit/ from inside a worktree
status: done
---
## Why

Task state written from a dispatched session is silently lost. `bp` resolves its store as the
literal relative path `.bit` (`cmd/root.go:12,30`), and a Claude Code session that edits files
always runs in a worktree under `.claude/worktrees/<slug>/`. Because `.bit/` is tracked in this
project, that worktree holds a full snapshot copy — so every status move, every rollup, and
every feedback note lands on the branch instead of in the canonical store.

The operator pays for this now. The board on main goes stale while work is in flight, status
moves have to be mirrored to the main checkout by hand, and the two feedback notes recorded
about this very bug (`BIT-23-003`, `BIT-23-004`) landed in a worktree rather than on main.
BIT-23 shipped `--bit-dir`/`BIT_DIR` as the escape hatch, but nothing sets it — it is prose
that has to be remembered, which is the failure mode it was meant to remove. It also blocks
the automation phase outright: unattended dispatch cannot depend on a human remembering to
export an environment variable.

## Summary

`bp` works out the canonical `.bit/` itself instead of trusting its cwd. When the working
directory sits inside `.claude/worktrees/<slug>`, everything from `.claude/` onward is cut off
to recover the main checkout, and the store resolves there. The `--bit-dir` flag and `BIT_DIR`
environment variable come out at the same time — they were the escape hatch for exactly this
problem and nothing ever set them.

One resolution point, no new configuration, no stored state, and no change to where `.bit/`
lives or how it is tracked.

## Visual aid

```
before                                          after
──────                                          ─────
cwd: <repo>/.claude/worktrees/bit-27/           cwd: <repo>/.claude/worktrees/bit-27/
  bp → ./.bit/            (branch copy)           cut at .claude/ → <repo>
       ✗ diverges, lost on abort                  bp → <repo>/.bit  (canonical)
```

## Decisions

- **Derive the path, never store it.** A computed answer self-corrects when the repo is moved
  or cloned. Recording the canonical path in frontmatter or config would commit a
  machine-specific absolute path into a tracked file, where it is wrong on every other machine.
- **`.bit/` stays tracked in-project, unchanged.** Tracking it here is deliberate — the
  project's own growth is visible in its own history. A global store under `~/.config`, a
  nested repo, an orphan branch, and a submodule were each considered and rejected: all of them
  trade "not entangled with branches" against "arrives with a clone", and none of them fix the
  resolution bug that causes the failure.
- **`--bit-dir` and `BIT_DIR` are removed.** Nothing in the codebase or the plugin passes the
  flag or sets the variable — they exist only in `cmd/root.go` and the two tests covering them.
  They were built to solve exactly the problem the derivation now solves, so keeping them would
  leave a second, unused way to answer the same question. Accepted tradeoff: there is then no
  manual override if Claude Code ever changes its worktree path convention, and the fix would
  have to be made in `bp`.
- **Only `.claude/worktrees/` is recognised, with no fallback.** The user creates branches, not
  worktrees; every worktree in play is created by Claude Code and lives at that path. This was
  measured, not assumed. A general upward walk for `.bit/` was considered and dropped as
  unnecessary.
- **Running `bp` from a project subdirectory stays unsupported.** It is broken today for the
  same missing-resolution reason, but it is not a workflow in use, so it is out of scope rather
  than fixed in passing.
- **A nested worktree resolves to the outermost checkout.** Cutting at the first
  `.claude/worktrees/` occurrence yields the true canonical store even for a worktree created
  inside a worktree.
- **Behaviour when the resolved `.bit/` is absent does not change.** Improving the missing-store
  error is a separate concern and is not bundled in here.

## Verses

- [x] Verse 1 — A `bp` call from inside a Claude Code worktree reads and writes the main
  checkout's `.bit/`: the operator can run any command from a dispatched session and see it
  land canonically, with the board on main staying live while work is in flight, and no
  `BIT_DIR` export and no hand-mirroring of status moves.
  Touches: `cmd/root.go` — the single `PersistentPreRunE` resolution point every command and
  the TUI already flow through.

## References

- `automation-notes.md` — the measured record of Claude Code's worktree behaviour (worktree
  path and branch naming, that editing is the trigger, that bars share one worktree). It is the
  authority behind the decision to recognise only `.claude/worktrees/`, and informs Verse 1.
- `.bit/feedback/BIT-23-003.md`, `.bit/feedback/BIT-23-004.md` — the two notes recording that
  the `BIT_DIR` mechanism exists but nothing sets it, and that the gap covers every `bp`
  subcommand rather than just task writes.