---
id: BIT-16.1
title: The skills/ layout loads and namespaces
status: done
phase: 1
phase_label: 'spike: delivery'
---
## **Verse 1 (spike)**

**Question:** Does one plugin holding its skills under `skills/<name>/SKILL.md` load at all, and
does the skill arrive namespaced as `/bit:<name>`?

**Yes looks like:** a session started with `claude --plugin-dir ./bit` offers the skill as
`bit:ping`, and invoking it returns the text from `bit/skills/ping/SKILL.md`.

**No looks like:** the skill doesn't appear, or appears un-namespaced as `ping`, or the plugin
fails to load. A No is a real result: it means the four-skill-one-plugin shape in the Decisions
is wrong and Verse 2's layout has to change — probably to one plugin per skill, which is what
the on-disk exemplar actually does.

This runs first because it needs no git, no push, and no network. If the layout is wrong, finding
out here costs a minute; finding out after a push means debugging distribution and layout at the
same time with no way to tell which one broke.

## Scope
- `bit/.claude-plugin/plugin.json` — the plugin manifest: `name: "bit"`, a description, and
  **no `version` field** (per the scope Decision, so the version resolves per-commit).
  **Kept** — this is Verse 2's real manifest, not scaffolding.
- `bit/skills/ping/SKILL.md` — a trivial probe skill whose body is one identifiable line.
  **Thrown away** — Verse 2 replaces it with the four real skills. Build it as cheaply as
  possible; it exists to be recognised in output, nothing more.

## References
- `https://code.claude.com/docs/en/plugins` — which directories sit at the plugin root vs inside
  `.claude-plugin/`, and skill namespacing. This bar is the check on that reading.
- `~/.claude/plugins/marketplaces/go-skills/go/.claude-plugin/plugin.json` — a valid manifest to
  copy the field shape from. Note it *does* carry a `version` and puts `SKILL.md` at the plugin
  root; we deliberately differ on both, which is why this bar exists.

## Method
- [ ] Write `bit/.claude-plugin/plugin.json`
- [ ] Write `bit/skills/ping/SKILL.md` with frontmatter (`name`, `description`) and a body
      containing one distinctive line — it has to be recognisable in a transcript later
- [ ] Run `claude plugin validate ./bit`

## Claude verifies
- [ ] `claude plugin validate ./bit` exits 0. It will **warn** about the missing `version` — that
      warning is expected and correct per the scope Decision. Do not pass `--strict`; it turns
      that expected warning into a failure.
- [ ] Both files are valid JSON / valid frontmatter (`plugin.json` parses, `SKILL.md` has a
      closing `---`)

## User verifies
- [ ] Start `claude --plugin-dir ./bit` in this repo. The `ping` skill is offered as `bit:ping`,
      and invoking it returns the distinctive line. This is the observation — validation passing
      only proves the manifest parses, not that the runtime loads the layout.

## Report back
- [ ] Only if the answer is **No**: take it to bit_scope. A No answers the scope's question early
      and differently — the plugin layout Decision is wrong, and Verses 2–4 get revised against
      the one-plugin-per-skill shape before anything is pushed. A Yes just unblocks the next bar.

## Commit (user)
`feat(plugin): add a bit plugin with a ping probe skill`