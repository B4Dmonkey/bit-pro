# BIT-50 research index

**Start at [decisions](decisions.md):** the operator decisions (moved from BIT-47 with the 2026-10-01 split) and the Claude defaults that settled this pass's open calls.

## Findings that change the scope
- **2026-10-02 review → [review-2026-10-02](review-2026-10-02.md):**
  - `merge-commit`'s landing formula (`rev-list --first-parent --ancestry-path h..trunk | tail -1`) **returns empty when the branch merged trunk back in**, which then silently resolves to HEAD. Replace it with "the oldest first-parent commit that has h as an ancestor".
  - The "all bars in one squash" rule misses real tracks: BIT-31 landed in #5 and #6. Make rung (b) per bar.
  - The no-git case belongs to no verse.
  - Hashless (pre-BIT-47) tracks need Verse 4, so consider the order 1, 2, 4, 3.
  - `task_landing`'s input and output are under-specified. Defaults are suggested there.
  - Missing Touches: `bot.md:54`, complete `SKILL.md:3,8`, do `:3,:100`, `serve_mcp.go:76`.
  - "Archive" should name `task_delete`.
- "The last bar commit is the track's commit" is wrong for true merges; use the first-parent rule, corrected in `review-2026-10-02` → [merge-commit](merge-commit.md).
- Single-commit GitHub squashes have no `* subject` list (subject = PR title + `(#N)`), and the same branch was squashed twice (#16/#17) → [landing-ladder](landing-ladder.md).
- A PR number maps to its squash locally via `(#N)` in the subject, so rung (c) can be verified without network → [landing-ladder](landing-ladder.md).
- A fourth state, "can't tell" (all hashes unresolvable or empty), is not "not done" → [done-classification](done-classification.md).
- `Update` can't touch a completed task, so bar repointing must precede `task_complete` → [fields-and-writes](fields-and-writes.md).
- Rebase-merges and a rebased v2 make every bar hash stale, and the ladder falls to rung (c).
- The open calls this pass raised are **settled as Claude defaults** in [decisions](decisions.md), though the notes below still call them open:
  - sign-off no longer completes (the track stays `doing` until completion);
  - the earliest duplicate squash on first-parent wins;
  - the track's `branch` is the trunk;
  - the landing logic lives in Go, behind a read-only MCP tool;
  - trunk is `origin/main`, else `main`;
  - mixed landed + unresolvable is partly done (check in).

## Topics
- [claim-audit-2026-10-02](claim-audit-2026-10-02.md): every claim in the body and topics graded CONFIRMED/WRONG/STALE with evidence, plus the build map: Complete/task_complete call sites and tests, prose lines and BIT-47/49 overlaps, and a green bar order.
- [decisions](decisions.md): operator decisions and Claude defaults. Read first.
- [review-2026-10-02](review-2026-10-02.md): independent review. Every file:line claim verified (with which anchors survive BIT-47/49 drift), real-git checks of each rung, the landing-formula bug, store write order, verse gaps, a `task_landing` shape, and hazards.
- [completion-today](completion-today.md): how the complete skill, MCP, CLI and store work now, what moves to post-landing, and who triggers it.
- [landing-ladder](landing-ladder.md): each rung's git commands verified on real history, no-fetch behaviour, gaps.
- [done-classification](done-classification.md): per-bar buckets and the track verdict.
- [merge-commit](merge-commit.md): scratch-repo proof that rung (a) handles merges. Its landing formula is wrong with back-merges; see `review-2026-10-02`.
- [fields-and-writes](fields-and-writes.md): track and bar `commit`/`branch`, BIT-47 inputs, write order. Its line cites into `.bit/tasks/BIT-46.md`/`BIT-47.md` predate the latest scope rewrites; see those tracks' Decisions instead.
- [verse-order](verse-order.md): the proposed walking-skeleton-first order, which the scope follows. `review-2026-10-02` suggests 1, 2, 4, 3.
