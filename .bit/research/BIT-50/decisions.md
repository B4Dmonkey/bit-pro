# Operator decisions

These are operator decisions, not research findings, unless an entry is marked "Claude default". This track was split out of BIT-47 on 2026-10-01. The completion decisions made on 2026-09-30 under BIT-47 (see BIT-47 topic `decisions`) moved here with the split.

## 2026-10-01
- **Split from BIT-47.** BIT-47 keeps the commit side: the commit skill, the do and bot-dev reorder, the MCP git inputs, and capture on research, feedback and retro writes. This track takes merge-aware completion.
- **Order:** BIT-45 → BIT-46 → BIT-48 → BIT-49 → BIT-47 → BIT-50. This track relies on the bar hashes BIT-47 records and on BIT-49's git helper.
- **Git facts come from the session dir, not the registered path** (BIT-49 topic `decisions`).

## 2026-09-30 (moved from BIT-47)
- **The track's `commit` is its landing or merge commit,** the anchor that matters most. Completion records it.
- **Bar hashes are best-effort and may break.** When completion detects a squash merge, it also points the track's bars at the squash commit.
- **Squash detection** compares the squash body with the bars' commit subjects, read from git through the bar hashes while they still resolve. Failing to match doesn't prove the work didn't land, so completion moves on to asking for the PR.
- **Completion runs after the work has landed, not at sign-off.**
- **The ladder for the landing commit:** the bars' commits on main, then a squash commit whose body lists their subjects, then ask the operator for the PR.
- **"Landed" means pushed.** When `origin/main` exists, only it counts.
- **No hardcoded author identity. bp never fetches.** Before suggesting archive, completion tells the operator to fetch, pull or push and try again.
- **Not done:** warn that it probably should be archived and let the operator decide. Work only on a local branch counts as not done.
- **Git fields may be empty** (`acme/`).
- **Cutover:** the operator alone decides when v2 is ready and merges the branch. This track doesn't gate it.

## Claude defaults (2026-10-01; the operator can veto these)
- Sign-off no longer completes the track. bit:do tells the operator to push or merge and then run `/bit:complete`, and the track stays `doing` until then. Completion marks everything `done` and files it.
- Trunk is `origin/main`, else `main`. Other default-branch names aren't supported yet.
- A bar that reached trunk through a true merge is landed, and the track's commit is the merge commit that brought it in. For direct commits, it's the newest bar commit on the first-parent chain (topic `merge-commit`).
- A squash match needs every bar's subject in one squash, and the squash's own subject line counts too (single-commit squashes). If several squashes match, the earliest on trunk's first-parent chain wins (topic `landing-ladder`).
- Bars are repointed to the squash before the track is filed.
- Bars are classed as landed, landed via squash, pushed but not merged, local only, unresolvable, or no hash. All landed means done, mixed means partly done (check in), none landed means not done, and nothing traceable means "can't tell", which asks for the PR (topic `done-classification`).
- Asking for the PR accepts a PR number (found locally by `(#N)` on trunk) or a commit.
- Rebase-merges aren't detected, and they fall to asking for the PR.
- The track's `branch` is the trunk it landed on.
- The landing logic lives in Go, behind a read-only MCP tool that the complete skill calls.
- The complete skill stops forcing unfinished bars to `done` and stops suggesting a filing commit.
- In a folder with no git, completion files the track without a landing commit after the operator confirms.

## Readiness pass, 2026-10-01 (Claude defaults; the operator can veto these)
- **A bar whose status isn't `done` makes the track partly done at best.** The classification above reads only hashes, and it can't see unfinished work: a never-committed bar has no hash, and a migrated bar carries HEAD at migration whatever its status. The check-in names the unfinished bars, and the skill marks them `done` only when the operator says so. `task_complete` still refuses unfinished bars (`task/store.go:112-130`).
- **Migrated hashes count like any other.** A done bar of a track migrated in flight reads as landed at the migration HEAD. Accepted: it only touches tracks that straddle cutover, and BIT-49's stamping is an operator decision.
- **The landing tool is named `task_landing`.**
- **The complete skill's evals are rewritten** (`bit/skills/complete/evals/evals.json`: eval 1 expects a suggested commit, eval 2 expects bars forced to `done` without asking).
- **bit:bot's sign-off routing (`bit/agents/bot.md:41`) changes** to match completion after landing.
- **Every bar that edits a skill or agent validates it** with skill-creator's `quick_validate.py` and `claude plugin validate ./bit`.
- **Dev runs never use `just install` on `v2`,** sandbox `XDG_DATA_HOME` and `HOME`, and exercise the ladder in scratch repos with a local bare `origin`.

Evidence: BIT-50 topics `completion-today`, `landing-ladder`, `done-classification`, `merge-commit`, `fields-and-writes` and `verse-order`; BIT-47 topics `soundness` and `soundness-2`; BIT-45 topic `history-anchors` (read its banner first).