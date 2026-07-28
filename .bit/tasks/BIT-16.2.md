---
id: BIT-16.2
title: The repo installs as its own marketplace
status: done
phase: 1
phase_label: 'spike: delivery'
---
## **Verse 1 (spike)**

**Question:** Can this repo publish itself as a marketplace over **GitHub** and have the `bit`
plugin install from it at project scope — given the repo is private?

**Yes looks like:** `claude plugin marketplace list` shows `bit-pro`;
`~/.claude/plugins/installed_plugins.json` gains a `"bit@bit-pro"` key whose record has
`"scope": "project"` and this repo's `projectPath`; and a session started **without**
`--plugin-dir` still offers `bit:ping`, served from
`~/.claude/plugins/cache/bit-pro/bit/<version>/`.

**No looks like:** `marketplace add` fails to authenticate against the private repo, or `install`
reports the plugin not found, or the relative `"source": "./bit"` in the marketplace manifest
isn't resolved. Any of these is a real result — it means distribution needs a public repo or a
different source type, which changes Verse 3's wiring.

This bar establishes the **baseline install record** that the next bar watches change. Without a
real install there is nothing for an update to arrive at.

## Scope
- `.claude-plugin/marketplace.json` at the **repo root** — `name: "bit-pro"` (this is the
  marketplace name the `bit@bit-pro` id refers to), and a `plugins` array with one entry:
  `name: "bit"`, `source: "./bit"`. **Kept** — Verse 2's real manifest.
- `.claude/settings.json` is modified as a *side effect* of `install --scope project`. Not hand-
  edited here; Verse 3 is what makes `bp init` write it deliberately.

## References
- `~/.claude/plugins/marketplaces/go-skills/.claude-plugin/marketplace.json` — the exact field
  shape for a single-repo-as-marketplace: top-level `name`, `plugins[]` with relative `source`.
- `https://code.claude.com/docs/en/plugin-marketplaces` — marketplace manifest and source types.

## Method
- [ ] Write `.claude-plugin/marketplace.json`
- [ ] Run `claude plugin validate .` on the repo root (validates the marketplace manifest)
- [ ] **Hand off to the user to commit and push.** The rest of this bar cannot run until the
      manifest is on `origin/main` — a marketplace added from GitHub reads the pushed repo, not
      the working tree.
- [ ] `claude plugin marketplace add B4Dmonkey/bit-pro`
- [ ] `claude plugin install bit@bit-pro --scope project`

**The source must be the GitHub repo, not a path.** `marketplace add` also accepts a local path,
and adding `.` would appear to work — but a path-source marketplace reads the working tree, so the
next bar's update test would pass without any push ever propagating. That is a false pass on the
one question this whole spike exists to answer.

## Claude verifies
- [ ] `claude plugin validate .` exits 0 (a missing-`version` warning is expected)
- [ ] `claude plugin marketplace list` includes `bit-pro`
- [ ] `~/.claude/plugins/installed_plugins.json` contains a `"bit@bit-pro"` key, and its record
      has `"scope": "project"` and this repo's path
- [ ] Record the record's `version`, `installPath`, and `lastUpdated` verbatim in the report — the
      next bar compares against these exact values, so they have to be captured before the edit

## User verifies
- [ ] Start plain `claude` in this repo — **no `--plugin-dir`** — and confirm `bit:ping` is
      offered. That proves it loaded from the installed cache rather than the working tree.
- [ ] `git status`: the only unexpected diff is `.claude/settings.json`, which the install wrote.
      Confirm that change is one you want checked in — Verse 3 later makes `bp init` write it.

## Report back
- [ ] Only if the answer is **No**: take it to bit_scope. A private-repo or relative-source
      failure changes the distribution story, which is the scope's central unknown, and Verses
      2–4 get revised before anything else is built.

## Commit (user)
`feat(plugin): publish the repo as its own plugin marketplace`