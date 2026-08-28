---
id: BIT-41.5
title: Observe a consumer against the pushed 0.1.0
status: done
approved: true
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
## Observed

Recorded 2026-08-27, on this machine, against the real `v0.1.0` published by BIT-41.4.

**Answer: yes/manual.** An already-installed consumer does not pick up a pushed, tagged
version on its own — not on a re-read, and not at session start. It changes only when told,
and only by `claude plugin update bit@bit-pro --scope project`.

### Precondition

```
$ git ls-remote --tags origin 'v*'
f2a267c875411597df30c88bccfeb9ce4278ec81	refs/tags/v0.1.0
7804d163915c490f2abf2dd382751b45aff5d13b	refs/tags/v0.1.0^{}

$ git ls-remote origin refs/heads/main
6bf72fb911fe96ebe8a543845b896fa9aba6fb1f	refs/heads/main

$ git merge-base --is-ancestor 7804d16 origin/main   # release commit IS on origin/main
$ git log --oneline -1 7804d16
7804d16 release: v0.1.0
```

Tag `v0.1.0` created 2026-08-27 18:30:26 -0400, annotated, message `bit 0.1.0`.
`bit/.claude-plugin/plugin.json` reads `"version": "0.1.0"` both at the tag and at HEAD.

### Before state

```
$ claude plugin list --json   # bit@bit-pro entry
{
  "id": "bit@bit-pro",
  "version": "4ebbe7cd5eff",
  "scope": "project",
  "enabled": true,
  "installPath": "/Users/appstack/.claude/plugins/cache/bit-pro/bit/4ebbe7cd5eff",
  "installedAt": "2026-07-28T22:30:26.717Z",
  "lastUpdated": "2026-08-22T14:50:28.837Z",
  "projectPath": "/Users/appstack/Developer/UniqueDataManagement/tools/bit-pro"
}

$ ls -1 ~/.claude/plugins/cache/bit-pro/bit/
4ebbe7cd5eff  520d1227b319  5c3aeb1549f3  714008867658  dc3036b87ad2  eedc4d7bed1c  fb11adeb8621
```

Every cached directory is a git sha. `lastUpdated` is five days stale (Aug 22), so the
consumer was already behind the tag pushed at 18:30 the same day this ran.

### No update command run

Re-read with nothing run in between — unchanged:

```
4ebbe7cd5eff  /Users/appstack/.claude/plugins/cache/bit-pro/bit/4ebbe7cd5eff  2026-08-22T14:50:28.837Z
```

Session start — a fresh `claude -p` process in this project (2.5s), then re-read — unchanged:

```
4ebbe7cd5eff  /Users/appstack/.claude/plugins/cache/bit-pro/bit/4ebbe7cd5eff  2026-08-22T14:50:28.837Z
```

Corroborating: the interactive session that ran this bar started at ~19:20, *after* the
18:30 tag, and loaded its skills from `…/cache/bit-pro/bit/4ebbe7cd5eff/skills/do`. Session
start does not sync.

### Told to update

```
$ claude plugin marketplace update bit-pro          # 1.4s
Updating marketplace: bit-pro...Refreshing marketplace cache (timeout: 120s)…
✔ Successfully updated marketplace: bit-pro

# re-read: still 4ebbe7cd5eff at .../4ebbe7cd5eff, lastUpdated 2026-08-22T14:50:28.837Z
```

The marketplace refresh alone does **not** move the installed version. It does fetch: a new
`~/.claude/plugins/cache/bit-pro/bit/0.1.0/` directory appeared at 19:29:23, carrying
`.claude-plugin/plugin.json` with `"version": "0.1.0"` — cached but not installed.

```
$ claude plugin update bit@bit-pro                  # 0.5s — FAILS
Checking for updates for plugin "bit@bit-pro" at user scope…
✘ Failed to update plugin "bit@bit-pro": Plugin "bit" is not installed at scope user

# re-read: unchanged, still 4ebbe7cd5eff
```

`update` defaults to **user** scope. This install is **project** scope, so the bare command
in this bar's Method does not work — it errors without touching anything. `claude/sync.go`
already passes `--scope project`, which is what actually works:

