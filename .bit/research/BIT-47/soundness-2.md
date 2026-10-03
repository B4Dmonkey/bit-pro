# BIT-47 soundness pass 2 (stub, v2 @ 6a1d345)

**Checked:** the revised operator decisions for feasibility, and against BIT-46 and v2-sketch.md.

## Confirmed
- `git log main`: 320 commits. 304 have committer `josiah`, 16 have committer `GitHub` (squash merges, subjects `(#N)`), and there are 0 merge commits. So "direct commits are the common case" holds.
- `bit/skills/complete/SKILL.md` runs at sign-off and files into `.bit/completed/` through `task_complete`. That is what this track changes.
- Bar commit messages exist only as body prose under `## Commit (user)` (`bit/skills/plan/SKILL.md:361,398`). The Open item is real. BIT-46 "bodies never split into fields" means body parsing, unless BIT-46's JSON gains a field.
- Empty git fields and the git helper that BIT-49 owns (it was BIT-46's before the 2026-09-30 split) are consistent between the tracks.

## Should-fix
- **"bp never fetches" plus "prefer origin/main" gives false "not done" results.** A squash merged on GitHub isn't in `origin/main` until the operator fetches, and bp won't fetch. Then the ladder falls through to "not done → warn it should be archived" for work that did land. The not-found message should say to fetch or pull and retry before suggesting archive. Symmetric case: direct commits on local `main` that aren't pushed are invisible when `origin/main` exists. Decide whether "landed" means pushed.
- **"Capture real values on every write" can't record a bar's own commit.** bit:do sets a bar `done` before the user commits it, so HEAD at that write is the parent. Only completion (or a later write) can see the landing commit. That's fine if on-write capture is meant as the "before" anchor. Say so when scoping.
- **v2-sketch contradicts this stub.** Its "Every record carries…" bullet says "Tracks are done on their own branch and squash-merged, so branch commits don't survive." Its "History anchors…" bullet says completion looks up "the merge (squash) commit". Both predate the direct-commit evidence and this stub's ordered ladder. The later sketch bullet ("Completion asks…") matches the stub. (Fixed in v2-sketch since.)

## Nit
- v2 itself reaches main at cutover through a merge of `v2` (likely a true merge commit, of which main has none today). Completing BIT-45/46/47 goes through this flow and needs a merge-commit case, or they're completed by hand.

Nothing here contradicts BIT-46's Decisions.