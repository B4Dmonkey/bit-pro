---
id: BIT-43.3
title: bit:scope authors tracks through tools
status: todo
phase: 2
phase_label: Scope, check, plan
---
## **Verse 2**

`bit:scope` is the first skill that *mints* a task. It uses only `task_create` and `task_update`
with no approval logic and no rollup, which is what makes it the entry point to the write surface.

No RED step, for the reason recorded on Verse 1's bars. Measured count: **5** matching references
today, **0** when done.

## Scope
- `bit/skills/scope/SKILL.md`:
  - line 20 — the "run `bp instructions`" opener → names `mcp__bit__task_create`,
    `mcp__bit__task_read` and `mcp__bit__task_update`. The "never hand-edit `.bit/tasks/*.md`" rule
    stays; it is the rule the whole surface exists to enforce.
  - line 27 — refine reads the body with `bp task read <id> --body` → `mcp__bit__task_read`, taking
    the `body` field
  - line 127 — the prose "`task create` prints the new track ID" → the tool *returns* `{id}`.
    "Prints" is a shell artifact; a structured return is not printed.
  - line 130 — the fenced `TRACK=$(bp task create "<scope title>" -d "$(cat scope-body.md)")` → a
    `mcp__bit__task_create` call with `title` and `body`. **The `scope-body.md` temp file
    disappears**: `body` is a JSON string, so the scope prose is passed directly. Drop the ID-capture
    guidance with it — there is nothing to capture from, the return carries the ID.
  - line 133 — `bp task update <id> -d "…"` → `mcp__bit__task_update` with `body`

## References
- `mcp-notes.md` — "Parity map": `task_create` takes `title, body?, parent?, after?, phase?,
  phase_label?` and returns `{id}`; `task_update` takes optional fields and returns `{id, approved}`.

## Migration
1. **Invoke skill-creator** with `bit/skills/scope/SKILL.md` and the five edits above.
2. **Delete the "writing a body from the shell" mechanics** this skill inherited from the contract —
   the `-d "$(cat …)"` idiom, the round-trip caveat, the trailing-newline note. None of it survives
   a JSON string parameter.
3. **Present the draft** to the user for accept/reject.

## Claude verifies
- [ ] `PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'` then
      `grep -cE "$PAT" bit/skills/scope/SKILL.md` — baseline **5**, must reach **0**
- [ ] `claude plugin validate ./bit` passes
- [ ] `quick_validate.py bit/skills/scope/` passes (kebab-case complaint is known noise)

## User verifies
- [ ] none — deterministic. Verse 2's capability isn't exercisable until its last bar pushes.

## Commit (user)
`refactor(skills): drive bit:scope through MCP tools`