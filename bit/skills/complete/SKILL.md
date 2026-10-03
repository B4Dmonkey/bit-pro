---
name: bit_complete
description: Complete one or more whole tracks after their work has landed (pushed or merged) — check where the work landed with `mcp__bit__task_landing`, record the landing commit, mark everything done and file the track and its bars as completed through `mcp__bit__task_complete`, so the work actually leaves the active list. Use whenever the user says "BIT-N is pushed, complete it", "it's merged, close out BIT-N", "mark the track done", "file BIT-N as completed", "we're done with BIT-N" — any time the user means a whole track (an ID with no dot) is finished, even if they don't say "complete". Setting the track's status to `done` alone is not enough; it leaves the track sitting in the active list, and this skill exists to close that gap. For a single bar ("mark BIT-24.2 done"), just set that bar's status — this skill is for tracks.
---

# Track Completion

Completion runs once the track's work is pushed or merged. "Completed" means the work is on trunk, with its landing commit recorded on the track. Your job is to check that it landed, then make the finished state real: the track `done`, its landing commit and branch recorded, and the track plus its bars filed as completed so they leave `task_list`, the board, and the TUI.

The failure this skill prevents: the track's status gets flipped to `done` and the work stops there. A `done` track that was never filed still clutters the active list and looks unfinished to every other tool. **A track isn't complete until `mcp__bit__task_complete` has run on it.** Filing is the last step, and the most important one.

Every write goes through the `mcp__bit__*` tools. They are the only way in.

## Which tracks

Take every track ID the user named. IDs are case-insensitive, and the tools normalize them. A track ID has no dot (`BIT-43`). If the user named a bar (`BIT-43.2`) and meant its track, use the track. If they named nothing and the conversation doesn't make the track obvious, ask which one.

## For each track

1. **Read it.** Run `mcp__bit__task_list` with `parent` set to the track to get its bars, and `mcp__bit__task_read` on the track to get its body.
2. **Check where it landed.** Run `mcp__bit__task_landing` with the track's `id`. It reads git and never writes. It returns `trunk`, `branch`, `verdict`, `landing`, the `unfinished` bars, and each bar's class.
3. **Stop on unfinished bars.** If `unfinished` isn't empty, stop and say: "finish or mark these bars done first: <IDs>". Change nothing.
4. **Stop if it hasn't landed.** If `verdict` isn't `done`, stop and say: "<ID> hasn't landed yet: fetch or pull, and push if the commits are only local, then run `/bit:complete <ID>` again". When `trunk` is `main` (there's no `origin/main`), leave out the push part. File nothing.
5. **Mark the track done.** On `done`, run `mcp__bit__task_update` on the track with `status: done`. If the body has unchecked verse items (`- [ ]`), check them off in the same call by passing the edited `body`, so the track body agrees with its status.
6. **File it.** Run `mcp__bit__task_complete` with the track's `id`, `commit` set to `landing`, and `branch` set to `branch`. This records the landing commit on the track, then files the track and all its bars as completed.
7. **Confirm it landed.** Run `mcp__bit__task_list` with no `parent`. The track must be gone from the list. If it's still there, or `task_complete` returned an error, report the exact error and stop. Never tell the user a track is complete when it's still listed.

The order matters: every write comes before `task_complete`, because `task_update` can't reach a task once it's filed.

If one track fails, finish the others and report each track's result separately.

## Report

Keep the report short, one line per track: the ID, its landing commit (first 12 characters) and branch, and that it was filed. There's nothing to commit afterwards: the store lives outside the repo, so filing leaves the working tree untouched.
