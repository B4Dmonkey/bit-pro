---
name: bit_commit
description: Commit one bar's work after asking the operator, then record the commit's full hash and branch on the bar and mark it `done`. Use whenever the user says "commit this bar", "commit BIT-N.M", or `/bit:commit`, and whenever bit_do's or theCreator's close-out reaches the commit. It reads the commit message from the bar's `## Commit` section, stages only the files the bar touched, shows the files, the diff stat and the message, and commits only on an explicit yes. A declined commit leaves the bar `doing` with nothing recorded. It only commits — it doesn't push, roll the track up, or start the next bar.
---

# Bar Commit

You commit one bar's work and record which commit it became. The bar ends `done` with its `commit` (the full SHA) and `branch` filled in, so later tools can trace the bar to the commit it landed as.

Every commit goes through the operator's explicit yes, asked in your own words, even on a machine whose permission settings would let the commit through silently. That ask is where the operator reads what is about to be committed.

Every bar write goes through `mcp__bit__task_read` and `mcp__bit__task_update`.

## Steps

1. **Read the bar.** Run `mcp__bit__task_read` on the bar ID. The commit message is the backticked line under `## Commit`, or under `## Commit (user)` on older bars. Refine it only if the work diverged from what the bar describes. If the bar already has a `commit` (a follow-up fix after an unwind, or a small cleanup), reuse that commit's subject line, `git log -1 --format=%s <commit>`, so squash matching later still finds the bar.

2. **No repo.** If `git rev-parse --is-inside-work-tree` fails, the work isn't in a git repo (an `acme/`-style folder, or work outside the repo). Say there's nothing to commit here. On the operator's OK, run `mcp__bit__task_update {id, status: "done"}` with no git fields, and stop.

3. **Pick the files.** Run `git status --porcelain`. Stage only what this bar touched: its Scope files and anything created for it. List any other changes in the tree and leave them alone.

4. **Nothing to commit.** If the bar touched nothing that's changed, say so. On the operator's OK, run `mcp__bit__task_update {id, status: "done"}` with no git fields, and stop. Never make an empty commit with `--allow-empty`.

5. **Ask, before staging anything.** Show the file list, the output of `git diff --stat -- <files>`, the names of any new untracked files among them, and the full commit message. Then wait for the operator. Only an explicit yes goes on.

6. **Declined.** A no, or anything that isn't a yes, means the commit was declined. Say so. The bar stays `doing`, nothing is staged, and nothing is recorded. Stop.

7. **Commit.** Run `git add -- <files>`, then `git commit -m "<message>"`, as two separate Bash calls. Each can raise its own permission prompt. A denied prompt counts as declined: report what's staged, leave the bar `doing`, record nothing, and stop.

8. **Hook fails or rewrites files.** Show the hook's output. If the hook changed files the bar touched, re-stage them and go back to step 5 with the new diff stat. If a lint or test hook failed, the bar's work isn't finished: stop, so the caller fixes it within the bar's scope and runs bit_commit again. Never bypass a hook with `--no-verify`. The bar stays `doing` until a commit succeeds.

9. **Read the result.** Run `git rev-parse HEAD` for the full 40-character SHA, and `git symbolic-ref --short -q HEAD` for the branch. That second command exits non-zero on a detached HEAD, and then the branch is `""`.

10. **Record.** Run one `mcp__bit__task_update {id, commit: "<sha>", branch: "<branch>", status: "done"}`. It returns `approved: true`, because git fields and a forward status move keep approval. A follow-up commit overwrites `commit` the same way.

11. **Report.** Give the short SHA, the branch and the subject line. Then hand back to the caller: bit_do rolls the track up, and theCreator pushes.

## What this skill does not do

- Push. theCreator pushes after a permitted commit; bit_do doesn't push.
- Roll the track up, or mark a track `done`.
- Pass `--no-verify` or `--allow-empty`.
- Amend a commit. A fix after the bar was committed is a new commit.
- Commit without the operator's explicit yes in prose.
