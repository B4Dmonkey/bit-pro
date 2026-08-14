---
id: BIT-21.2
title: Frontmatter id flips, title and prose do not
status: todo
phase: 1
phase_label: Migration
---
## **Verse 1**

Renaming the file is not enough — the ID is also stored inside it. This step rewrites the `id:`
frontmatter field, and the same test pins the blast radius: `title:` and body prose must come
through untouched, because the scope deliberately leaves prose citations alone.

## Scope
- `update/normalize.sh` — after renaming, rewrite each task file's `id:` value in place.
- `update/normalize_test.sh` — extend the fixture with a lowercase title and a body that cites
  an ID in prose.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_id_frontmatter_is_uppercased_and_nothing_else`
     - **Behavior:** the `id:` value becomes uppercase; every other byte of the file is
       unchanged.
     - **Setup:** `.bit/tasks/bit-1.md` with `id: bit-1`, `title: bit rot in the ingest step`,
       and a body line reading `follow-on work tracked at bit-1.2`. Run the script.
     - **Assertions:** `BIT-1.md` contains exactly `id: BIT-1`; the title line still reads
       `title: bit rot in the ingest step` character for character; the body line still reads
       `follow-on work tracked at bit-1.2`.
     - **Boundary:** three lowercase `bit` occurrences in one file, only one of which is a
       carrier — proves the rewrite is anchored to the `id:` field rather than applied to the
       file, which is the failure mode that would silently corrupt every scope's prose.
   - [ ] Confirm fails: `id: bit-1` is still present — the file was renamed but not rewritten

2. **Implement (GREEN):**
   - [ ] In `normalize.sh`, for each task file rewrite only a line matching `^id: ` , uppercasing
     the value. Use `sed -i ''` (BSD sed on macOS needs the empty backup argument).

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): uppercase the id frontmatter field`