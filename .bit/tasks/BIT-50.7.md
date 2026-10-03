---
id: BIT-50.7
title: Each bar is classed pushed, local, unresolvable or no hash, and the verdict says partly, not done or can't tell
status: todo
phase: 2
phase_label: operator decides partial landings
---
## **Verse 2**

`not_landed` hides the distinction the operator acts on. A track with one bar landed and one pushed to a branch should check in, not just say "not done", and a track whose hashes don't resolve should say "can't tell", not "not done". A partly landed track contradicts the binary verdict, and that forces the per-bar classes and the scope's four-way verdict.

## Scope
- `git/git.go`: `func RemoteContains(ctx context.Context, run Runner, dir, sha string) bool`, which runs `branch -r --contains <sha>`. Non-empty output means true. An error or empty output means false.
- `landing/landing.go`:
  - Replace `NotLanded` with `Pushed Class = "pushed"`, `Local = "local"`, `Unresolvable = "unresolvable"` and `NoHash = "no_hash"`.
  - Classification in order:
    - an empty commit is `NoHash`;
    - a commit that isn't 40 hex characters, or doesn't resolve, is `Unresolvable`;
    - an ancestor of trunk is `Landed`;
    - a commit `RemoteContains` finds is `Pushed`;
    - anything else is `Local`.
  - `Partly Verdict = "partly"` and `CantTell = "cant_tell"`. The verdict, with landed meaning class `Landed`:
    - no bar landed and none `Pushed` or `Local`: `CantTell`;
    - at least one landed and every bar with a hash landed: `Done` (`NoHash` bars don't block);
    - some landed, and the rest with a hash are `Pushed`, `Local` or `Unresolvable`: `Partly`;
    - none landed, and at least one `Pushed` or `Local`: `NotDone`.
  - `Landing` is the newest landing among landed bars for `Done` and `Partly` (scope: proceeding on a partly done track uses it), and `""` otherwise.
- Existing rows that assert `NotLanded` change: `TestCheck/a bar committed but not pushed` expects `Local`, and `TestCheck/a malformed commit never reaches git` expects `Unresolvable`.
- Tests: `landing/landing_test.go` and `git/git_test.go`.

## References
- research `done-classification`, with its stale "operator decision" line settled as partly done (`claim-audit-2026-10-02`), via `mcp__bit__research_read BIT-50`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/some bars landed and one only pushed to a branch`
     - **Behavior:** a track whose work is half on trunk reads as partly done, so the operator is asked rather than told it isn't done.
     - **Setup:** `a := r.Commit("feat(bit): one")`, then push `main`. `r.Git("checkout", "-b", "feat")`, `b := r.Commit("feat(bit): two")` and `r.Git("push", "origin", "feat")`. Bars `{BIT-1.1 done a}` and `{BIT-1.2 done b}`.
     - **Assertions:** `Verdict == Partly`, `Landing == a`, and the classes are `[Landed, Pushed]`.
     - **Boundary:** one landed bar beside one pushed but unmerged bar.
   - [ ] Confirm fails: `Verdict == NotDone`.

2. **Implement (GREEN):**
   - [ ] `RemoteContains`, the classes and the four-way verdict.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/nothing landed and one bar local`: `b` on an unpushed `feat` gives `NotDone`, class `Local`. *Boundary:* not on any remote.
   - [ ] `TestCheck/nothing landed and one bar pushed to a branch`: `b` pushed to `origin/feat` gives `NotDone`, class `Pushed`.
   - [ ] `TestCheck/every hash unresolvable`: a bar with commit `0123456789abcdef0123456789abcdef01234567` gives `CantTell`, class `Unresolvable`. *Boundary:* well formed but unknown, which is what a rebase leaves behind once the originals are gone (after gc or in a fresh clone).
   - [ ] `TestCheck/no bar has a hash`: bars with `Commit ""` give `CantTell`, class `NoHash`.
   - [ ] `TestCheck/a bar with no hash beside landed bars`: `{a}` (landed) and `{""}` give `Done`, `Landing == a`. *Boundary:* `NoHash` doesn't block, for example a bar check marked done without committing it.
   - [ ] `TestRemoteContains` (table, `fakeGit`): `"  origin/feat"` gives true, `""` gives false, and an error gives false.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): bars are classed and a track can be partly done or untraceable`