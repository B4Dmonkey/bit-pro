---
id: BIT-41.2
title: just release computes and guards, without acting
status: done
phase: 1
phase_label: cut a version
---
## **Verse 1**

`just release <level>` exists but can only refuse or print: it computes the next version and
enforces every guard, while writing nothing, committing nothing and tagging nothing. Splitting
the guards from the action means the first thing exercised is the code that stops a bad release,
and this bar has no side effects to undo if a guard is wrong.

## Scope
- `Justfile` — new `release` recipe taking a `level` argument, as a `#!/usr/bin/env bash` recipe
  body with `set -euo pipefail` (the existing `install` recipe is the precedent for a shebang
  recipe in this Justfile).

## Steps
- [ ] Level guard: accept exactly `major` | `minor` | `patch`. Anything else exits non-zero,
      naming what was passed and the three accepted values. A version string is unrepresentable
      because the argument is never read as one.
- [ ] Branch guard: `git rev-parse --abbrev-ref HEAD` must be `main`, else refuse naming the
      current branch. Reason: `git describe` only sees tags reachable from HEAD, so a tag cut off
      main is invisible to the build that reads it while still visible to the monotonic guard's
      `git tag --list` — the two would disagree.
- [ ] Dirty guard, repo-wide: `git diff-index --quiet HEAD`. Tracked-and-uncommitted refuses;
      untracked files pass. This matches the semantics measured on `claude plugin tag`, but covers the whole
      repo — its check is scoped to the plugin directory, so a modified `cmd/root.go` currently
      sails through. (Measured today: `git status --porcelain -uno` shows ` M .bit/tasks/BIT-39.md`,
      so this guard fires as the repo stands.)
- [ ] Read the current version from `bit/.claude-plugin/plugin.json`; absent means "no prior
      version".
- [ ] Compute the next version: no prior version → `0.1.0` regardless of level (the baseline
      case); otherwise bump the named component and zero the components below it.
- [ ] Monotonic guard: highest existing tag is
      `git tag --list 'v*' --sort=-v:refname | head -1` — git's own version sort, so
      `0.10.0` outranks `0.9.0`. Strip `v`, then require the computed version to sort
      strictly above it: `printf '%s\n%s\n' "$highest" "$next" | sort -V | tail -1` must equal
      `$next`, and `$next` must differ from `$highest`. `sort -V` verified working on this
      machine. With no tags the guard passes trivially, which is the baseline case.
- [ ] Print the computed version and exit 0. Nothing else.

## Claude verifies
- [ ] `just --list` shows `release` with its doc comment.
- [ ] `just release bogus` exits non-zero and names the three accepted levels.
- [ ] With a tracked file modified, `just release minor` exits non-zero and says the tree is
      dirty; with only untracked files present it does not refuse.
- [ ] On a clean tree, `just release minor`, `just release major` and `just release patch` each
      print `0.1.0` — with no prior version the level cannot change the answer.
- [ ] Monotonic guard fires: `git tag v9.0.0` (local, throwaway), then `just release minor`
      exits non-zero naming both the computed `0.1.0` and the existing `9.0.0`. Clean up with
      `git tag -d v9.0.0`.
- [ ] `git status --porcelain -uno`, `git tag` and `git log -1` are unchanged after every run
      above. This bar's whole claim is that it has no side effects.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(release): just release <level> computes and guards without acting`