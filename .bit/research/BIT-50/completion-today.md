# Completion today, and what moves when it runs after landing

**Checked:** how a track is completed now (skill, MCP, CLI, store), and what changes when completion runs after the work is pushed instead of at sign-off.

## What the code does (v2 @ 6a1d345)
- `bit/skills/complete/SKILL.md`: the operator's invocation *is* the sign-off. Steps: list bars, force every unfinished bar to `done` (no asking), set track `done` and tick verse boxes via `task_update`, then `task_complete`, then confirm the track left `task_list`. Suggests a `chore(bit): file completed BIT-N` commit; user commits. No git reads at all.
- `cmd/serve_mcp.go:76-82` description, `:160-162` `taskCompleteInput{ID}` only, `:417-430` handler → `store.Complete(in.ID)`.
- `cmd/task/complete.go`: `bp task complete <id>` → same `Store.Complete`.
- `task/store.go:108-147`: `Complete` = `relocateTree(completedDir, id, force=false)`: refuses with `UnfinishedBarsError` if any bar isn't `done`, then `os.Rename`s bars then track into `.bit/completed/`. It never sets status and touches no git.
- `task/store.go:37-39,164`: `Path`/`Load` only look in `.bit/tasks/`, so `Update` on a task **after** it was completed fails. Consequence: bar repointing and the track's `commit` must be written *before* the move (or `task_complete` must take `commit` and write it in the same call — which is what BIT-47 promises). Note BIT-46 moves storage to JSON+md, so re-check these paths after it lands.
- `task/task.go:19-28`: `Task` has no `branch`/`commit` yet (BIT-46 adds them empty).

## Who triggers completion today
`bit/skills/do/SKILL.md:95` (rollup never sets track `done`), `:102-110` (Track sign-off): when the last bar is done, do tells the user the track is ready and stops; on sign-off it hands to `/bit:complete`. `:131` repeats "don't declare a track done on its own".

## What changes when completion must run after landing
- Sign-off and completion become two events. At sign-off the work is usually only committed (maybe pushed to a branch / PR open), so the ladder would answer "not landed" every time if it ran there.
- do's sign-off text (`:106-108`) must stop handing straight to `/bit:complete`; it should instead tell the operator to push/merge and run `/bit:complete <track>` afterwards. **Unknown (operator decision):** whether sign-off still sets the track `done` (status) while filing waits for landing, or the track stays `doing` until completion. The complete skill currently does both in one pass (steps 2-4).
- The complete skill's step 2 ("force every bar done without asking") conflicts with the new "partly done → check in with operator" rule: completion must classify bars from git *before* forcing anything.
- Nothing in code triggers completion automatically; after the change it is still operator-invoked (or offered by bit:bot when it sees a track whose bars are all done). No hook/daemon exists for it.

**Verdict:** confirmed the stub's claim; completion is pure file moves at sign-off with no git. Touches should include `bit/skills/complete/SKILL.md`, `bit/skills/do/SKILL.md:102-110`, `cmd/serve_mcp.go:160-162,417-430`, `task/store.go:108-147`, plus `bit/agents/bot-dev.md` / `bit/agents/bot.md` if they mention sign-off.
