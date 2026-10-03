# True merge commits (incl. v2 itself)

**Checked:** does the ladder handle bars reaching main through a real merge (e.g. merging the `v2` branch)?

- Today main has **no** merge commits (`git log --merges main` empty) and `v2` == `main` == `origin/main` at `6a1d345`, so the v2 merge hasn't happened and there's no real example; tested in a scratch repo (main: m1, m2, `Merge branch feat` (--no-ff), m3; feat: b1, b2).
- `git merge-base --is-ancestor b2 main` → 0. So rung (a) already says "landed" for merged bars; no new rung needed.
- But "the last bar commit is the track's commit" is wrong for a merge: b2 is not on the first-parent chain (`git rev-list --first-parent main | grep b2` → 0). The landing commit is found with `git rev-list --first-parent --ancestry-path <b2>..main | tail -1` → `Merge branch feat` (verified). For a direct commit that formula returns the *next* commit, so first check first-parent membership and use the bar itself when it's on the chain.
- v2 risk: if the operator rebases v2 onto main before merging (or fast-forwards after a rebase), every bar hash recorded on v2 goes stale → unresolvable → rung (c). A plain `git merge v2` (or fast-forward without rebase) keeps hashes valid.

**Verdict:** rung (a) handles merges for "landed"; the landing-commit pick needs the first-parent/ancestry-path rule, not "last bar".
