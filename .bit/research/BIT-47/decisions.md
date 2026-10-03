# Operator decisions

These are operator decisions, not research findings, unless an entry is marked "Claude default". The 2026-09-30 set answers the two open items in topic `soundness-2` and the stub's old "Open" section. The 2026-10-01 set came from the consistency review.

# 2026-09-30

## Commit fields
- Records gain a `commit` field on both tracks and bars. It's defined in BIT-46's format verse (Verse 2), and this track fills it.
- **The track-level commit is the anchor that matters most.** It's the track's landing or merge commit. (2026-10-01: completion records it, which is now BIT-50.)
- **Bar-level hashes are best-effort,** and it's acceptable for them to break. When completion detects a squash merge, the track update also points the track's bars at the squash commit (now BIT-50).

## Commit first, then store the hash
- A bar is committed first, and its hash is stored afterwards.
- This track changes that ordering in the do skill and in bot-dev. BIT-49's sweep only removes the `.bit/` staging lines.
- HEAD isn't captured on every write.

## Squash detection (moved to BIT-50 on 2026-10-01)
- Completion compares the squash body with the bars' commit subjects, read from git through the bar hashes while they still resolve.
- Because bar hashes may break, failing to match doesn't prove the work didn't land. Completion moves on to asking for the PR.
- This replaces an earlier inferred line that leaned on bar hashes as the main source.

## Unpushed means not done (moved to BIT-50 on 2026-10-01)
- "Landed" means pushed.
- When `origin/main` exists, a commit that's only on local `main` is not done.

## Cutover
- The operator alone decides when v2 is ready and merges the branch.
- This track doesn't gate cutover.

# 2026-10-01

## The split
- This track keeps the commit side: the commit skill, the do and bot-dev reorder, the MCP git inputs, and capture on research, feedback and retro writes.
- Merge-aware completion moves to the new track BIT-50, which comes after this one. The order is BIT-45 → BIT-46 → BIT-48 → BIT-49 → BIT-47 → BIT-50.

## Claude makes the commits
- In v2, Claude makes the commits, through a dedicated, focused commit skill.
- The skill commits, then records the hash.
- do and bot-dev use it. do's "the user commits — you never run the commit yourself" text (`bit/skills/do/SKILL.md:99`) is retired.

## Permission
- The commit skill always asks the operator's permission before committing, and that includes bot-dev.
- The operator's Claude Code permissions explicitly reject the commit tool, so every commit triggers a permission prompt.
- Dispatching bot-dev does not count as permission. bot-dev stops and asks before every commit.
- bot-dev's text that describes unattended commit-and-land "without an operator sitting in front of it", on bars with no User verifies items (`bit/agents/bot-dev.md:3,12,20`), is rewritten here.

## MCP inputs
- `task_update` and `task_complete` accept `commit` and `branch`.
- The research and feedback write tools (`research_write`, `feedback_add`, and BIT-49's `retro_write`) record `commits` entries.
- `commits` is a list of `{sha, branch, at}`. The first entry is HEAD when the record is created, and a later write appends an entry when HEAD has moved.
- "HEAD isn't captured on every write" (above) applies to tracks and bars. Research, feedback and retro writes read HEAD so they can tell whether it has moved.
- Git facts come from the session dir, not the registered path (BIT-49 topic `decisions`).

## Claude defaults (the operator can veto these)
- A bar whose commit is declined stays `doing`, and nothing is recorded.
- The skill asks in its own prose too, not only through the permission prompt.
- The commit skill only commits. bot-dev keeps its push step after a permitted commit, and bit:do doesn't push.
- bit:plan's `## Commit (user)` heading becomes `## Commit`, and the skill reads either heading.
- Writing `commit` or `branch` doesn't revoke approval (`task/store.go:298-303` today revokes on any content change).
- Git facts are read where the request comes in and passed to the store.
- The CLI doesn't gain `--commit`/`--branch` flags.
- check and complete are left alone here. complete's commit suggestion is rewritten in BIT-50.
- The commit skill gets evals in the shape of `bit/skills/complete/evals/evals.json`.

## Readiness pass, 2026-10-01 (Claude defaults; the operator can veto these)
- **`bp feedback add` (CLI) captures `commits` the same way, from the current folder.** A note's commits then don't depend on the entry point. `cmd/feedback_add.go` joins Verse 4's Touches.
- **Every bar that edits a skill or agent validates it** with skill-creator's `quick_validate.py` and `claude plugin validate ./bit` (memory `bit-skill-edits-run-skill-creator`).
- **Dev runs never use `just install` on `v2`,** sandbox `XDG_DATA_HOME` and `HOME`, and test skills through `--plugin-dir` in a scratch git repo.