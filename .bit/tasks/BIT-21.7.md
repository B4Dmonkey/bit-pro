---
id: BIT-21.7
title: Re-running on a normalized project changes nothing
status: done
phase: 1
phase_label: Migration
---
## **Verse 1**

All five carriers now flip. This step proves the script is safe to run more than once and safe
to point at a project that is already correct — the two situations the operator will actually
be in, given two of the three projects on this machine are already uppercase.

## Scope
- `update/normalize.sh` — no new behaviour expected; make it idempotent if it is not.
- `update/normalize_test.sh` — two tests over the same fixture builder.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_second_run_changes_nothing`
     - **Behavior:** running the script twice produces the same tree as running it once.
     - **Setup:** the full five-carrier lowercase fixture. Run the script, snapshot the tree
       (`find "$root/.bit" -type f | sort` plus a checksum of each file's contents), run it
       again, snapshot again.
     - **Assertions:** the two snapshots are identical.
     - **Boundary:** run count 1 → 2, the point at which a non-idempotent rewrite (double
       uppercasing, a re-appended suffix, a rename onto an existing name) would show.
   - [ ] `test_already_uppercase_project_is_untouched`
     - **Behavior:** a project that never needed migrating comes through unchanged.
     - **Setup:** a fixture built entirely uppercase — `BIT-1.md` with `id: BIT-1`, an order
       list, `BIT-1-001.md`, `prefix = "BIT"`. Snapshot, run, snapshot.
     - **Assertions:** snapshots identical; the script still exits 0.
     - **Boundary:** zero files needing change — the lower bound, and the state bit-pro and the
       marketplace clone are already in.
   - [ ] Confirm fails: expected to pass if the earlier bars were written cleanly; if either
     fails, the cause is the rename step renaming a file onto itself

2. **Implement (GREEN):**
   - [ ] Only act when the uppercase form differs from what is on disk — skip the rename and the
     rewrites otherwise.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): make normalization idempotent`