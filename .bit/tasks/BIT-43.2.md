---
id: BIT-43.2
title: bit:feedback records through feedback_add
status: todo
phase: 1
phase_label: Narrow skills
---
## **Verse 1**

`bit:feedback` adds exactly one write — `feedback_add` — to the read surface Verse 1's first bar
proved. It is the smallest write path in the pipeline, which is why it goes before any skill that
mints or updates tasks. This bar also closes the verse: it pushes both rewritten skills and
refreshes the plugin, which is what makes them real anywhere.

No RED step, for the reason recorded on the previous bar. The measured count: this file carries
four `bp` driving references today and must carry zero when the bar is done.

## Scope
- `bit/skills/feedback/SKILL.md` — four driving references become tool calls:
  - line 12 — the "run `bp instructions`" opener → names `mcp__bit__task_list`,
    `mcp__bit__task_read` and `mcp__bit__feedback_add`. The rule it enforces ("never hand-write a
    file under `.bit/feedback/`") stays and gets *stronger*, not weaker: the tool is now the only
    surface that writes notes.
  - line 30 — `bp task list` (disambiguating which track a note belongs to) → `mcp__bit__task_list`
  - line 104 — the fenced `bp feedback add "$TRACK" -d "$(cat note.md)"` block → a
    `mcp__bit__feedback_add` call with `track` and `body`. **The temp-file dance disappears
    entirely**: `body` is a JSON string, so there is no `note.md`, no `$(cat …)`, and no quoting
    hazard. Say what the tool returns — `{path}` — where the old text said "prints the note's path".
  - line 117 — the "hand-write" prohibition citing `bp feedback add` → cites the tool

## References
- `mcp-notes.md` — "Parity map": `feedback_add` takes `track` and `body` and returns `{path}`. Its
  note under the table on whole-body writes explains why passing prose as a JSON string removes the
  file round-trip.

## Migration
1. **Invoke skill-creator** with `bit/skills/feedback/SKILL.md` and the four edits above.
2. **Delete the body-authoring section's shell mechanics.** Building the note in a file existed
   only to survive shell quoting; with a JSON string parameter the guidance is simply to write the
   note's prose.
3. **Present the draft** to the user for accept/reject.
4. **Close the verse:** push `bit/skills/retro/SKILL.md` and `bit/skills/feedback/SKILL.md` to
   GitHub. The plugin installs from the default branch, so an unpushed edit reaches no project —
   including this one.

## Claude verifies
- [ ] `PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'` then
      `grep -cE "$PAT" bit/skills/feedback/SKILL.md` — baseline today is **4**, must reach **0**
- [ ] `claude plugin validate ./bit` passes
- [ ] `quick_validate.py bit/skills/feedback/` passes (kebab-case complaint is known noise)

## User verifies
- [ ] Push to GitHub, then `claude plugin marketplace update bit-pro` followed by
      `claude plugin update bit@bit-pro --scope project`. Both succeed.
- [ ] **Whole slice:** in a project with the refreshed plugin, run `/bit:feedback` on a real
      correction. The note lands in `.bit/feedback/` and the transcript shows a
      `mcp__bit__feedback_add` call — no `bp` Bash invocation anywhere in the run. That is the
      verse's capability: the narrow skills work end to end without the shell.

## Commit (user)
`refactor(skills): drive bit:feedback through MCP tools`