```
$ claude plugin update bit@bit-pro --scope project  # 1.4s
Checking for updates for plugin "bit@bit-pro" at project scope…
✔ Plugin "bit" updated from 4ebbe7cd5eff to 0.1.0 for scope project
  (/Users/appstack/Developer/UniqueDataManagement/tools/bit-pro). Restart to apply changes.

$ claude plugin list --json   # bit@bit-pro entry
{
  "id": "bit@bit-pro",
  "version": "0.1.0",
  "scope": "project",
  "enabled": true,
  "installPath": "/Users/appstack/.claude/plugins/cache/bit-pro/bit/0.1.0",
  "installedAt": "2026-07-28T22:30:26.717Z",
  "lastUpdated": "2026-08-27T23:29:57.271Z",
  "projectPath": "/Users/appstack/Developer/UniqueDataManagement/tools/bit-pro"
}
```

So the cache directory name is itself an observation, and it confirms the pydantic-skills
precedent: a versioned plugin caches under its version (`…/bit/0.1.0`), a sha-versioned one
under its sha. `installedAt` is unchanged by an update; `lastUpdated` moves.

### Candidate "latest" sources

Each timed three times, warm. `real` seconds.

| Source | Cost | Offline | Reports |
| --- | --- | --- | --- |
| `git ls-remote --tags origin 'v*'` | 0.20 / 0.18 / 0.18 | no — unbounded hang, see below | true latest published tag |
| marketplace clone `bit/.claude-plugin/plugin.json` | 0.00 / 0.00 / 0.00 | yes | latest *as of the last marketplace update* |
| plugin cache directory name | 0.00 / 0.00 / 0.00 | yes | what has been fetched, not what is installed |
| `claude plugin list --json` | 0.28 / 0.28 / 0.28 | yes (verified under a black-holed proxy) | installed only — never latest |

The local marketplace clone at `~/.claude/plugins/marketplaces/bit-pro` is a full git
checkout of `origin/main` (HEAD `6bf72fb`, **no tags fetched** — `git tag -l 'v*'` is empty),
and its `bit/.claude-plugin/plugin.json` reads `0.1.0`. It is free to read and works
offline, but it is only as fresh as the last `claude plugin marketplace update`, which is
itself a network operation that self-bounds at 120s by its own report.

`.claude-plugin/marketplace.json` carries no version field, as the scope already recorded —
it is not a candidate on its own; the version lives in the clone's `plugin.json`.

**Network cost is the load-bearing measurement.** Against a black-holed route,
`git ls-remote` blocked for **75.0s** before failing. Git 2.50.1 (Apple Git-155) documents
no connect-timeout knob — `git help config` offers only `http.lowSpeedLimit` /
`http.lowSpeedTime`, which govern transfer stalls, not connect — so `GIT_HTTP_CONNECT_TIMEOUT`
has no effect (measured: it did not shorten the 75s). An **external** bound does work:
wrapping the same black-holed call in a 2s alarm returned in 2.03s, exit 142, no output.
Against a refused connection it failed in 0.02s.

For Verse 4 this means any network read must be wrapped by the caller (`exec.CommandContext`
with a deadline), never trusted to bound itself.

### What this settles for Verse 4

- The notice is **obligatory** — nothing updates on its own, so an operator can sit on a
  stale plugin indefinitely, exactly as this machine did for five days.
- The notice must **name the command**, and the command is
  `claude plugin update bit@bit-pro --scope project`. Without `--scope project` it fails
  outright on a project-scope install.
- Cheapest correct source is the marketplace clone's `plugin.json` (free, offline, no
  network), at the cost of being stale until a `marketplace update` runs. `git ls-remote` is
  authoritative but must be deadline-wrapped.
- `claude plugin list --json` answers "what am I running", not "what is latest" — it is the
  installed side of the comparison, not the published side.

### Not observed here

Whether a **version-to-version** bump (0.1.0 → 0.2.0) behaves the same as the sha → 0.1.0
transition observed above. That is BIT-41.6.