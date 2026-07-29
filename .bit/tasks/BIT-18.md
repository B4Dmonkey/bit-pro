---
id: BIT-18
title: Separate completed work from soft-deleted work
status: doing
---
## Why

Finished work and thrown-away work land in the same place. `bp task archive` (a track
signed off) and `bp task delete` (a bar dropped mid-plan) both call the same relocate into
`.bit/archive/`, so once a file is there nothing tells a shipped deliverable apart from a
mistake. This project's archive is already 125 files across sixteen completed tracks, and
anyone reading back over finished work — a person, or a future retro pass — has to guess.
It also makes "archive" mean two opposite things at once, which is why signing a track off
reads like filing it away rather than recording that it shipped.

## Summary

Split the one destination in two. `bp task complete <id>` files a signed-off track and its
bars into `.bit/completed/`. `bp task delete` keeps its soft-delete job but writes to
`.bit/archive/tasks/`, mirroring `.bit/`'s own layout so archive has somewhere to put any
other artifact it may one day hold. The `bp task archive` command goes away — the folder
keeps the name, the verb doesn't. Everything that quietly depended on scanning one
directory (ID reservation, the feedback note's track lookup) learns about both. Finally the
skills and the `bp` contract are updated to name the new verb, so a live session files
finished work in the right place.

## Visual aid

```
before                          after

.bit/                           .bit/
  tasks/                          tasks/
    BIT-17.md                       BIT-18.md
  archive/                        completed/
    BIT-16.md   <- signed off       BIT-17.md      <- signed off
    BIT-16.3.md <- dropped bar    archive/
  feedback/                         tasks/
    BIT-17-001.md                     BIT-16.3.md  <- dropped bar
                                  feedback/
                                    BIT-18-001.md
```

## Decisions

- **Two verbs, two destinations.** `bp task complete` → `.bit/completed/`; `bp task delete`
  → `.bit/archive/tasks/`. `bp task archive` is removed rather than kept as a synonym —
  a command that means "file this as done" and a folder that means "soft delete" is the
  confusion being fixed, so neither name is left pointing at the other's job.
- **`completed/` is flat.** `.bit/completed/BIT-17.md`. It only ever holds tasks, so it
  needs no interior structure.
- **`archive/` mirrors `.bit/`.** `.bit/archive/tasks/BIT-16.3.md`. Archive is the soft
  delete for anything under `.bit/`, so its layout matches what it shadows and a future
  `.bit/archive/feedback/` needs no new rule.
- **IDs stay reserved across all three directories.** `NextID`/`NextChildID` take the
  highest number found in `tasks/`, `completed/`, and `archive/tasks/` and count up, so an
  ID is never re-minted onto different work and older commit messages stay valid. The
  legacy flat `.bit/archive/*.md` path is *not* scanned — so until a project's one-time move
  below has run, its flat archive doesn't reserve anything. Do the move before creating new
  work in that project.
- **A feedback note attaches to an active, completed, or archived track.** Notes are most
  valuable about finished work, so the track lookup accepts all three and neither
  completing nor deleting a track touches `.bit/feedback/`.
- **`complete` has no `--force`.** A sign-off is a claim that the work is finished, so the
  all-bars-`done` guard is the point of the verb rather than an obstacle to it. Mark the bars
  `done` first. `delete` keeps `--force`, because dropping unfinished work is the normal case
  there.
- **The existing flat archives move by hand, once per project, as the last step of the whole
  track.** Four projects across two devices have one; it isn't code and isn't a step in this
  plan, and it can't be sequenced earlier because not every project is reachable from here:

  ```bash
  mkdir -p .bit/completed && git mv .bit/archive/*.md .bit/completed/
  ```

## Verses

- [x] Verse 1 — Signing a track off files it as completed, not in the same bin as deleted work: `bp task complete <id>` moves the track and its bars to `.bit/completed/`, and the two things that read the old single directory keep working — the next created ID still counts up past the completed one, and a feedback note still attaches to a track that's been completed.
  Touches: `task/store.go` (the relocate destination and `highestReserved`), `task/feedback.go` (`trackExists`), a new `cmd/task_complete.go`, `cmd/task.go` — where to look to verify.
- [x] Verse 2 — Deleting work files it under archive, laid out like the rest of `.bit/`: `bp task delete` writes to `.bit/archive/tasks/`, and `bp task archive` no longer exists, so the only way to file something as done is to say `complete`.
  Touches: `task/store.go`, `cmd/task_archive.go` (removed), `cmd/task.go` — where to look to verify.
- [ ] Verse 3 — A live session files finished work in the right place: bit_do's track sign-off, the feedback skill's wording, and the `bp instructions` contract all name `complete` and the two destinations, so following the skills produces the new layout instead of the old one.
  Touches: `bit/skills/do/SKILL.md`, `bit/skills/feedback/SKILL.md`, `assets/bit-cli.md` — where to look to verify.