---
id: BIT-41.15
title: Cut and publish v1.0.0
status: done
approved: true
phase: 6
phase_label: declare stable
---
## **Verse 6**

Cut and publish `v1.0.0` with the recipes Verses 1 and 2 built. This is the deliberate major the
scope names: a completed track would normally be a minor and no migration is involved, but until
this track lands no version on this project could be compared to anything, so a stable number
would have been asserting something untrue. It happens exactly once — BIT-39 lands as 1.1.0.

It is also the first time the notice fires against a genuinely published release rather than a
doctored clone, so the operator sees the whole chain — cut, publish, detect — end to end.

## Scope
- `bit/.claude-plugin/plugin.json` — `1.0.0`, written by `just release major`. No hand edit.
- `Justfile`, `scripts/` — unchanged. This verse *uses* the machinery; changing it here would mean
  the earlier verses were not finished.

## Steps
- [ ] Confirm the tree is clean and every earlier bar is committed — `just release` refuses a dirty
      tree, so this is the precondition, not a step to work around.
- [ ] `just release major` → `1.0.0`, the `release: v1.0.0` commit, and the local `v1.0.0` tag.
- [ ] `git push` — the branch. `release-push` deliberately does not move `origin/main`.
- [ ] `just release-push` — publishes the tag.
- [ ] `just install` — rebuilds `bp` so it reports `1.0.0`.

## Claude verifies
- [ ] `git tag --list 'v*'` lists `v0.1.0` and `v1.0.0`; `git ls-remote --tags origin 'v*'` lists
      both.
- [ ] `bp --version` prints `bp version 1.0.0`, and no version string was typed by hand anywhere —
      `major` was the only argument.
- [ ] `just test` and `just lint` pass.

## User verifies
- [ ] The chain, on a real release. After the push, the installed plugin is still `0.1.0` while
      origin publishes `1.0.0`:
      1. `bp task list` — silent on the first run if the clone has not caught up yet; it fires the
         detached refresh.
      2. `bp task list` again — stderr reads
         `bp: bit plugin 0.1.0 → 1.0.0 available — run: claude plugin update bit@bit-pro --scope project`.
      3. Run that exact command, verbatim, as printed. It must succeed — `--scope project` is the
         flag whose absence made it fail outright during the Verse 3 spike.
      4. `bp task list` — silent again. That silence is the whole track's answer to "am I running
         current?"

## Commit (user)
`chore(release): declare 1.0.0`