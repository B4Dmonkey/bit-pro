---
id: BIT-27.3
title: Remove --bit-dir and BIT_DIR
status: done
phase: 1
phase_label: Canonical store
---
## **Verse 1**

The derivation now answers the question `--bit-dir` and `BIT_DIR` were built to answer, so they
come out. Nothing in the codebase or the plugin passes the flag or sets the variable — they live
only in `cmd/root.go` and the two tests covering them — and leaving them in would keep a second,
unused way to resolve the same path.

## Scope
- `cmd/root.go` — delete `bitDirFlag`, the `PersistentFlags().StringVar` registration, and the
  `os.Getenv("BIT_DIR")` branch. `os` stays imported for `os.Getwd`.
- `cmd/root_test.go` — delete `TestBitDirEnvVar_RoutesListToCanonicalDir` (it asserts the removed
  mechanism). Keep `TestBitDirEnvVar_DefaultIsRelativeDotBit` — it is the non-worktree
  contradiction the derivation depends on — but rename it to
  `TestBitDir_OutsideWorktreeUsesRelativeDotBit`, since its current name refers to a variable
  that no longer exists.

## TDD cycle

This bar removes behaviour rather than adding it, so there is no new RED test — the test being
deleted *is* the spec being retired. The surviving tests are what keep it honest: if deleting the
flag and env branches breaks resolution, `TestBitDir_OutsideWorktreeUsesRelativeDotBit` and the
two worktree tests from the previous bars go red.

1. **Remove:**
   - [ ] Delete `TestBitDirEnvVar_RoutesListToCanonicalDir` from `cmd/root_test.go`.
   - [ ] Rename `TestBitDirEnvVar_DefaultIsRelativeDotBit` to
     `TestBitDir_OutsideWorktreeUsesRelativeDotBit`; its body is unchanged.
   - [ ] Delete the `bitDirFlag` variable, its `StringVar` registration, and the flag/env branches
     from `PersistentPreRunE`, leaving the derivation as the only resolution.
   - [ ] Drop the now-unused `path/filepath` or `os` imports from `cmd/root_test.go` if the
     deletion orphans them.

2. **Confirm nothing else referenced them:**
   - [ ] `grep -rn "bit-dir\|BIT_DIR\|bitDirFlag" .` returns hits only in `.bit/` task bodies and
     `automation-notes.md` — prose, not code. Any hit under `cmd/`, `claude/`, or the plugin
     assets means this removal is incomplete.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes
- [ ] `just install`
- [ ] `bp --help` output contains no `--bit-dir`
- [ ] `BIT_DIR=/tmp/definitely-not-a-store bp task list` still lists the real board — proof the
  env read is gone, without a permanent test pinning the absence of a removed feature

## User verifies
- [ ] Whole slice — a write from inside a Claude Code worktree lands on the canonical board.
  From the main checkout:
  1. `git worktree add .claude/worktrees/verify-bit-27 -b verify-bit-27`
  2. `cd .claude/worktrees/verify-bit-27 && bp task create "worktree resolution smoke test" -d scratch`
     — note the ID it prints
  3. `cd` back to the main checkout and run `bp task list` — **the new ID is there.**
  4. Clean up: `bp task delete <that ID>`, then
     `git worktree remove .claude/worktrees/verify-bit-27 --force` and
     `git branch -D verify-bit-27`
  Before this verse, step 3 would show nothing — the task would have landed in the worktree's
  snapshot copy of `.bit/` and been lost with the branch.

## Commit (user)
`refactor(cmd): remove --bit-dir and BIT_DIR now that the store path is derived`