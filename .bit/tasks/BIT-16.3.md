---
id: BIT-16.3
title: A pushed edit reaches the installed plugin
status: done
phase: 1
phase_label: 'spike: delivery'
---
## **Verse 1 (spike)**

This is the bar the whole scope rests on. The two before it built and installed a plugin; this one
edits it, pushes, and watches whether the change arrives.

**Question:** Does pushing a skill change deliver it to a repo that already has the plugin
installed?

**Yes looks like:** after `claude plugin marketplace update bit-pro` and `claude plugin update
bit@bit-pro --scope project`, a **restarted** session with no `--plugin-dir` loads the *edited*
ping text. Corroborating: the `bit@bit-pro` record's `lastUpdated` has moved past the value
captured in the previous bar, and a new `installPath` exists under
`~/.claude/plugins/cache/bit-pro/bit/`.

**No looks like:** either update command reports nothing available, or a restarted session still
serves the old text. This is a real, useful outcome — it means shipping a skill fix needs a
reinstall rather than an update, and the "release on their own cadence" claim in the Why has to be
restated before Verses 2–4 are planned.

**Watch `lastUpdated`, not `gitCommitSha`.** The scope's risk entry names `gitCommitSha`, but the
one versionless plugin installed on this machine (`skill-creator@claude-plugins-official`) records
`"version": "unknown"` with **no `gitCommitSha` field at all** — while its `lastUpdated` did move.
Since this scope deliberately omits `version`, a missing SHA field is a plausible normal outcome
and must not be read as a No. The loaded text is the primary observation; the JSON fields
corroborate it.

## Scope
- `bit/skills/ping/SKILL.md` — change one identifiable line so the old and new text are impossible
  to confuse (e.g. a version marker in the body). **Thrown away** — Verse 2 replaces this file.

## References
- `https://code.claude.com/docs/en/plugin-marketplaces` — version resolution and the auto-update
  defaults. The scope notes auto-update is off by default for third-party marketplaces, so this
  bar tests the *explicit* update path only. Whether it can arrive unprompted stays unverified and
  is not this bar's question.

## Method
- [ ] Edit `bit/skills/ping/SKILL.md`, changing one distinctive line
- [ ] **Hand off to the user to commit and push.** Nothing after this can run until the edit is
      on `origin/main`.
- [ ] `claude plugin marketplace update bit-pro` — refreshes the catalog clone
- [ ] `claude plugin update bit@bit-pro --scope project` — must match the install scope from the
      previous bar; the flag defaults to `user`, which would target a plugin that isn't there
- [ ] Note what each command printed, verbatim, including a "nothing to update" — that output *is*
      the No result

## Claude verifies
- [ ] Both update commands exit 0
- [ ] The `bit@bit-pro` record in `~/.claude/plugins/installed_plugins.json` has a `lastUpdated`
      later than the value captured in the previous bar
- [ ] The edited line is present in the file under the record's `installPath` — this proves the
      cache itself was refreshed, independent of what any session loads

## User verifies
- [ ] **Restart the session** (`claude plugin update` says a restart is required) and start plain
      `claude` with **no `--plugin-dir`**. Invoke `bit:ping` and confirm it returns the *edited*
      line.

      This is the one check that must not be got wrong: `--plugin-dir` takes precedence over an
      installed plugin, so a session started with it would serve the working tree and show the
      edited text no matter whether the push propagated. That is a false pass on the scope's
      central question. Plain `claude`, restarted, or the observation is worthless.

## Report back
- [ ] Take the answer to bit_scope either way. The unknown becomes a Decision recording what was
      actually observed — which commands ran, what they printed, and which text loaded — and
      Verses 2, 3, and 4 get revised against it before they are planned.

## Commit (user)
`chore(plugin): edit the ping probe to observe update delivery`