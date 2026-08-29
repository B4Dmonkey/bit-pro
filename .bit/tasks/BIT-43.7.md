---
id: BIT-43.7
title: Both agents drive the typed surface
status: todo
phase: 3
phase_label: Do + agents
---
## **Verse 3**

Both agents are thin — they orient and route rather than drive a long procedure — but they are the
*dispatch* case the whole track exists for: a `bit:bot-dev` session spawned into a worktree has no
operator watching it choose `mv` over the typed surface. This bar closes Verse 3 by pushing both
agents and `bit:do`.

No RED step, for the reason recorded on Verse 1's bars. Measured counts: `bot.md` **4** → **0**,
`bot-dev.md` **1** → **0**.

## Scope
- `bit/agents/bot.md`:
  - line 25 — *"Run `bp instructions` before your first `bp` write in a session … it is the single
    source of truth. Don't work from remembered flags; they drift."* **Delete this instruction
    outright.** Its whole purpose was to defeat flag drift, and a schema in the tool list cannot
    drift — the parameters are enumerated at call time. Replace it with a short pointer that the
    `mcp__bit__*` tools are the write surface.
  - line 27 — the read-only orientation shortcut, `bp task list` / `bp task list --parent <TRACK>`
    → `mcp__bit__task_list` with and without `parent`
  - line 44 — *"flipping a track to `done` and filing it with `bp task complete` is the user's
    sign-off, not yours"* → cite `mcp__bit__task_complete`. The boundary it draws is unchanged.
  - line 60 — the plan check's `bp task list` → `mcp__bit__task_list`
- `bit/agents/bot-dev.md`:
  - line 45 — *"Don't set the track `done` and don't run `bp task complete`"* → cite
    `mcp__bit__task_complete`
  - line 44 — *"Never run `bp approve`"* **stays a CLI reference and must not change.** There is no
    approve tool by design; this agent clearing its own gate is exactly what the line forbids, and
    naming the CLI command is what makes the prohibition concrete. It does not match the sweep
    pattern, so it needs no exception note.

## References
- `mcp-notes.md` — under "Relationship to the automation phase": *"A `bit:bot-dev` session spawned
  by the daemon into a worktree has no operator watching it choose `mv` over `bp task move`."* That
  is this bar's WHY.

## Migration
1. **Invoke skill-creator** with both agent files and the edits above.
2. **Present the draft** to the user for accept/reject.
3. **Close the verse:** push `bit/skills/do/SKILL.md`, `bit/agents/bot.md` and
   `bit/agents/bot-dev.md` to GitHub.

## Claude verifies
- [ ] `PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'` then
      `grep -cE "$PAT" bit/agents/bot.md` — baseline **4**, must reach **0**
- [ ] `grep -cE "$PAT" bit/agents/bot-dev.md` — baseline **1**, must reach **0**
- [ ] `grep -n 'bp approve' bit/agents/bot-dev.md` still finds the prohibition
- [ ] `claude plugin validate ./bit` passes

## User verifies
- [ ] Push to GitHub, then `claude plugin marketplace update bit-pro` and
      `claude plugin update bit@bit-pro --scope project`. Both succeed.
- [ ] **Whole slice:** with the refreshed plugin, run `/bit:do` on a real approved bar. The
      approval gate reads correctly, the bar moves `doing` → `done`, the track rolls up, and the
      transcript shows only `mcp__bit__*` calls — no `bp` Bash invocation. That is the verse's
      capability: the execution end of the pipeline runs without the shell.

## Commit (user)
`refactor(agents): drive bit:bot and bit:bot-dev through MCP tools`