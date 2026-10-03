# BIT-47 research index

**Start at [decisions](decisions.md):** the operator decisions of 2026-09-30 and 2026-10-01 (the split: merge-aware completion moved to BIT-50; Claude commits through a dedicated commit skill that always asks permission, bot-dev included; the MCP git inputs; `commits` lists on research, feedback and retro), plus the Claude defaults that settled this pass's open calls.

This topic set also serves BIT-50 (merge-aware completion). Earlier history-anchor research lives in BIT-45 topic `history-anchors` (read its superseded banner first).

## Independent review (2026-10-02): findings that change the scope
→ [review-2026-10-02](review-2026-10-02.md)
- All file:line refs are correct today. Most will be stale after BIT-46/49, so anchor by section names (table in the topic).
- The permission settings **ask**, they don't reject: `~/.claude/settings.json:15-21` (`git commit`, `git add`, `git push`). Headless/`--bg` bot-dev runs can't answer the prompt (unknown).
- Missing Touches: do `:73` (User-verifies path), do `:106` "commit suggested" (shared with BIT-50), and do `:75` "user reads the diff when they commit".
- Undecided: nothing-to-commit (likely in v2 once `.bit/` leaves the repo), pre-commit hook failures (this repo's hooks run fmt/lint/test), no-git/out-of-repo bars, unwind after a commit, detached HEAD, the exact "HEAD moved" comparison. Each has a suggested default in the topic.
- `task_complete` commit/branch has no consumer here. Suggest moving it to BIT-50. Verse 3 isn't a prerequisite for Verse 1.
- Hazard: BIT-46's research rewrite must preserve the `commits` sidecar, or the append can't work. Verse 4's `task/feedback.go` pointer is stale after BIT-49.

## Code pass (2026-10-01): findings that matter for scope
- bit:do's Verified good close-out (`bit/skills/do/SKILL.md:89-100`) marks done (`:91`), rolls up (`:92-98`), then suggests the commit (`:99`); the reorder rewrites `:3,:17,:75,:77,:79,:99,:100,:129`. No decline path exists for a refused commit (**settled, Claude default:** the bar stays `doing`). → [commit-skill](commit-skill.md)
- bot-dev's commit/push behaviour is `bit/agents/bot-dev.md:3,12-38`. Who pushes was open (**settled, Claude default:** the commit skill only commits; bot-dev keeps its push step after a permitted commit; bit:do doesn't push). → [commit-skill](commit-skill.md)
- Contradictions elsewhere: plan `:304` "Claude never commits" and `## Commit (user)` headings `:361,:398`; complete `:30` "the user runs the commit"; check `:34,:38` matches by message, marks done without a hash. (**Settled:** plan is rewritten here and the heading becomes `## Commit`; complete is rewritten in BIT-50; check is left alone.) → [commit-skill](commit-skill.md)
- `task.Store.Update` revokes approval on content changes (`task/store.go:298-303`); commit/branch must stay out of that. → [mcp-git-inputs](mcp-git-inputs.md)
- The session dir (`CLAUDE_PROJECT_DIR`) is known in the MCP handler but lost inside `Store` (`bitdir.ForRoot` strips worktrees), so git capture has to read HEAD at the entry point, not from the store root. Research/feedback files have no metadata today; `commits` needs BIT-46's sidecar. → [capture](capture.md)
- CLI parity for `--commit/--branch` was open (**settled, Claude default:** no CLI flags). → [mcp-git-inputs](mcp-git-inputs.md)
- New skills auto-register from `bit/skills/<name>/`; evals exist only for analyze, complete, feedback. → [commit-skill](commit-skill.md)

## Earlier passes
- Soundness pass 2: fetch/origin gap, on-write capture timing (since decided: commit first), v2-sketch contradiction (since fixed). → [soundness-2](soundness-2.md)
- First pass: most work lands direct to main; squash authored by GitHub noreply; bar commit messages are body prose. → [soundness](soundness.md)

## Topics
- [decisions](decisions.md): operator decisions and Claude defaults. Read first.
- [claim-audit-2026-10-02](claim-audit-2026-10-02.md): verdict on every claim (64 confirmed, 2 wrong, 8 stale), call sites/tests per git field, line-overlap table with BIT-49/50, a green bar order, and planning stallers.
- [review-2026-10-02](review-2026-10-02.md): independent re-check of every line ref (robust vs stale), missing "user commits" text, bit:commit edge cases with suggested defaults, the capture comparison rule, verse order and the task_complete move.
- [commit-skill](commit-skill.md): skill packaging and evals; exact do/bot-dev close-out lines and what changes; other skills that assume the user commits.
- [mcp-git-inputs](mcp-git-inputs.md): input structs for task_update/task_complete/feedback_add/research_write, Patch and the approval trap, CLI equivalents.
- [capture](capture.md): where write-time HEAD/branch capture sits on research and feedback writes; session dir vs store dir.
- [verse-order](verse-order.md): dependency-driven walking-skeleton order (the scope follows it).
- [soundness-2](soundness-2.md): fetch/origin gap, on-write capture timing, merge-commit case for v2.
- [soundness](soundness.md): real `git log main` evidence and feasibility of each operator decision.
