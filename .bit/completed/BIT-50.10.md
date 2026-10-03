---
id: BIT-50.10
title: A shallow clone is reported as shallow and its bars aren't classed
status: done
approved: true
phase: 2
phase_label: operator decides partial landings
---
## **Verse 2**

In a shallow clone, older commits are missing and the ancestry answers are wrong, so landed bars read as unresolvable. The scope says to report the clone as shallow, with "run `git fetch --unshallow`", instead of classing its bars. A shallow clone whose first bar sits below the depth forces it.

## Scope
- `git/git.go`: `func IsShallow(ctx context.Context, run Runner, dir string) bool`, which runs `rev-parse --is-shallow-repository`. Output `"true"` means true, and anything else or an error means false.
- `landing/landing.go`:
  - `Report` gains `Shallow bool \`json:"shallow"\``.
  - After the trunk resolves, when `IsShallow` is true, return `Trunk` and `Branch` as resolved, `Shallow: true`, `Verdict: CantTell`, `Landing ""`, every bar with `Class ""`, and `Unfinished` filled, with no ancestry calls. No bar is classed `Unresolvable`.
  - The skill checks `shallow` before `verdict` (BIT-50.11).
- Test files: `landing/landing_test.go` and `git/git_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/a shallow clone is reported and not classed`
     - **Behavior:** a shallow clone gets "unshallow and retry", not a false "unresolvable".
     - **Setup:** `r := gittest.New(t)`. Commit `a`, `b` and `c`, then push. Run `r.Git("clone", "--depth", "1", "file://"+r.Origin, sd)`, where `sd` is a new temp path. `file://` is required, because a plain local path ignores `--depth`. Run `Check` with `Dir: sd` and bars `{a}` and `{c}`.
     - **Assertions:** `Shallow == true`, `Verdict == CantTell`, and both classes are `""`.
     - **Boundary:** a bar commit below the clone's depth.
   - [ ] Confirm fails: `Report` has no `Shallow` field (compile error).

2. **Implement (GREEN):**
   - [ ] `IsShallow`, the field and the early return.

3. **More tests (RED → GREEN):**
   - [ ] `TestIsShallow` (table, `fakeGit`): `"true"` gives true, `"false"` gives false, and an error gives false.
   - [ ] Existing full-clone rows still report `Shallow == false` (add the assertion to `TestCheck/every bar pushed`).

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): task_landing reports a shallow clone instead of classing its bars`