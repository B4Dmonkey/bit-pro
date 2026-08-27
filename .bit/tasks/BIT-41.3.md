---
id: BIT-41.3
title: just release cuts the baseline
status: doing
approved: true
phase: 1
phase_label: cut a version
---
## **Verse 1**

`just release <level>` now acts: it writes the computed version into `plugin.json`, validates
strictly, commits, and creates the tag locally. Its first real use is this project's baseline —
after it, `bp --version` reports a number that can be compared to something.

The recipe commits, so **the operator runs the first real release**, not Claude. Claude's checks
here cover the recipe's shape and that BIT-41.2's guards still refuse.

## Scope
- `Justfile` — extend the `release` recipe past the guards.
- `bit/.claude-plugin/plugin.json` — gains `"version": "0.1.0"`, written by the recipe rather
  than by hand.

## Steps
- [ ] Write the computed version into `plugin.json`, preserving its other fields and formatting
      (`$schema`, `name`, `displayName`, `description`, `author`).
- [ ] `claude plugin validate bit --strict` — must exit 0 before anything is committed. With
      `version` and `author` both present this is the first point strict validation can pass. If
      it still refuses, stop and report rather than dropping `--strict`.
- [ ] Commit only that file: `git commit -m "release: v$next" bit/.claude-plugin/plugin.json`.
      The path-limited form keeps an unrelated staged change out of the release commit. Note the
      repo's pre-commit hooks run `just fmt`, `just lint` and `just test` at this point — a
      failure aborts the release, which is the correct outcome.
- [ ] `claude plugin tag bit` — creates `bit--v$next` on the release commit and revalidates that
      `plugin.json` and the enclosing `marketplace.json` entry agree. **Never pass `-f`**: its
      dirty-tree and tag-exists checks are the protection. No `--push` either — publishing is
      BIT-41.4's separate, deliberate step.
- [ ] Print the tag that was created.

## Claude verifies
- [ ] `just --list` still shows `release`, and `claude plugin tag bit --dry-run` prints the tag
      it would create without creating it.
- [ ] BIT-41.2's guards still refuse: bad level, off-main, dirty tracked tree, and the
      throwaway-`bit--v9.0.0` monotonic case. Nothing regressed by adding the action.

## User verifies
- [ ] Commit this bar first, then run `just release minor` yourself on a clean main — it writes a
      commit and a tag to your history, which is why it's yours to run. Then confirm:
      `git tag --list 'bit--v*'` → `bit--v0.1.0`; `git log -1 --format=%s` → `release: v0.1.0`;
      `git show --stat HEAD` touches only `bit/.claude-plugin/plugin.json`;
      `claude plugin validate bit --strict` exits 0; and `just install` followed by `bp --version`
      → `bp version 0.1.0`. Whole slice: a version got cut and reported, and you typed no version
      number to get it. (`just install` must be a separate `just` invocation — `version :=` is
      evaluated when the justfile loads, so a chained run inside one process would still hold the
      pre-tag value.)

## Commit (user)
`feat(release): just release <level> writes, validates, commits and tags`