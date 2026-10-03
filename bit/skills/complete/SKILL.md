---
name: bit_complete
description: Sign off one or more finished tracks — mark every bar done, mark the track done, and file the track and its bars as completed through `mcp__bit__task_complete`, so the work actually leaves the active list. Use whenever the user says "mark BIT-N done", "mark the track as done", "close out BIT-N", "sign off on this track", "file it as completed", "we're done with BIT-N", or gives sign-off at the end of a bit_do cycle — any time the user means a whole track (an ID with no dot) is finished, even if they don't say "complete". Setting the track's status to `done` alone is not enough; it leaves the track sitting in the active list, and this skill exists to close that gap. For a single bar ("mark BIT-24.2 done"), just set that bar's status — this skill is for tracks.
---

# Track Completion

The user invoking this is the sign-off. Your job is to make the track's finished state real on disk: every bar `done`, the track `done`, and the track plus its bars filed as completed so they leave `task_list`, the board, and the TUI.

The failure this skill prevents: the track's status gets flipped to `done` and the work stops there. A `done` track that was never filed still clutters the active list and looks unfinished to every other tool. **A track isn't complete until `mcp__bit__task_complete` has run on it.** Filing is the last step, and the most important one.

Every write goes through the `mcp__bit__*` tools. They are the only way in.

## Which tracks

Take every track ID the user named. IDs are case-insensitive, and the tools normalize them. A track ID has no dot (`BIT-43`). If the user named a bar (`BIT-43.2`) and meant its track, use the track. If they named nothing and the conversation doesn't make the track obvious, ask which one.

## For each track

1. **Read it.** Run `mcp__bit__task_list` with `parent` set to the track to get its bars, and `mcp__bit__task_read` on the track to get its body.
2. **Mark every unfinished bar done.** Run `mcp__bit__task_update` with `status: done` on each bar that isn't already `done`, including bars that were never started. Don't stop to ask: the user already signed off, and `task_complete` refuses a track with any unfinished bar.
3. **Mark the track done.** Run `mcp__bit__task_update` on the track with `status: done`. If the body has unchecked verse items (`- [ ]`), check them off in the same call by passing the edited `body`, so the track body agrees with its status. `task_complete` only moves files and never sets the status, so skipping this step files a track that still says `doing`.
4. **File it.** Run `mcp__bit__task_complete` on the track. This files the track and all its bars as completed.
5. **Confirm it landed.** Run `mcp__bit__task_list` with no `parent`. The track must be gone from the list. If it's still there, or `task_complete` returned an error, report the exact error and stop. Never tell the user a track is complete when it's still listed.

If one track fails, finish the others and report each track's result separately.

## Report

Keep the report short, one line per track: the ID, how many bars you flipped (name them if they were never started, so the user can see what got force-closed), and that it was filed. Then suggest a commit. Suggest a message like `chore(bit): file completed BIT-43`, or list every ID if you filed several. The user runs the commit; you don't.
