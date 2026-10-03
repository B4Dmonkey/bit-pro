---
id: BIT-50.17
title: /bit:complete files a squash-landed track with each bar repointed to its own squash
status: todo
phase: 4
phase_label: squash-landed tracks
---
## **Verse 4**

The repoint step already exists (BIT-50.14) and is driven by `repoint: true`, so squash-landed bars are written before filing with no new flow. This bar makes the skill say so, adds the eval, and holds the verse's end-to-end check.

## Scope
- `bit/skills/complete/SKILL.md` (re-read first):
  - Where the repoint step is described, say that bars landed through a GitHub squash (`class: squash`) are repointed to their own squash, and the track records the newest one.
  - **Report:** name the squash-landed bars and their squash commits.
- `bit/skills/complete/evals/evals.json`: new eval 11, `"BIT-31 is merged, complete it"`. Bars 1–3 landed in one squash and bar 4 in a later single-commit squash. Expected: each bar gets `task_update` with its own squash commit before `task_complete`, and the track is filed with the later squash.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/complete/`. Only the known kebab-case failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `python3 -m json.tool bit/skills/complete/evals/evals.json > /dev/null`
- [ ] `just lint` and `just test` pass

## User verifies
This uses BIT-50.6's sandbox.
- [ ] In the dev session, create a track with two bars. On a branch `feat`, commit each bar through `bit:commit`, then run `git push origin feat`.
- [ ] Squash it the way GitHub does: `git checkout main && git merge --squash feat && git commit -m "Probe (#4)" -m "* <bar 1 subject>" -m "* <bar 2 subject>" && git push`.
- [ ] `/bit:complete` files the track without asking anything. `completed/<track>.json` and both bars' `completed/<bar>.json` have `"commit"` = `git rev-parse HEAD` (the squash).
- [ ] Whole slice: a track that landed as GitHub squashes completes on its own, with every bar pointing at the squash that carried it.

## Commit
`feat(bit): complete repoints squash-landed bars before filing`