---
id: BIT-50.13
title: A PR number finds its (#N) commit on trunk, and two matches are refused with both SHAs
status: todo
phase: 3
phase_label: PR or commit answer
---
## **Verse 3**

The operator often knows the PR number rather than a SHA. A PR answer has to find the commit whose subject ends in ` (#N)` on trunk's first-parent chain, locally and without `gh`. A query carrying only a PR number forces the lookup.

## Scope
- `git/git.go`: `func PRCommits(ctx context.Context, run Runner, dir, ref string, n int) ([]string, error)`. It runs `log --first-parent --format=%H%x00%s <ref>`, splits on lines and then on NUL, and keeps the SHAs whose subject ends with `fmt.Sprintf(" (#%d)", n)`. The `)` delimits, so `(#1)` never matches `(#17)`, and a revert subject ending `(#5)"` doesn't match either.
- `landing/landing.go`:
  - `Query` gains `PR int`.
  - `type AmbiguousPRError struct{ PR int; SHAs []string }`. Its `Error()` lists each SHA on its own line, so the skill can show them and ask which one.
  - With `PR > 0`: no match gives `ErrNotOnTrunk` ("PR #N isn't on <trunk>: fetch and retry"). One match is used exactly as a commit answer (BIT-50.12). More than one gives `*AmbiguousPRError`.
- `cmd/serve_mcp.go`: `taskLandingInput` gains `PR int \`json:"pr,omitempty"\``.
- Test files: `landing/landing_test.go`, `git/git_test.go` and `cmd/serve_mcp_test.go`.

## References
- research `landing-ladder` ("(#N) lookup") and `review-2026-10-02` §2, via `mcp__bit__research_read BIT-50`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/a pr number finds its squash on trunk`
     - **Behavior:** a PR number is enough to complete a track that git couldn't place.
     - **Setup:** on `feat`, commit `f := r.Commit("feat(tui): overlay")`. `checkout main`, then `merge --squash feat` and `commit -m "Worktree bit 31 (#5)"`, so `s` = HEAD. Push `main`. Bar `{BIT-1.1 done ""}` with `Query.PR = 5`.
     - **Assertions:** `Verdict == Done`, `Landing == s`, `Bars[0].Repoint == true`.
     - **Boundary:** a GitHub-style squash subject.
   - [ ] Confirm fails: `Query` has no `PR` field (compile error).

2. **Implement (GREEN):**
   - [ ] `PRCommits`, the PR branch and `AmbiguousPRError`.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/a pr that isn't on trunk is refused`: `PR = 6` gives `ErrNotOnTrunk`.
   - [ ] `TestCheck/two commits ending in the same pr number are refused`: two trunk commits, `"x (#7)"` and `"y (#7)"`, give `errors.As` to `*AmbiguousPRError` with both SHAs. *Boundary:* more than one match, so the skill asks which.
   - [ ] `TestCheck/pr 1 doesn't match pr 17`: only `"z (#17)"` exists, so `PR = 1` gives `ErrNotOnTrunk`. *Boundary:* a prefix of the number.
   - [ ] `TestPRCommits` (table, `fakeGit`): output `"<a>\x00x (#5)\n<b>\x00Revert \"x (#5)\"\n<c>\x00y (#15)"` with `n=5` gives `[a]`. An error is returned wrapped.
   - [ ] `TestTaskLandingHandler/an ambiguous pr is a tool error listing both commits`: `IsError`, with the text containing both SHAs.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): task_landing finds a PR's commit on trunk by its (#N) subject`