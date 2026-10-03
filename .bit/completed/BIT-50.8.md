---
id: BIT-50.8
title: An unfinished bar makes an otherwise landed track partly done
status: done
approved: true
phase: 2
phase_label: operator decides partial landings
---
## **Verse 2**

Hashes can't see unfinished work. A bar that was never committed has no hash, and a migrated bar carries HEAD whatever its status. So a track whose hashes all landed but which still has a `doing` bar would read `done`. The scope says an unfinished bar makes the track partly done at best, and this test forces it.

## Scope
- `landing/landing.go`: after the hash verdict, if `Unfinished` isn't empty and the verdict is `Done`, it becomes `Partly`. The other verdicts are unchanged, and `Landing` keeps the newest landed bar's landing.
- Test file: `landing/landing_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/an unfinished bar makes a landed track partly done`
     - **Behavior:** the operator is asked about unfinished bars even when every hash landed.
     - **Setup:** commit `a` and `b`, then push `main`. Bars `{BIT-1.1 done a}` and `{BIT-1.2 doing b}`.
     - **Assertions:** `Verdict == Partly`, `Landing == b`, `Unfinished == []string{"BIT-1.2"}`.
     - **Boundary:** every hash landed, with one status other than `done`.
   - [ ] Confirm fails: `Verdict == Done`.

2. **Implement (GREEN):**
   - [ ] The unfinished downgrade.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/a todo bar with no hash`: `{a}` (done, landed) and `{BIT-1.2 todo ""}` give `Partly`. *Boundary:* a `NoHash` bar that is also unfinished, which would otherwise be ignored.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): an unfinished bar makes a landed track partly done`