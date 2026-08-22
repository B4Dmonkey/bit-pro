# Abort a run

A dispatched track went wrong. Undo the code, put task state back to untouched, leave the
track ready to review and re-approve from scratch.

Give this to Claude along with the track ID.

## Assumptions

- Only the user pushes, so nothing has left this machine and every change is local.
- `.bit/` was committed on the main checkout **before** approving. The reset is only as good
  as that commit — check `git log -1 -- .bit/` before trusting step 4.

## Do these in order. Confirm with me before step 3.

1. **Stop every live session for this track.**
   `claude agents --json` and find rows whose `name` starts `bit/<TRACK>` — including dotted
   bar names. For each, `claude stop <id>`. Report what you stopped; if there were none, say
   so rather than assuming.

2. **Show me what's about to be thrown away.**
   `git log --oneline main..worktree-bit-<TRACK>` and `git diff --stat main...worktree-bit-<TRACK>`.
   Paste the output. Do not skip this — it is the last chance to notice the run produced
   something worth keeping.

3. **Reap the worktree and branch.** After I confirm.
   `claude rm <id>` for each session removes the worktree and deletes the branch together.
   If no session remains to reap it, do it by hand — the worktree is locked, so a plain
   `git worktree remove` will refuse:
   ```
   git worktree unlock .claude/worktrees/bit-<TRACK>
   git worktree remove  .claude/worktrees/bit-<TRACK>
   git branch -D worktree-bit-<TRACK>
   ```
   Verify with `git worktree list` and `git branch` that both are gone.

4. **Reset task state.** From the main checkout, never from a worktree.
   - Every bar under the track back to `todo`: `bp task update <BAR> -s todo`
   - The track itself back to `todo`
   - Clear approval on the track and on every bar
   - If bars were added or edited during the run, restore the pre-run state instead:
     `git checkout -- .bit/` — this is why the assumption above matters.

5. **Confirm the reset.**
   `bp task list --parent <TRACK>` — every bar `todo`, nothing approved, no worktree, no
   branch, no live sessions. Report the list back to me.

## Do not

- Push, or touch any remote.
- Touch `main` — no reset, no rebase, no checkout of the abandoned branch onto it.
- Delete anything under `.bit/feedback/` — notes about what went wrong are the point.
- Mark the track `done` or run `bp task complete`. This is an abort, not a completion.
