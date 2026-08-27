---
id: BIT-41.5
title: Observe a consumer against the pushed 0.1.0
status: todo
phase: 3
phase_label: 'spike: update'
---
## **Verse 3 (spike)**

**Question:** Does an already-installed consumer pick up a newly pushed, tagged version on its
own, only when told, or not at all — and what can a consumer cheaply read to learn "latest"?

**Yes/automatic looks like:** `claude plugin list --json` reports `"version": "0.1.0"` for
`bit@bit-pro` with no explicit command run.
**Yes/manual looks like:** it keeps reporting the git sha until `claude plugin update
bit@bit-pro`, and reports `0.1.0` after.
**No looks like:** it reports the git sha either way. That is a real answer the bar succeeds at
producing, not a failure — it means Verse 4 must source "latest" itself.

## Scope
- The installed `bit@bit-pro` plugin on this machine — read only. Nothing is built here and
  nothing is thrown away; the release being observed (`0.1.0`) is real and kept.

## References
- `START-HERE.md`, the "Versioning" section — the baseline this observes against: the plugin
  currently lists as `"version": "4ebbe7cd5eff"` (the git-sha fallback), and
  `claude plugin list --json` was measured at ~0.3s. Its open question — "Does Claude auto-detect
  a version bump once we start versioning? Unverified" — is exactly this bar.

## Method
- [ ] Precondition: `v0.1.0` is on origin (BIT-41.4) **and** the release commit is on
      `origin/main` (`git push` — `release-push` deliberately doesn't move it). Confirm with
      `git ls-remote --tags origin 'v*'` and `git ls-remote origin refs/heads/main`.
- [ ] Record the before state: `claude plugin list --json` for `bit@bit-pro`, and the cache
      directory name under `~/.claude/plugins/cache/bit-pro/bit/`. Today that name is a git sha
      (`4ebbe7cd5eff`, `fb11adeb8621`), while a versioned plugin caches under its version instead
      (`~/.claude/plugins/cache/pydantic-skills/ai/0.1.0/`) — so the directory name is itself one
      of the observations.
- [ ] Without running any update command, re-read `claude plugin list --json`. Then restart a
      Claude session and re-read, since "on its own" plausibly means at session start.
- [ ] Then `claude plugin marketplace update bit-pro` and re-read. Then
      `claude plugin update bit@bit-pro` and re-read. Record which step, if any, changed the
      reported version — these are two separate events, and `claude/sync.go` already runs them in
      that order.
- [ ] Inventory the candidate "latest" sources and time each: `git ls-remote --tags origin
      'v*'`, the `marketplace.json` entry, the plugin cache directory name, and
      `claude plugin list --json`. Record cost and whether each works offline — the Decision that
      the notice must never block, delay or fail `bp` makes cost part of the answer, not a
      footnote.
- [ ] Record every observation verbatim in this bar's body: `bp task read BIT-41.5 --body` into a
      file, append an `## Observed` section, then `bp task update BIT-41.5 -d "$(cat …)"`. This is
      what lets BIT-41.6 and the hand-back cite measurements instead of recollection.

## Claude verifies
- [ ] Every command above ran and its output is recorded in the `## Observed` section.
- [ ] The answer is stated as exactly one of automatic / manual / neither — not "seems to work".

## User verifies
- [ ] Read the `## Observed` section against your own `claude plugin list --json`: the version it
      reports for `bit@bit-pro` matches what the bar recorded.

## Commit (user)
`docs(bit): record what an installed consumer sees when 0.1.0 is published`