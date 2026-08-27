---
id: BIT-41.6
title: Observe a version-to-version bump at 0.2.0
status: todo
approved: true
phase: 3
phase_label: 'spike: update'
---
## **Verse 3 (spike)**

**Question:** Is a version-to-version bump (`0.1.0` → `0.2.0`) detected the same way as the
sha-to-version transition BIT-41.5 measured? The first could be detected as mere difference; the
second requires comparing semver — and the second is the case that will actually recur.

**Same looks like:** whichever step surfaced `0.1.0` in BIT-41.5 surfaces `0.2.0` here.
**Different looks like:** any divergence — the sha→version move showed up automatically but
version→version needs the explicit update, or the reverse.

## Scope
- `bit/.claude-plugin/plugin.json` — `0.2.0`, written by `just release minor`. Kept: `0.2.0` is a
  real release of this project, not a throwaway probe.

## Method
- [ ] Cut and publish `0.2.0` with the recipes: `just release minor`, then `git push`, then
      `just release-push`. This is also the recipes' second real exercise, and the first where the
      level actually determines the answer — a prior version exists now, so `minor` on `0.1.0`
      must produce `0.2.0` rather than the baseline `0.1.0`.
- [ ] Repeat BIT-41.5's observation sequence exactly: before state, no-command re-read, session
      restart, `claude plugin marketplace update bit-pro`, `claude plugin update bit@bit-pro`.
      Same commands, same order, so the two transitions are comparable.
- [ ] Record the observations in this bar's body the same way BIT-41.5 did, and state explicitly
      whether the two transitions behave identically.

## Claude verifies
- [ ] `git tag --list 'v*'` → both `v0.1.0` and `v0.2.0`; after `just install`,
      `bp --version` → `bp version 0.2.0`.
- [ ] Both transitions are recorded in the `## Observed` section with an explicit same/different
      verdict.

## User verifies
- [ ] Run `just release minor`, `git push` and `just release-push` yourself — they commit and
      publish. Then confirm `bp --version` reports `0.2.0` and that you typed no version number to
      get there. Whole slice: the recipes have now cut two releases, and the second one's number
      followed from the level alone.

## Report back
- [ ] Take the answer to bit_scope: the "does a pushed version reach an installed consumer"
      unknown becomes a Decision recording automatic / manual / neither, which source serves as
      "latest", and at what cost. Verse 4 is then revised against it and planned — it is
      deliberately unplanned until this lands, because an automatic update makes the notice nearly
      pointless while a manual one makes it obligatory and requires it to name the command to run.

## Commit (user)
`docs(bit): record the 0.1.0 -> 0.2.0 update observation`