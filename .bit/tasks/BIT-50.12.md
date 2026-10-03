---
id: BIT-50.12
title: A commit answer places every bar that hasn't landed, and is refused when it isn't on trunk
status: todo
phase: 3
phase_label: PR or commit answer
---
## **Verse 3**

When a track's hashes are empty or stale, or its work isn't on trunk, the operator knows which commit landed the work. A rebase-merge or a GitHub squash that doesn't list the bars' subjects leaves the original commits resolving on a branch or locally, so those bars read as pushed or local. A query carrying the landing commit has to place every bar that hasn't landed and come back `done`, and that forces the answer path. Bars that landed keep their own commit (scope).

## Scope
- `landing/landing.go`:
  - `Query` gains `Commit string`, the operator's answer. It's empty when there's no answer.
  - `BarResult` gains `Repoint bool \`json:"repoint"\``. When it's true, the skill writes the bar's `Landing` as its `commit` before filing.
  - Sentinels: `ErrBadAnswer` ("not a commit") and `ErrNotOnTrunk` ("not on <trunk>: fetch and retry", wrapped with the trunk name).
  - With `q.Commit` set:
    - It must match `^[0-9a-fA-F]{4,40}$`. A human may abbreviate, and the check keeps it from reaching git as an option. Otherwise return `ErrBadAnswer`.
    - `ResolveCommit` it to the full SHA. If it doesn't resolve, or `IsAncestor(full, trunk)` is false, return `ErrNotOnTrunk`.
    - The answer's landing uses the same first-parent rule as any bar (BIT-50.4).
    - Bars classed `NoHash`, `Unresolvable`, `Pushed` or `Local` get `Landing = answer landing` and `Repoint = true`, and count as landed for the verdict. Their `Class` is unchanged.
  - The track's `Landing` is still the newest landing, which equals the answer when no other bar landed.
  - The answer is ignored for `no_git` and `shallow`, whose early returns come first.
- `cmd/serve_mcp.go`: `taskLandingInput` gains `Commit string \`json:"commit,omitempty"\``, passed as `Query.Commit`. The description gains one sentence: `commit` (and later `pr`) is the operator's answer when git can't place the work or it isn't on trunk.
- Test files: `landing/landing_test.go` and `cmd/serve_mcp_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/a commit answer places bars that haven't landed`, a table with one row per unplaced class. Each row builds its bar in a fresh `gittest` repo, then `x := r.Commit("feat(bit): landed by hand")` on `main` and pushes `main`. `Query.Commit = x[:12]`.
     - **Behavior:** the operator's commit completes a track whose own hashes are empty, stale, or off trunk.
     - **Rows:**
       - `no hash`: bar `{BIT-1.1 done ""}`, want class `NoHash`.
       - `stale hash`: bar `{BIT-1.1 done "0123456789abcdef0123456789abcdef01234567"}`, want class `Unresolvable`.
       - `pushed to a branch`: on `feat`, `b := r.Commit("feat(bit): one")` and `r.Git("push", "origin", "feat")`, then `checkout main`. Bar `{BIT-1.1 done b}`, want class `Pushed`.
       - `local only`: on an unpushed `feat`, `b := r.Commit("feat(bit): one")`, then `checkout main`. Bar `{BIT-1.1 done b}`, want class `Local`.
     - **Assertions (every row):** `Verdict == Done`, `Landing == x`, and the bar has `Landing == x`, `Repoint == true` and the row's class.
     - **Boundary:** an abbreviated answer, with each class of unplaced bar, including a rebase-merge's surviving originals.
   - [ ] Confirm fails: `Query` has no `Commit` field (compile error).

2. **Implement (GREEN):**
   - [ ] The answer path, `Repoint` and the sentinels.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/an answer not on trunk is refused`: `y` is committed on an unpushed `feat`, with `Query.Commit = y`. The result is `errors.Is(err, ErrNotOnTrunk)`. *Boundary:* a real commit off trunk.
   - [ ] `TestCheck/an answer that isn't a commit is refused`: `"--all"` and `"zz"` both give `ErrBadAnswer`, and the recording fake runner saw neither value. It reuses `TestCheck/a malformed commit never reaches git`'s fake, which answers `rev-parse --git-dir`, the trunk `rev-parse` and `rev-list --first-parent`, so the refusal comes from the answer check and not from a missing key. *Boundary:* untrusted operator input.
   - [ ] `TestCheck/a landed bar keeps its own commit`: bar `{a}` landed, bar `{b}` pushed only to `origin/feat`, and bar `{""}`, answer `x`. Bar 1 has `Repoint == false` and `Landing == a`. Bars 2 and 3 have `Repoint == true` and `Landing == x`. *Boundary:* repoint every bar that hasn't landed, and only those.
   - [ ] `TestTaskLandingHandler/a commit answer reaches the check`: a track whose only bar has no hash. `task_landing {id, commit: x}` gives `verdict == "done"` and `bars[0].repoint == true`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): task_landing places unlanded bars at the operator's commit`