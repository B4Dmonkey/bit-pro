---
id: BIT-21.1
title: Uppercase task filenames under tasks/
status: done
approved: true
phase: 1
phase_label: Migration
---
## **Verse 1**

Establishes `update/normalize.sh` and its bash harness, and makes the first carrier work: task
filenames under `.bit/tasks/`. The fixture seeds two differently-named files, so a hardcoded
single rename cannot satisfy the test — a real loop is forced from the start.

## Scope
- `update/normalize.sh` — new. Takes one or more project roots, resolves `<root>/.bit`, renames
  task files under `.bit/tasks/` so the ID stem is uppercase.
- `update/normalize_test.sh` — new. The harness: builds a throwaway lowercase project under
  `$(mktemp -d)`, runs the script, asserts, exits non-zero on the first failure.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_task_filenames_are_uppercased`
     - **Behavior:** every task file under `.bit/tasks/` ends up with an uppercase ID stem, and
       the `.md` extension is left alone.
     - **Setup:** temp project root containing `.bit/tasks/bit-1.md` and `.bit/tasks/bit-1.2.md`
       — a track and a bar, so the dotted form is exercised — each with valid frontmatter. Run
       `bash update/normalize.sh "$root"`.
     - **Assertions:** `.bit/tasks/BIT-1.md` and `.bit/tasks/BIT-1.2.md` both exist; neither
       `bit-1.md` nor `bit-1.2.md` remains as a distinct entry; the extension is exactly `.md`,
       never `.MD`; `.bit/tasks/` still holds exactly 2 files.
     - **Boundary:** two files sharing a prefix — the lower bound at which a hardcoded single
       rename stops working. The `.MD` assertion pins the other easy mistake: uppercasing the
       whole filename instead of just the stem.
   - [ ] Confirm fails: `update/normalize.sh: No such file or directory`

2. **Implement (GREEN):**
   - [ ] `update/normalize.sh`, `#!/usr/bin/env bash` with `set -euo pipefail`. Iterate the
     project roots passed as arguments; for each, loop `.bit/tasks/*.md`, uppercase the stem via
     `tr '[:lower:]' '[:upper:]'` (avoid `${v^^}` — `/bin/bash` here is 3.2), keep `.md`.
   - [ ] Rename through a temp name (`mv a tmp && mv tmp A`) so a case-only rename lands on a
     case-insensitive filesystem. Filesystem correctness only — git visibility is its own bar.
   - [ ] `update/normalize_test.sh` with the fixture builder and a `fail()` that prints the
     failing assertion and exits 1.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): uppercase task filenames in tasks/`