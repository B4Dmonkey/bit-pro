---
id: BIT-50.14
title: /bit:complete asks for the PR or a commit when it can't tell or the work isn't on trunk, and repoints those bars before filing
status: todo
phase: 3
phase_label: PR or commit answer
---
## **Verse 3**

`task_landing` can now check an answer. This bar makes the complete skill ask for one when it can't place the work, offer one when the track reads not done (a rebase-merge or an unlisted squash leaves the bars pushed or local), and write the repointed bars before filing, because `task_update` can't reach a filed task. This is skill prose, so the evals carry the behaviour.

## Scope
- `bit/skills/complete/SKILL.md` (re-read first):
  - **`cant_tell`:** replace BIT-50.11's stop. Say git can't place the work, then ask: "Fetch and retry, or tell me the PR number or a commit that landed it." On an answer, call `task_landing {id, pr}` or `{id, commit}` again and act on the new verdict.
    - A "not on trunk" error: relay it ("fetch and retry"). File nothing.
    - An ambiguous-PR error: show the listed SHAs, ask which one, and call again with `commit`.
    - "Not a commit": ask again.
  - **`not_done`:** keep BIT-50.11's "fetch, pull or push and retry" and the keep-or-archive choice, and add "or tell me the PR number or a commit that landed it", for work that reached trunk by a rebase-merge or a squash that doesn't list the bars. An answer is handled exactly as for `cant_tell`.
  - **Before filing, in every filing path:** for each bar with `repoint: true`, run `mcp__bit__task_update {id: <bar>, commit: <its landing>}` (BIT-47.2's `commit` input).
  - Write the full order once: unfinished bars to `done` (only with the operator's OK), repoint bars, the track to `done`, then `task_complete {id, commit, branch}`.
  - **Report:** name the repointed bars.
- `bit/skills/complete/evals/evals.json`:
  - new eval 9, `"complete BIT-12"`, whose bars have no hashes. Expected: it asks for a PR or a commit. The operator answers `#5`. `task_landing` is called with `pr: 5`, each bar gets `task_update` with `commit` before `task_complete`, and the track is filed with that commit.
  - new eval 10, `"complete BIT-13"`, whose bars are pushed only to `origin/feat`, and `feat` was rebase-merged onto `origin/main`. Expected: it says the track hasn't landed, offers fetch-and-retry, keep-or-archive, and a PR number or commit. The operator answers with the rebased tip's SHA. `task_landing` is called with `commit`, each bar gets `task_update` with `commit` before `task_complete`, and the track is filed with that commit. Nothing is archived.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/complete/`. Only the known kebab-case failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `python3 -m json.tool bit/skills/complete/evals/evals.json > /dev/null`
- [ ] `just lint` and `just test` pass

## User verifies
This uses BIT-50.6's sandbox.
- [ ] In the dev session, create a track with one bar and mark the bar `done` with no commit recorded. Then in `$SB/proj`, run `git commit --allow-empty -m "Probe (#3)" && git push`.
- [ ] `/bit:complete` on it asks for a PR or a commit. Answer `99`: it says the PR isn't on `origin/main`, to fetch and retry, and nothing is filed.
- [ ] Run it again and answer `3`. It files the track. `completed/<track>.json` and the bar's `completed/<bar>.json` both have `"commit"` = `git rev-parse HEAD`.
- [ ] Rebase-merge: create a second track with one bar, commit it through `bit:commit` on `feat` and run `git push origin feat`. Then `git checkout main && git cherry-pick feat && git push`. `/bit:complete` says it hasn't landed and offers a PR or a commit as well as keep or archive. Answer `git rev-parse main`. It files the track, and the bar's `completed/<bar>.json` has `"commit"` = that SHA.
- [ ] Whole slice: when git can't place a track's work, or the work reached trunk under new hashes, a PR number or commit from the operator completes it, checked on trunk.

## Commit
`feat(bit): complete asks for the PR or commit when it can't tell or the work isn't on trunk`