---
id: BIT-21.8
title: A root without .bit/ is an error, not a silent skip
status: done
phase: 1
phase_label: Migration
---
## **Verse 1**

The invocation contract: the script takes project roots, so a directory with no `.bit/` inside
is operator error and must say so. A silent skip is the dangerous outcome — the operator would
see exit 0, believe a project was migrated, and only find out later when a note gets clobbered.

## Scope
- `update/normalize.sh` — argument validation and exit status.
- `update/normalize_test.sh` — negative cases.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_directory_without_bit_is_an_error`
     - **Behavior:** a root with no `.bit/` fails loudly and changes nothing anywhere.
     - **Setup:** two roots — a valid lowercase fixture and an empty directory. Run
       `bash update/normalize.sh "$empty"`, then run it with both roots.
     - **Assertions:** exit status is non-zero; stderr names the offending directory; when both
       roots are passed, the valid project is left exactly as it was — the script validates every
       argument before touching any of them.
     - **Boundary:** one bad argument among several — the case where partial work would leave the
       operator with a half-migrated set and no clear signal which half.
   - [ ] `test_no_arguments_is_an_error`
     - **Behavior:** invoking with no roots is refused rather than treated as "nothing to do".
     - **Setup:** run `bash update/normalize.sh` with no arguments.
     - **Assertions:** exit status non-zero; stderr shows a usage line naming the
       `<project-dir>...` form.
     - **Boundary:** argument count 0 — the lower bound.
   - [ ] Confirm fails: the empty directory currently produces exit 0 with no message

2. **Implement (GREEN):**
   - [ ] Validate every argument up front: at least one root, and `<root>/.bit` a directory for
     each. Print usage to stderr and exit non-zero otherwise. Only start rewriting once all
     arguments pass.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): validate project roots before migrating`