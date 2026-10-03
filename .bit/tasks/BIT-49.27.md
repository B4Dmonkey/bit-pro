---
id: BIT-49.27
title: The pipeline skills and the analyze eval no longer name .bit/
status: todo
approved: true
phase: 3
phase_label: no .bit/ in sight
---
## **Verse 3**

analyze, check, complete, do, plan and scope still send Claude to `.bit/` paths in their descriptions and bodies. retro, learn and feedback were rewritten in BIT-49.8. This is a wording sweep with no behavior change, and it leaves the commit-staging lines to BIT-49.29.

## Scope
Re-grep first: `grep -n '\.bit' bit/skills/*/SKILL.md bit/skills/*/evals/*.json`. Line numbers are from 6a1d345.
- Wording rules (from the scope's Decisions):
  - Every "Never hand-edit `.bit/tasks/*.md`" becomes "the `mcp__bit__*` tools are the only way in", with no path named: `do :15`, `:132`; `plan :10`; `scope :20`; `complete :12`; `check :12`.
  - `.bit/completed/` mentions become "files it as completed": `do :95`, `:104`, `:108`, `:131`; `complete :3`, `:8`, `:23`, `:30`. This is a minimal wording change. BIT-50 rewrites the semantics, and BIT-47/BIT-50 anchor on section names, not lines.
  - Other location mentions ("find the track in `.bit/`", "`.bit/research/<track>/`") name the tool instead (`mcp__bit__task_read`, `mcp__bit__research_*`) or say "the store": `do :3` (twice), `:8`, `:43`; `plan :3`, `:8`; `scope :3`, `:10`, `:20`, `:110`, `:205`, `:227`; `analyze :3`, `:17`, `:64`.
  - `do :75`'s last sentence and `do :99` are left alone here (BIT-49.29).
- `bit/skills/analyze/evals/evals.json:14`: replace the `.bit/research/` expectation with the `mcp__bit__research_write` call it implies. This is the only eval that names `.bit/`.
- Text only, with no test files.

## Change checklist
- [ ] Apply the rules above.
- [ ] `grep -n '\.bit' bit/skills/*/SKILL.md bit/skills/*/evals/*.json` shows only `do/SKILL.md`'s two staging lines (`:75`, `:99`).
- [ ] `git diff --stat` touches only the six skills and the one eval file.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); for s in analyze check complete do plan scope; do uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/$s/; done`. Only the known kebab-case `name:` failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `python3 -m json.tool bit/skills/analyze/evals/evals.json > /dev/null`
- [ ] `just lint` and `just test` pass

## User verifies
- none. The verse's end-to-end run is on BIT-49.30.

## Commit (user)
`docs(bit): pipeline skills drop .bit/ paths`