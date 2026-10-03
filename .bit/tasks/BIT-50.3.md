---
id: BIT-50.3
title: With no origin/main, local main is trunk, and with neither, the check errors
status: todo
approved: true
phase: 1
phase_label: complete after push
---
## **Verse 1**

A project with no `origin` remote can never have `origin/main`, so with BIT-50.2's rule every bar there reads as not landed. A repo whose commits sit on local `main` with no remote has to come back `done`, and that forces the fallback the scope decides: `origin/main` when it exists, otherwise `main`.

## Scope
- `landing/landing.go`:
  - Trunk resolution tries `ResolveCommit(..., "refs/remotes/origin/main")` first, reported as `Trunk "origin/main"`. Next it tries `refs/heads/main`, reported as `Trunk "main"`. With neither, it returns `ErrNoTrunk`, an exported sentinel: `errors.New("no trunk: neither origin/main nor main exists")`. Trunk is always named `main` (scope).
  - `Branch` is always `"main"` once a trunk resolves.
  - An `origin` remote whose `origin/main` was never fetched also falls back to local `main`. That's the literal reading of "origin/main when it exists".
- Test file: `landing/landing_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/no origin makes local main trunk`
     - **Behavior:** in a repo with nowhere to push, work on local `main` counts as landed.
     - **Setup:** `r := gittest.New(t)` and `r.Git("remote", "remove", "origin")`, which also drops `refs/remotes/origin/*`. Then `a := r.Commit("feat(bit): local")`. Bar `{BIT-1.1 done a}`.
     - **Assertions:** `Trunk == "main"`, `Branch == "main"`, `Verdict == Done`, `Landing == a`.
     - **Boundary:** `origin/main` absent, `main` present.
   - [ ] Confirm fails: `Verdict == NotDone`, because `origin/main` doesn't resolve.

2. **Implement (GREEN):**
   - [ ] The two-step trunk resolution and `ErrNoTrunk`.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/no main at all is an error`: run `r.Git("remote", "remove", "origin")`, then `r.Git("branch", "-m", "main", "trunk")`. The result is `errors.Is(err, ErrNoTrunk)`. *Boundary:* neither ref exists, and other default-branch names aren't supported.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): local main is trunk when origin/main doesn't exist`