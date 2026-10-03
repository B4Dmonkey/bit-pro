# Natural verse order (Q6)

Dependencies found in the code:
- The commit skill needs somewhere to put the hash: `task_update` must accept `commit`/`branch` (`cmd/serve_mcp.go:151-158`, `task/store.go:260-303`). Without it the skill can commit but not record.
- The do/bot-dev reorder needs the commit skill to call.
- Write-time capture on research/feedback is independent of the other three (different handlers, `cmd/serve_mcp.go:449-481`), but needs BIT-49's git helper and BIT-46's `commits` field.

Suggested walking skeleton:
1. **Thin end-to-end:** `task_update` takes `commit`/`branch` (not revoking approval) + minimal `bit:commit` skill (ask, `git commit`, read sha, `task_update` bar with commit+branch+done) + bit:do's Verified good reordered to use it. One bar goes commit → hash → done in a real session. This is the usable slice; splitting the MCP input off as its own verse delivers nothing a user sees.
2. **bot-dev rewrite:** asks before every commit, uses the skill; decide push.
3. **Sweep the "user commits" text** in plan (`:304`, `## Commit (user)` `:361,:398`), complete (`:30`), check (`:34,:38`), do description. Could fold into 1/2.
4. **Write-time capture** on research and feedback (and retro once BIT-49 adds it). Independent; could go first or in parallel if 1 stalls.
- `task_complete` commit/branch: mostly BIT-50's use; can ride with 1 (cheap) or be left to BIT-50.
