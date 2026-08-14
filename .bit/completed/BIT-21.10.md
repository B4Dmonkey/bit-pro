---
id: BIT-21.10
title: README makes the migration runnable by Claude
status: done
phase: 1
phase_label: Migration
---
## **Verse 1**

The script exists and is tested; this is the part that makes it usable the way it is meant to be
used — handed to Claude with a list of directories. `update/README.md` is read on demand rather
than auto-loaded, so it has to be self-contained: someone pointed at it should be able to run the
migration correctly without reading the script.

Not test-driven — it is documentation. The check is that a fresh reader can act on it.

## Scope
- `update/README.md` — new. What the migration does, the exact invocation, what to do before and
  after.

## Content
- [ ] What this is: a one-time migration making every task ID uppercase. Not a product feature —
  there is no `bp doctor`, and nothing detects an un-migrated project.
- [ ] The exact invocation: `bash update/normalize.sh <project-dir>...`, taking directories that
  *contain* `.bit/`, with a worked example naming more than one root.
- [ ] The five carriers it rewrites, and the two things it deliberately does not touch (body
  prose and git commit messages that cite an ID — nothing parses them).
- [ ] Run it on a git-clean tree so the rename shows up as a reviewable diff; the script uses
  `git mv --force`, and a case-only rename is invisible to git otherwise.
- [ ] After: `git status` should show renames, and `bp task list` in each project should show
  every ID uppercase.

## Claude verifies
- [ ] `update/README.md` exists and the invocation line in it matches the script's actual usage
  string character for character
- [ ] `bash update/normalize_test.sh` exits 0

## User verifies
- [ ] Whole slice: copy one lowercase project to a throwaway directory, hand a fresh Claude
  session only `update/README.md` and that path, and ask it to normalize the project. Observe it
  runs the migration without further instruction, and that `bp task list` in the copy afterwards
  shows every ID uppercase with bars still nested under their tracks.

## Commit (user)
`docs(update): add migration instructions`