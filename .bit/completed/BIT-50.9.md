---
id: BIT-50.9
title: A project folder with no git reports no_git instead of failing
status: done
approved: true
phase: 2
phase_label: operator decides partial landings
---
## **Verse 2**

An `acme/`-style folder has no git, so `Check` fails today with `ErrNoTrunk`. The scope wants completion to offer filing with no landing commit there, so `task_landing` has to answer `no_git` and not fail.

## Scope
- `git/git.go`: `func IsRepo(ctx context.Context, run Runner, dir string) bool`, which runs `rev-parse --git-dir`. Success means true.
- `landing/landing.go`: `NoGit Verdict = "no_git"`. `IsRepo` is checked first. When it's false, return `Report{Verdict: NoGit}` with `Trunk`, `Branch` and `Landing` empty, each bar's `ID`, `Status` and `Commit` with `Class ""` and `Landing ""`, `Unfinished` filled, and a nil error.
- Test files: `landing/landing_test.go`, `git/git_test.go`, and `cmd/serve_mcp_test.go` (in shape). Both no-git tests run the real `git.ExecRunner` without a `gittest.New` repo, so each calls `gittest.Isolate(t)` (BIT-49.9) first. Otherwise an inherited `GIT_DIR` from the pre-commit hook makes `rev-parse --git-dir` succeed against bit-pro's repo.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/a folder with no git`
     - **Behavior:** a project outside git gets an answer, not an error.
     - **Setup:** `gittest.Isolate(t)`. `Query{Dir: t.TempDir(), Bars: [{BIT-1.1 done ""}, {BIT-1.2 todo ""}]}` with `git.ExecRunner`.
     - **Assertions:** `err == nil`, `Verdict == NoGit`, `Trunk == ""`, `Unfinished == []string{"BIT-1.2"}`.
     - **Boundary:** `rev-parse --git-dir` fails.
   - [ ] Confirm fails: `errors.Is(err, ErrNoTrunk)`.

2. **Implement (GREEN):**
   - [ ] `IsRepo` and the early return.

3. **More tests (RED → GREEN):**
   - [ ] `TestTaskLandingHandler/a project folder with no git reports no_git`: `gittest.Isolate(t)`. The MCP sandbox registers a plain `t.TempDir()` as `ACME` with track `ACME-1` and one done bar. `mcpSessionWithGit(t, dir, git.ExecRunner)` and `task_landing {id: "ACME-1"}` give `verdict == "no_git"` and not `IsError`. *Boundary:* the whole handler path in a non-git project.
   - [ ] `TestCheck/a malformed commit never reaches git` (existing): `IsRepo` now runs first, so the recording fake also answers `rev-parse --git-dir` with `".git"`. The assertions are unchanged.
   - [ ] `TestIsRepo` (table, `fakeGit`): `".git"` gives true, and an error gives false.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): task_landing reports no_git for a folder outside git`