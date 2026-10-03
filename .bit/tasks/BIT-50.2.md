---
id: BIT-50.2
title: A bar that isn't on origin/main makes the track not done
status: todo
approved: true
phase: 1
phase_label: complete after push
---
## **Verse 1**

A bar that's committed but never pushed can't pass with the hardcoded `done`. That forces a real ancestry check of each bar's commit against `origin/main`. The skill also needs the unfinished bars, because Verse 1 refuses to complete a track that has any.

## Scope
- `git/git.go`, beside `ReadHead`, each function taking `Runner` as its second parameter:
  - `func ResolveCommit(ctx context.Context, run Runner, dir, rev string) (string, bool)`: runs `rev-parse --verify -q <rev>^{commit}` and returns the full SHA. It returns `("", false)` on an error or empty output.
  - `func IsAncestor(ctx context.Context, run Runner, dir, sha, ref string) bool`: runs `merge-base <sha> <ref>` and compares the output to `sha`, returning `false` on error. It reads output rather than `merge-base --is-ancestor`'s exit status, because `Runner` folds the exit code into `err` and can't tell 1 (not an ancestor) from 128 (bad object).
- `landing/landing.go`:
  - The trunk is `ResolveCommit(..., "refs/remotes/origin/main")`, reported as `"origin/main"`. The fallback is BIT-50.3's.
  - Each bar's commit must match `^[0-9a-f]{40}$` (full SHAs, per the scope) before it reaches git, so a stored value can never be read as a git option. A malformed commit is `NotLanded` without calling git.
  - Otherwise, call `ResolveCommit`, then `IsAncestor(sha, trunk)`. The bar is `Landed` with `Landing = sha`, or `NotLanded Class = "not_landed"` with `Landing ""`. Verse 2 replaces `not_landed` with finer classes.
  - `NotDone Verdict = "not_done"`. The verdict is `Done` when every bar is `Landed`. Otherwise it's `NotDone`, with the report's `Landing ""`.
  - The report's `Landing` is still the last landed bar's commit in bar order. BIT-50.4 contradicts that.
  - `Report` gains `Unfinished []string \`json:"unfinished"\``: the IDs of bars whose status isn't `done`, in bar order and never nil. It doesn't affect the verdict yet (that's Verse 2).
- Tests:
  - new `landing/landing_test.go`. `TestCheck` holds subtests over `gittest` repos that call `Check(ctx, git.ExecRunner, Query{Dir: r.Dir, Bars: ...})`.
  - `git/git_test.go`, already in shape since BIT-49.9, so there's no step 0.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/a bar committed but not pushed`
     - **Behavior:** work that only exists on local `main` isn't done, because "landed" means pushed.
     - **Setup:** `r := gittest.New(t)`. `a := r.Commit("feat(bit): one")`, then `r.Git("push", "origin", "main")`, then `b := r.Commit("feat(bit): two")`, which isn't pushed. Bars `{BIT-1.1 done a}` and `{BIT-1.2 done b}`.
     - **Assertions:** `Verdict == NotDone` and `Landing == ""`. `Bars[0]` is `{Class: Landed, Landing: a}`. `Bars[1]` is `{Class: NotLanded, Landing: ""}`.
     - **Boundary:** one bar present on local `main` but absent from `origin/main`.
   - [ ] Confirm fails: `Verdict == Done` (hardcoded).

2. **Implement (GREEN):**
   - [ ] `ResolveCommit`, `IsAncestor`, the commit pattern check, classification and the verdict rule.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/every bar pushed`: the same as the RED case with `b` pushed too, giving `Done` with `Landing == b`. *Boundary:* the happy path over real git.
   - [ ] `TestCheck/a malformed commit never reaches git`: a recording fake runner that knows only `rev-parse --verify -q refs/remotes/origin/main^{commit}`. Bars `{"--output=/tmp/x"}` and `{"abc123"}` are both `NotLanded`, and no recorded call contains either value. *Boundary:* an untrusted stored value that's neither a full SHA nor safe as a git argument.
   - [ ] `TestCheck/an unfinished bar is listed`: bars `{BIT-1.1 done a}` (pushed) and `{BIT-1.2 todo ""}` give `Unfinished == []string{"BIT-1.2"}`. *Boundary:* a status other than `done`.
   - [ ] `TestResolveCommit` (table, `fakeGit`): output `"<sha>\n"` gives `(sha, true)`. An error gives `("", false)`.
   - [ ] `TestIsAncestor` (table, `fakeGit`): `merge-base <sha> origin/main` returning `sha` gives true. Another SHA gives false. An error gives false.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): a bar not on origin/main makes the track not done`