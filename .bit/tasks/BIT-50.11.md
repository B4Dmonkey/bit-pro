---
id: BIT-50.11
title: /bit:complete checks in on partly done, not done, no git and shallow, and the operator decides
status: todo
phase: 2
phase_label: operator decides partial landings
---
## **Verse 2**

`task_landing` now tells the cases apart. This bar replaces Verse 1's two flat stops in the complete skill with the scope's check-ins. A partly done track gets a check-in, a not done track gets the archive-or-keep choice, a folder with no git gets confirm-and-file, a shallow clone gets "unshallow and retry", and can't tell gets "fetch and retry". The PR question, on both can't tell and not done, is BIT-50.14's. This is skill prose, so the evals carry the behaviour.

## Scope
- `bit/skills/complete/SKILL.md`, **For each track** (re-read it first, since BIT-50.6 wrote it). After `task_landing`, act on the first case that applies:
  - **`shallow` is true:** "this clone is shallow, so git can't see where the work landed: run `git fetch --unshallow`, then `/bit:complete <ID>` again". Stop.
  - **`done`:** as in Verse 1.
  - **`partly`:**
    - Name the landed bars, the rest with their class (pushed but not merged, local only, unresolvable), and the unfinished bars.
    - Say the unlanded ones need fetch, pull or push (no "push" when `trunk` is `main`).
    - Ask whether to complete anyway.
    - On a yes: `task_update` each unfinished bar to `done` (only with this OK), mark the track `done`, then `task_complete {id, commit: <landing>, branch}`.
    - On a no: change nothing.
  - **`not_done`:**
    - "<ID> hasn't landed: fetch or pull, and push if the commits are only local, then run `/bit:complete <ID>` again." Leave out "and push if the commits are only local" when `trunk` is `main` (no `origin/main`).
    - Then warn that if the work was abandoned, the track probably should be archived, and let the operator choose. Keep it open as a reminder, which changes nothing, or archive it with `mcp__bit__task_delete {id, force: true}` when any bar is unfinished (plain `{id}` otherwise), only after they confirm.
    - A missing PR is a signal, not a rule.
    - BIT-50.14 adds "or tell me the PR number or a commit that landed it" to this check-in, for work that reached trunk by a rebase-merge or an unlisted squash.
  - **`no_git`:** "this folder isn't in git, so there's no landing commit to record". Ask to confirm. On a yes, mark unfinished bars `done` (only with that OK) and the track `done`, then run `task_complete {id}` with no `commit` or `branch`.
  - **`cant_tell`:** "git can't place this track's work: none of its bar commits resolve. Fetch and retry." Stop, and file nothing. BIT-50.14 adds the PR-or-commit question here.
  - Remove Verse 1's "finish or mark these bars done first" stop. Unfinished bars are now part of the `partly` check-in. They're never marked `done` without the operator's OK, and `task_complete` still refuses them.
  - **Report:** add the outcome per track: filed, kept open, archived, or waiting on fetch.
- `bit/skills/complete/evals/evals.json`:
  - eval 2 (a todo bar, the rest landed): the check-in names the todo bar and asks. It's flipped and filed only after a yes.
  - new eval 5: `"complete BIT-9"`, with one bar on `origin/main` and one only on `origin/feat`. Expected: the check-in names the pushed-but-unmerged bar, and nothing is filed without a yes.
  - new eval 6: `"close out BIT-10"`, in a repo with an `origin` remote (so `trunk` is `origin/main`) and every bar only on a local branch. Expected: "not landed: push and retry", keep-or-archive offered, and `task_delete` runs only after the operator picks archive.
  - new eval 7: `"complete ACME-3"` in a folder with no git. Expected: it says there's no landing commit, asks, and then calls `task_complete` without `commit`.
  - new eval 8: `"complete BIT-11"` in a shallow clone. Expected: `git fetch --unshallow` and retry, with nothing filed.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/complete/`. Only the known kebab-case failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `python3 -m json.tool bit/skills/complete/evals/evals.json > /dev/null`
- [ ] `just lint` and `just test` pass

## User verifies
This uses BIT-50.6's sandbox (`$SB/proj` with `$SB/origin.git`, and the dev `claude` session).
- [ ] Track A has one bar, committed through `bit:commit` on a branch `feat` and pushed with `git push origin feat`. `/bit:complete` says it hasn't landed and offers keep or archive. Choose keep, and `bp task list` still shows it.
- [ ] Track B has bar 1 committed on `main` and pushed, and bar 2 on `feat`, pushed. `/bit:complete` names bar 2 as pushed but not merged and asks. Answer yes. `completed/<B>.json` has `"commit"` = bar 1's SHA.
- [ ] Run `mkdir $SB/acme && cd $SB/acme && printf 'acme\n' | bp add .`, then start the dev session there (as in BIT-50.6, with `cd $SB/acme`) and create a track with one bar marked done. `/bit:complete` there says there's no landing commit and asks. Answer yes. `completed/ACME-1.json` has `"commit": ""`.
- [ ] Whole slice: whenever a track isn't cleanly landed, the operator is told why and decides, and nothing is filed or archived without their say.

## Commit
`feat(bit): complete checks in when a track isn't cleanly landed`