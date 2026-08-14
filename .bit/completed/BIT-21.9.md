---
id: BIT-21.9
title: Case-only renames become visible to git
status: done
phase: 1
phase_label: Migration
---
## **Verse 1**

Measured during planning: a plain two-step `mv` case-only rename leaves `git status --porcelain`
completely **empty** on a case-insensitive filesystem — git never learns the file was renamed, so
the migration cannot be committed or reviewed, and the next checkout restores the lowercase name.
`git mv --force` reports `R  bit-1.md -> BIT-1.md`. This contradicts the rename helper the earlier
bars built, which is what forces this step.

## Scope
- `update/normalize.sh` — use `git mv --force` when the project root is a git working tree, plain
  `mv` otherwise.
- `update/normalize_test.sh` — a git-backed fixture alongside the existing plain one.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_renames_are_visible_to_git`
     - **Behavior:** in a git repository the migration shows up as staged renames.
     - **Setup:** fixture root with `git init`, the lowercase five-carrier tree committed
       (`git -c user.email=… -c user.name=… commit`). Run the script, then read
       `git status --porcelain`.
     - **Assertions:** output is non-empty; it contains a rename entry `R  ` for
       `.bit/tasks/bit-1.md -> .bit/tasks/BIT-1.md`; the feedback note rename appears too.
     - **Boundary:** the case-only rename specifically — the one rename git cannot detect from
       the filesystem alone, and the reason this needs testing at all rather than being assumed.
   - [ ] `test_non_git_project_still_migrates`
     - **Behavior:** the git path does not become a requirement.
     - **Setup:** the existing non-git fixture from the earlier bars.
     - **Assertions:** all five carriers still flip; exit 0.
     - **Boundary:** absence of `.git` — proves the branch is a capability, not a precondition.
   - [ ] Confirm fails: `git status --porcelain` is empty after the migration runs

2. **Implement (GREEN):**
   - [ ] Detect a git working tree per root (`git -C "$root" rev-parse --is-inside-work-tree`) and
     branch the rename helper: `git mv --force` inside a repo, the temp-name `mv` outside one.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): make case-only renames visible to git`