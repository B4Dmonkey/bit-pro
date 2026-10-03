---
id: BIT-47.5
title: 'bit:plan''s bars end in a ## Commit section, and plan no longer says Claude never commits'
status: todo
approved: true
phase: 2
phase_label: the skills describe Claude committing
---
## **Verse 2**

From BIT-47.4 on, Claude commits through `bit:commit`, but bit:plan still says "**Claude never commits.**" and its templates still write `## Commit (user)`. This is a wording change only, so it has no Go test. `bit:commit` already reads both headings (BIT-47.3), so bars written before this change keep working. This bar completes the verse.

## Scope
Anchor on text, not lines (today `:304`, `:361`, `:398` at `6a1d345`; BIT-49.27 touched other lines of this file).
- `bit/skills/plan/SKILL.md`:
  - `## Verification split`, the paragraph "**Claude never commits.** The plan includes a suggested commit message per step, but committing is always the user's action." → "**Claude commits through bit_commit, and only after the operator says yes.** Each bar ends with a `## Commit` section holding its suggested message. bit_commit reads it, asks, commits, and records the hash on the bar."
  - The TDD bar template under `## Plan format`: `## Commit (user)` → `## Commit`.
  - The template under `### A spike bar's body`: `## Commit (user)` → `## Commit`.
  - Nothing else changes. The "Commit." in the contradiction recipe and "earns one commit" are generic.
- Text only, with no test files.

## Change checklist
- [ ] The three edits.
- [ ] `grep -n 'Commit (user)\|never commits\|always the user' bit/skills/plan/SKILL.md` finds nothing.
- [ ] `grep -rn -i 'commit (user)\|never commits\|user commits\|user runs the commit' bit/skills bit/agents` shows only `bit/skills/commit/SKILL.md`'s mention that older bars say `## Commit (user)`, and complete's report line, which BIT-50 rewrites.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/plan/`. Only the known kebab-case `name:` failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic (the greps above cover the whole verse)

## Commit
`docs(bit): bit:plan bars carry a Commit section that Claude commits`