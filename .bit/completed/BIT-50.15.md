---
id: BIT-50.15
title: A bar whose subject a squash on trunk lists lands at the earliest such squash after it
status: done
approved: true
phase: 4
phase_label: squash-landed tracks
---
## **Verse 4**

A GitHub squash leaves the bar's own commit off trunk, so a squash-landed bar reads as `pushed` or `local`. The squash's body lists each branch commit as a `* subject` line, and a single-commit squash's subject is the commit subject plus ` (#N)`. A squash-landed track that comes back `not_done` contradicts the current rule, and that forces the per-bar squash rung.

## Scope
- `git/git.go`:
  - `type Squash struct{ SHA string; Time int64; Subject string; Listed []string }`:
    - `Subject` has the trailing ` (#N)` stripped;
    - `Listed` holds the text after `* ` on each body line that starts with exactly `* `.
  - `func Squashes(ctx context.Context, run Runner, dir, ref string) ([]Squash, error)`:
    - Runs `log --first-parent --format=%H%x00%ct%x00%s%x00%b%x1e <ref>`, splits records on `\x1e` (trimming each) and fields on NUL.
    - Keeps only commits whose subject ends ` (#N)`, newest first. A direct commit with a `* ` bullet in its body is never a squash.
  - `func Subject(ctx context.Context, run Runner, dir, sha string) (string, int64, bool)`: runs `log -1 --format=%s%x00%ct <sha>`.
- `landing/landing.go`:
  - `Squash Class = "squash"`.
  - Read `Squashes(trunk)` once, but only when some bar is `Pushed` or `Local`.
  - For each such bar, get its `Subject`. The match is the **oldest** squash with `Time >= the bar's time` whose `Listed` contains the subject or whose stripped `Subject` equals it.
  - A match sets `Class = Squash`, `Landing = squash SHA` and `Repoint = true`, and the bar counts as landed.
  - `Unresolvable` bars aren't matched, because they have no subject to read.
  - This runs before the operator's answer is applied (BIT-50.12) and before the verdict, so an answer repoints only bars the rung left unplaced. The track's `Landing` is still the newest landing.
- `git/gittest/gittest.go`: every `Git` call sets `GIT_AUTHOR_DATE` and `GIT_COMMITTER_DATE` to `"<1759406400+r.n> +0000"`, git's internal date format, and increments `r.n`. Commit times then strictly increase, so the time rule is deterministic.
- Test files: `landing/landing_test.go` and `git/git_test.go`.

## References
- research `landing-ladder` (squash body shape, from `git show -s 54abfeb`) and `review-2026-10-02` §2 "Rung (b)" and §4 "Squash rule", via `mcp__bit__research_read BIT-50`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/a bar listed in a multi commit squash lands at the squash`
     - **Behavior:** a track squash-merged on GitHub completes on its own.
     - **Setup:**
       - On `feat`: `b1 := r.Commit("feat(tui): play prompt")`, `b2 := r.Commit("feat(tui): queue view")`, then `push origin feat`.
       - On `main`: `merge --squash feat`, then `commit -m "Worktree bit 31 (#5)" -m "* feat(tui): play prompt\n\nbody one\n\n* feat(tui): queue view\n\nbody two"`, so `s` = HEAD. Push.
       - Bars `{b1}` and `{b2}`.
     - **Assertions:** `Verdict == Done` and `Landing == s`. Both bars have `Class == Squash`, `Landing == s` and `Repoint == true`.
     - **Boundary:** a `* subject` list with prose body paragraphs between the lines.
   - [ ] Confirm fails: both bars are `Pushed`, and `Verdict == NotDone`.

2. **Implement (GREEN):**
   - [ ] `Squashes`, `Subject`, the rung and the dated `gittest` commits.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/a single commit squash matches on its subject`: on `feat`, `c := r.Commit("fix(tui): render overlay")`. Then `merge --squash` and `commit -m "fix(tui): render overlay (#6)"` with no body. Bar `{c}` is `Squash`. *Boundary:* no `* ` list.
   - [ ] `TestCheck/the earliest squash after the bar wins`: squash `#16`, then a second squash `#17` with the same `* ` lines (the #16/#17 duplicate). The bar lands at `#16`. *Boundary:* duplicate squashes, earliest chosen.
   - [ ] `TestCheck/a track over two squashes lands at the newer`: bars `b1` and `b2` are squashed in `#5`, then `b3` on `feat` is squashed alone as `"<b3 subject> (#6)"`. `b1` and `b2` land at `#5`, `b3` at `#6`, and `Landing == #6`. *Boundary:* the BIT-31 shape, which the old one-squash rule couldn't complete.
   - [ ] `TestCheck/a mention that isn't a list line doesn't match`: a squash body `"see feat(tui): play prompt"` with no `* ` leaves the bar `Pushed`.
   - [ ] `TestCheck/a direct commit with a matching bullet isn't a squash`: a trunk commit `"chore: notes"` whose body has `* feat(tui): play prompt` leaves the bar `Pushed`.
   - [ ] `TestSquashes` (table, `fakeGit`): two records, one squash and one direct. Only the squash is returned, with `Subject` stripped and `Listed` parsed.
   - [ ] `TestSubject` (table, `fakeGit`): `"x\x001759406400"` gives `("x", 1759406400, true)`. An error gives `false`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): a bar listed in a squash on trunk lands at that squash`