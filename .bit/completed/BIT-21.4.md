---
id: BIT-21.4
title: Contradiction forces all three task directories
status: done
approved: true
phase: 1
phase_label: Migration
---
## **Verse 1**

Task files live in three directories, not one: `.bit/tasks/`, `.bit/completed/`, and
`.bit/archive/tasks/`. The contradiction is direct — the previous bars' loop only walks
`tasks/`, so a fixture with files in the other two cannot pass without generalising it.

This matters beyond tidiness: the ID-minting scan reads all three, so a lowercase file left in
`completed/` is exactly what re-mints a live ID onto a new task.

## Scope
- `update/normalize.sh` — apply the rename and both frontmatter rewrites across all three task
  directories.
- `update/normalize_test.sh` — fixture seeded in `completed/` and `archive/tasks/`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_completed_and_archived_tasks_are_normalized`
     - **Behavior:** a task file is normalized wherever it lives, not only in `tasks/`.
     - **Setup:** `.bit/tasks/bit-1.md`, `.bit/completed/bit-2.md` (with `id: bit-2`), and
       `.bit/archive/tasks/bit-3.md` (with `id: bit-3`). Run the script.
     - **Assertions:** `.bit/completed/BIT-2.md` and `.bit/archive/tasks/BIT-3.md` exist and each
       contains its uppercase `id:`; `.bit/tasks/BIT-1.md` is still correct; no lowercase task
       filename remains under `.bit/`.
     - **Boundary:** all three directories populated at once — the upper bound of the directory
       set. `archive/tasks/` is the nested one and the easiest to miss with a shallow glob.
   - [ ] Confirm fails: `completed/bit-2.md` and `archive/tasks/bit-3.md` are untouched

2. **Implement (GREEN):**
   - [ ] Lift the per-directory work into a function and call it for `tasks`, `completed`, and
     `archive/tasks`. Skip a directory that does not exist rather than erroring — a project with
     nothing archived is normal.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): normalize completed and archived tasks`