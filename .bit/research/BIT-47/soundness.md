# BIT-47 soundness pass (stub, v2 @ 6a1d345)

**Checked:** whether the operator decisions are feasible given the code and bit-pro's real `git log main`.

## What the history actually looks like
- `git log main --format=%an | sort | uniq -c`: **304 `josiah` <lowgen0@gmail.com>** (local commits straight to main) vs **16 `Josiah Wilkins` <5740397+B4Dmonkey@users.noreply.github.com>, committer `GitHub`** (squash merges #1-#17).
- Squash subjects are the branch name, not the track: `54abfeb Worktree agent a6848e37a1646d392 (#17)`, `84886bb Worktree bit 39 (#15)`, `82dd6f7 Worktree zesty strolling raccoon (#8)`.
- Squash bodies do list the branch commit subjects as `* <subject>` bullets (`git show -s 54abfeb`, `84886bb`), so body matching is possible.
- Recent tracks landed **without any PR**: BIT-44 is commits `26f0435`..`8decf35` directly on main; so were the research tools (`af07620`..`a7d9ea7`).
- Bar commit messages live only in body prose: `## Commit (user)` then a backticked line (`.bit/completed/BIT-43.1.md:65-66`). Actual commits drift from them (`291b508 ufeat(version)…`, `0f609d4 the next commit since copy is borken…`).
- bp has no git exec (`grep exec.Command`: claude, launchctl only). All git capture is new code.

## Issues
- **blocking — "no merge commit → ask for PR → no PR = abandoned → archive" misfires on direct-to-main work.** Most bit-pro tracks ship as one commit per bar on main with no squash and no PR. The rule would archive shipped work (BIT-44 today). The heuristic needs a second branch: bars' commits found individually on main = landed.
- **should-fix — "authored by the operator" has no identity to match.** Squash author is the GitHub noreply address, local author is `lowgen0@gmail.com` (`git config user.email`); neither is the account email. Needs a configured identity set, or drop the author filter and rely on message matching.
- **should-fix — bar commit messages aren't a field.** BIT-46 Decision: bodies are never split into fields. Matching needs either a `commit_message` JSON field (a BIT-46 format change) or parsing `## Commit` out of the body. Matching must be fuzzy/subset (typos, edits at commit time).
- **should-fix — completion timing.** Today `bit_complete` (`bit/skills/complete/SKILL.md`) and `task_complete` run at sign-off, before the PR merges. "Complete only after merge" means a new waiting state or a later re-run; who triggers it, and from where, is unspecified. `bit/skills/complete` and `cmd/serve_mcp.go` task_complete are the real Touches.
- **should-fix — which main.** Squash merges happen on GitHub; local `main` is stale until pulled, and sessions often run in a worktree on another branch. Decide `main` vs `origin/main` and whether bp fetches. Git facts must come from the session dir (worktree), not the registered path (BIT-45 `project-resolution`).
- **nit — v2 itself.** BIT-45/46/47 land on `v2` and reach main via the cutover merge (likely a true merge, not squash); the flow should handle a merge commit.
- **nit — dependency.** "Capture on every write" needs the JSON git fields from BIT-46 Verse 2 and the git helper that BIT-49 Verse 2 (migrate) adds (it was the original BIT-46 Verse 4 before the split); nothing else outside those tracks is required.

## Sound
Empty git fields for non-git projects (`acme/`) are consistent with BIT-46's Decision.