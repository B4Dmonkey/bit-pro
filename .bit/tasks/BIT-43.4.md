---
id: BIT-43.4
title: bit:check audits through structured reads
status: todo
phase: 2
phase_label: Scope, check, plan
---
## **Verse 2**

`bit:check` reads a whole track and its bars, and writes only the occasional status cleanup. It is
the widest *read* of any skill, so it is where structured returns pay off most — no tab-column
counting, no `head -1`.

No RED step, for the reason recorded on Verse 1's bars. Measured count: **8** matching references
today, **0** when done.

## Scope
- `bit/skills/check/SKILL.md`:
  - line 12 — the "run `bp instructions`" opener → names the read tools it actually uses
  - line 19 — resolving the track with `bp task list`, then its bars with
    `bp task list --parent <track>` → `mcp__bit__task_list` (omit `parent` for everything, set it to
    the track ID for that track's bars). The "tracks are the rows whose ID has no dot" heuristic
    stays true and gets easier: each row carries a `parent` field, empty for a track.
  - line 32 — `bp task list --parent <track>` → `mcp__bit__task_list`
  - line 34 — `bp task read <bar> --body` → `mcp__bit__task_read`, taking `body`
  - line 38 — the stale-status fix `bp task update <bar> -s done` → `mcp__bit__task_update` with
    `status: "done"`. **Note what the schema now guarantees:** `status` is an enum of
    `todo|doing|done`, so the old "status is stored verbatim — spelling matters" warning is dead
    weight. A typo is rejected by the schema rather than silently stored. Remove any inherited
    spelling caution rather than restating it.
  - line 45 — track/bar sync fix `bp task update <track> -d "…"` → `mcp__bit__task_update` with `body`
  - line 101 — "it stays a plain file rather than a bp task" → reword to "rather than a tracked
    task". This is prose, not a command, but leaving the literal string in defeats the sweep and
    the sentence reads the same without it.
  - line 190 — the recap listing `bp task list` / `bp task read <id> --body` → the tools

The check report itself (`<track-id>-check.md` in the repo root) stays a plain file. It is a
transient audit artifact the retro skill consumes, not tracked work — unchanged by this bar.

## References
- `mcp-notes.md` — "Parity map": `task_list` returns an array of
  `{id, title, status, approved, phase, phase_label, parent}`, so the five-column tab format and its
  "count tabs rather than assuming the phase label is the fourth field" caveat both disappear.

## Migration
1. **Invoke skill-creator** with `bit/skills/check/SKILL.md` and the eight edits above.
2. **Delete the output-parsing guidance.** Column positions, tab counting, `head -1` for a summary
   line — all of it was reading a text format that no longer exists.
3. **Present the draft** to the user for accept/reject.

## Claude verifies
- [ ] `PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'` then
      `grep -cE "$PAT" bit/skills/check/SKILL.md` — baseline **8**, must reach **0**
- [ ] `claude plugin validate ./bit` passes
- [ ] `quick_validate.py bit/skills/check/` passes (kebab-case complaint is known noise)

## User verifies
- [ ] none — deterministic.

## Commit (user)
`refactor(skills): drive bit:check through MCP tools`