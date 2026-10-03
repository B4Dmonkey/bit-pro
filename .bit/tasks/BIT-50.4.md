---
id: BIT-50.4
title: A bar lands at the oldest first-parent trunk commit that contains it, and the track at the newest of those
status: todo
phase: 1
phase_label: complete after push
---
## **Verse 1**

A bar that reached trunk through a true merge is an ancestor of trunk, but its own commit isn't the anchor. The anchor is the merge that brought it in. That contradicts "landing = the bar's commit", and a track whose bars were committed out of order contradicts "the last bar's commit". Together they force the scope's rule.

## Scope
- `git/git.go`:
  - `func FirstParents(ctx context.Context, run Runner, dir, ref string) ([]string, error)`: runs `rev-list --first-parent <ref>` and returns full SHAs newest first. Empty output gives an empty slice, and an error is wrapped.
  - `func AncestryPath(ctx context.Context, run Runner, dir, sha, ref string) ([]string, error)`: runs `rev-list --ancestry-path <sha>..<ref>`, deliberately **without** `--first-parent`.
- `landing/landing.go`:
  - Every git call after trunk resolution passes the resolved trunk SHA, not the ref name, so one `Check` reads one snapshot of trunk.
  - `FirstParents(trunk)` is read once per `Check`, and an index maps each SHA to its position.
  - A bar's landing is the oldest first-parent commit in `{sha} ∪ AncestryPath(sha, trunk)`, the one with the highest index. When `sha` is itself on the chain, it lands at itself.
  - Never use `rev-list --first-parent --ancestry-path … | tail -1`. It returns nothing for a bar committed before its branch back-merged trunk, and an empty result fed to `git log -1` silently gives HEAD (research `review-2026-10-02` §2, `claim-audit-2026-10-02` WRONG 1).
  - The report's `Landing` is the newest bar landing, the one with the lowest index. It's well defined because every landing is on one chain.
  - Errors from `FirstParents` or `AncestryPath` are returned wrapped, since the trunk already resolved.
- Tests: `landing/landing_test.go` and `git/git_test.go`.

## References
- research `review-2026-10-02` §2 "Rung (a)" and `merge-commit` (the scratch-repo shapes), via `mcp__bit__research_read BIT-50`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/a bar merged through a true merge lands at the merge`
     - **Behavior:** a bar that reached trunk through a merge is anchored to the merge commit that brought it in.
     - **Setup:** `r := gittest.New(t)`. `r.Git("checkout", "-b", "feat")`, then `b1 := r.Commit("feat(bit): on a branch")`. `r.Git("checkout", "main")`, then `r.Commit("chore: trunk moves")`. `r.Git("merge", "--no-ff", "-m", "Merge branch 'feat'", "feat")`, then `m := r.Git("rev-parse", "HEAD")` and `r.Git("push", "origin", "main")`. Bar `{BIT-1.1 done b1}`.
     - **Assertions:** `Bars[0].Landing == m`, `Landing == m`, `Verdict == Done`.
     - **Boundary:** the bar's commit is off the first-parent chain.
   - [ ] Confirm fails: `Bars[0].Landing == b1`.

2. **Implement (GREEN):**
   - [ ] `FirstParents`, `AncestryPath`, the per-bar landing and the newest-landing rule.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/a bar from before a back merge lands at the merge`:
     - Setup: on `feat`, `b1`, then `git merge --no-ff -m "Merge branch 'main' into feat" main` after a trunk commit, then `b2`. On `main`, `merge --no-ff feat` gives `m`. Then one more trunk commit `m4` and a push.
     - Assertions: bars `b1` and `b2` both land at `m`, and `Landing == m`, not `m4`.
     - *Boundary:* the shape that breaks the old formula, seen on `origin/worktree-bit-39`.
   - [ ] `TestCheck/a fast forward merge lands each bar at itself`: on `feat`, commit `b1` and `b2`. Then `checkout main`, `merge --ff-only feat` and push. Landings are `b1` and `b2`, and `Landing == b2`. *Boundary:* merged bars that sit on the first-parent chain.
   - [ ] `TestCheck/the newest landing wins whatever the bar order`: `c2 := r.Commit("feat(bit): two")`, then `c1 := r.Commit("feat(bit): one")`, then push. Bars in order `{BIT-1.1 c1}` and `{BIT-1.2 c2}` give `Landing == c1`. *Boundary:* the bar order disagrees with trunk order, contradicting "the last bar's commit".
   - [ ] `TestCheck/a malformed commit never reaches git` (existing, BIT-50.2): `Check` now reads `FirstParents` on every call, so the recording fake also answers `rev-list --first-parent <its trunk sha>` with that SHA. The assertions are unchanged.
   - [ ] `TestFirstParents` (table, `fakeGit`): two lines give two SHAs in order. Empty output gives an empty slice. An error is returned and wraps the runner's error.
   - [ ] `TestAncestryPath` (table, `fakeGit`): the key is `rev-list --ancestry-path <sha>..origin/main`. The cases are as for `TestFirstParents`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): a bar lands at the oldest first-parent trunk commit that contains it`