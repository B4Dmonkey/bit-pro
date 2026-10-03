---
id: BIT-49.29
title: do and bot-dev stop staging .bit/ with the code
status: todo
phase: 3
phase_label: no .bit/ in sight
---
## **Verse 3**

v1 committed `.bit/tasks/*.md` status changes in the same commit as the code. In v2 the store isn't in the repo, so those instructions point at nothing. The scope says to delete only the staging lines. Who commits, when it asks, and in what order is BIT-47's, so nothing here rewords bot-dev's commit behaviour.

## Scope
Re-grep first: `grep -n '\.bit' bit/skills/do/SKILL.md bit/agents/bot-dev.md`.
- `bit/skills/do/SKILL.md`:
  - `:75`: delete the last sentence, the one about marking it now to keep the `.bit/tasks/*.md` status change in the tree for the same commit. The rest of the line stays.
  - `:99`: delete the sentence "The `.bit/tasks/*.md` changes from steps 1–2 are part of the working tree, so they go into the same commit as the code — mention that." The rest of step 3 stays.
- `bit/agents/bot-dev.md:27-28`: delete the bullet "The `.bit/tasks/*.md` status changes … part of the same commit as the code" and the `plus \`.bit/\`` clause of the staging bullet. The rest of each line (stage what the bar touched, leave unrelated changes alone) stays word for word.
- These are deletions only, so no tests are added.

## Change checklist
- [ ] Delete the three pieces.
- [ ] `grep -rn '\.bit' bit/skills bit/agents bit/.claude-plugin` finds nothing.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/do/`. Only the known kebab-case failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `just lint` and `just test` pass

## User verifies
- none. The verse's end-to-end run is on BIT-49.30.

## Commit (user)
`docs(bit): do and bot-dev stop staging .bit/ with the code`