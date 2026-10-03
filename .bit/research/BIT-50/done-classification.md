# Deciding done / partly done / not done from bar commits

**Checked:** whether bar hashes alone can separate done, partly done and not done.

Per bar with a recorded `commit`, classify (commands in `landing-ladder`):
1. `cat-file -e` fails (exit 128) → **unresolvable** (gc'd, rebased away, or recorded in another clone). Not evidence of not-done.
2. `merge-base --is-ancestor <h> <trunk>` = 0 → **landed**.
3. Its subject is found in a trunk squash that matches the track (rung b) → **landed via squash**.
4. `git branch -r --contains <h>` non-empty → **pushed, not merged** (e.g. `cebf42d` → `remotes/origin/worktree-agent-a6848e37a1646d392`).
5. Otherwise → **local only** (counts as not done, per decisions).
Bars with an empty `commit` → **no evidence**.

Track verdict:
- all bars-with-hashes landed → **done** (record landing commit).
- some landed, some pushed/local → **partly done** → check in with operator.
- none landed, ≥1 local/pushed → **not done** → tell operator to fetch/pull/push and retry; then suggest archive or keep as reminder. Pushed-not-merged is the "probably an open PR" signal.
- every hash unresolvable or empty → **can't tell** → rung (c), ask for the PR. Do not report this as not done.
- mixed landed + unresolvable → treat unresolvable as unknown; likely "done" if the landed ones include the newest bar, but **operator decision**.

Caveat: all answers are as of the last fetch (remote-tracking refs are stale until fetched). `branch -r --contains` also only knows remote branches that still exist locally (`git fetch --prune` would remove deleted ones).

**Verdict:** decidable from git without network, with an explicit "unknown" bucket the stub's three states don't name.
