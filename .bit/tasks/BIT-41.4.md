---
id: BIT-41.4
title: just release-push publishes the tag
status: doing
approved: true
phase: 2
phase_label: publish
---
## **Verse 2**

`just release-push` publishes the tag as a separate, deliberate invocation. A local tag is
trivially deletable; a pushed one is not, so the irreversible half gets its own command. Its
first real use publishes `bit--v0.1.0` — the baseline BIT-41.5 observes against.

It pushes the **tag only**. Moving `origin/main` stays the operator's own `git push`.

## Scope
- `Justfile` — new `release-push` recipe, no arguments.

## Steps
- [ ] Dirty guard, same semantics as `release`: `git diff-index --quiet HEAD` refuses on tracked
      changes and ignores untracked files.
- [ ] Resolve the tag from `plugin.json`'s current version — `bit--v$version`. Refuse if that tag
      does not exist locally (`git rev-parse -q --verify "refs/tags/bit--v$version"`), naming the
      missing tag: that's the case of running `release-push` before `release`.
- [ ] `git push origin "bit--v$version"` — the tag only. This recipe deliberately does not move
      `origin/main`.
- [ ] Print the pushed tag.

## Claude verifies
- [ ] `just --list` shows `release-push` with its doc comment.
- [ ] With a tracked file modified, `just release-push` exits non-zero without contacting origin.
- [ ] With `plugin.json` naming a version whose tag does not exist locally, it refuses and names
      the missing tag.

## User verifies
- [ ] `git push` first (the branch — the recipe doesn't move it), then `just release-push`. This
      is the step that publishes, so it's yours to run. Confirm
      `git ls-remote --tags origin 'bit--v*'` lists `bit--v0.1.0`, and
      `git ls-remote origin refs/heads/main` points at the `release: v0.1.0` commit.

## Commit (user)
`feat(release): just release-push publishes the tag as a separate step`