---
id: BIT-50.6
title: /bit:complete files a landed track with its landing commit, and sign-off points to it
status: done
approved: true
phase: 1
phase_label: complete after push
---
## **Verse 1**

The Go side can now say whether a track landed and record where. This bar moves the skills over: completion runs after push, asks `task_landing`, files only a `done` track, and otherwise says "not landed yet". bit:do's sign-off now points to that step. This is a wording-and-behaviour edit to skills, so it adds no Go tests. The skill evals carry the behaviour.

## Scope
Re-read each file first. BIT-49.27/.28/.29 and BIT-47.4 reword these sections, so anchor on section names, not line numbers.
- `bit/skills/complete/SKILL.md`:
  - **description:** completes a whole track **after its work has landed** (pushed or merged). It checks where the work landed with `mcp__bit__task_landing`, records the landing commit, marks everything done and files it. Triggers: "BIT-N is pushed, complete it", "it's merged, close out BIT-N", "mark the track done", "file BIT-N as completed", "we're done with BIT-N". Drop "gives sign-off at the end of a bit_do cycle". Keep the line that a single bar just gets its status set.
  - **intro:** replace "The user invoking this is the sign-off" with: completion runs once the work is pushed or merged, and "completed" means the work is on trunk, with its landing commit recorded.
  - **For each track:**
    1. Read it (unchanged).
    2. `mcp__bit__task_landing {id}`.
    3. If `unfinished` isn't empty, stop: "finish or mark these bars done first: <IDs>". Nothing is changed. Verse 2 replaces this with a check-in.
    4. If `verdict` isn't `done`, stop with "<ID> hasn't landed yet: fetch or pull, and push if the commits are only local, then run `/bit:complete <ID>` again". Leave out "push" when `trunk` is `main` (no `origin/main`). Nothing is filed.
    5. On `done`: `mcp__bit__task_update` the track with `status: done`, ticking unchecked verses in the same `body`.
    6. `mcp__bit__task_complete {id, commit: <landing>, branch: <branch>}`.
    7. Confirm with `task_list` (unchanged).
  - Say why the order matters: every write comes before `task_complete`, because `task_update` can't reach a filed task.
  - Delete the old step 2, which forced bars to `done` without asking.
  - **Report:** one line per track: the ID, the landing commit (first 12 characters) and branch, and filed. Delete the commit suggestion. In v2 the store is outside the repo, so filing leaves nothing to commit.
- `bit/skills/complete/evals/evals.json`:
  - eval 1 becomes `"BIT-7 is pushed to main, complete it"`. Expected: `task_landing` is called and reports `done`; the track is set `done` with its verses checked off; `task_complete` is passed `commit` = the landing and `branch` = `main`; the track is absent from `task_list`; no commit is suggested.
  - eval 2 keeps `"we're done with bit-7, close it out"`, with BIT-7 holding a todo bar. Expected: it stops and names the unfinished bar, flips nothing and files nothing.
  - new eval 4, `"complete BIT-8"`, whose bars are committed but not pushed. Expected: "not landed yet: fetch or pull, push, and run `/bit:complete BIT-8` again", and nothing is filed.
  - eval 3 is unchanged.
- `bit/skills/do/SKILL.md`:
  - **description:** the sign-off sentence becomes: when every bar is done it stops; the operator pushes (or merges) and runs `/bit:complete`, which records where the track landed and files it.
  - **intro tool list:** drop `mcp__bit__task_complete` from do's tools, since do never calls it, and fix the count.
  - **Roll the track up → Track status:** the track stays `doing` until `/bit:complete` runs after the push. Drop "the human signs it off … marks it done".
  - **Compaction point:** on the last bar, point at **Track sign-off** (keep the pointer).
  - **Track sign-off:** rewrite. When every bar is done, tell the operator the track is ready, and that the next step is to push (or merge) the work and then run `/bit:complete <track>`, which checks that it landed, records the landing commit and files it. The track stays `doing` until then. Don't set it `done` and don't call `task_complete`. If they aren't ready, nothing changes.
  - **What this skill does not do:** "Declare a track done on its own" now says that after the last bar the operator pushes and runs `/bit:complete`.
- `bit/agents/bot.md`:
  - **What you do yourself:** in the last paragraph, when a track is fully done, say so, tell them to push (or merge) and then run `/bit:complete`, and stop.
  - **routing table:** the row becomes "a track's work is pushed or merged and they want it closed out or completed → `bit:complete`".
- `bit/agents/bot-dev.md`, the **Track sign-off** bullet under "What stays the operator's": add "tell the operator to push if needed, then run `/bit:complete`".
- `cmd/task/complete.go` `Short`: `"File a track and its bars as completed, with no landing check (manual override)"`.
- `README.md`: re-grep `grep -n 'sign' README.md` after BIT-49.30. The `bp task complete` row uses the same wording as `Short`, and any "signed-off work" becomes "completed work".

## Change checklist
- [ ] Apply the edits.
- [ ] `grep -n 'sign-off\|signed' bit/skills/complete/SKILL.md` finds nothing that says sign-off completes a track.
- [ ] `grep -n 'task_complete' bit/skills/do/SKILL.md` finds nothing.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); for s in complete do; do uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/$s/; done`. Only the known kebab-case `name:` failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `python3 -m json.tool bit/skills/complete/evals/evals.json > /dev/null`
- [ ] `just lint` and `just test` pass

## User verifies
This checks the whole verse in a sandbox. Never `just install`.
- [ ] Set up as in BIT-46.11 (`$SB/bin/bp`, the fake `claude`, and `HOME`/`XDG_DATA_HOME` exported). Then build a project with a local bare origin: `git init --bare -b main $SB/origin.git && git init -b main $SB/proj && cd $SB/proj && git remote add origin $SB/origin.git && git -c user.name=t -c user.email=t@t commit --allow-empty -m init && git push -u origin main && printf 'demo\n' | bp add .`
- [ ] In a new shell without the `HOME` export, with `$SB/dev.json` as in BIT-46.11, run `cd $SB/proj && claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config`. Ask for a track with one bar, add a file, and commit it for that bar through `bit:commit`.
- [ ] Before pushing, `/bit:complete DEMO-1` says it hasn't landed, tells you to push, and `bp task list` still shows `DEMO-1`.
- [ ] Run `git push`, then `/bit:complete DEMO-1`. It reports filed at `git rev-parse HEAD` on `main`. `bp task list` no longer shows `DEMO-1`, and `$SB/data/bit/DEMO/completed/DEMO-1.json` has `"commit"` = that SHA, `"branch": "main"` and `"status": "done"`.
- [ ] Whole slice: a track is completed only once its work is pushed, and it records the commit it landed at.

## Commit
`feat(bit): complete runs after push and records the landing commit`