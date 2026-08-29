---
id: BIT-43.6
title: bit:do runs a bar through tools
status: done
approved: true
phase: 3
phase_label: Do + agents
---
## **Verse 3**

`bit:do` is the most complex surface in the pipeline: the approval gate, the status rollup, and
`task_complete`. It goes last because everything it does depends on reads and writes the earlier
verses already proved.

No RED step, for the reason recorded on Verse 1's bars. Measured count: **14** matching references
today, **0** when done.

## Scope
- `bit/skills/do/SKILL.md`:
  - line 15 — the "run `bp instructions`" opener → names `mcp__bit__task_read`,
    `mcp__bit__task_list`, `mcp__bit__task_update`, `mcp__bit__task_complete`
  - line 25 — finding the track: `bp task read BIT-21 | head -1` and the whole-board `bp task list`
    → `mcp__bit__task_read` and `mcp__bit__task_list`. **`head -1` goes** — a structured return has
    no header line to strip, so "confirms the track exists in a single line" becomes "returns its
    title and status". Keep the ID-uppercasing rule; that is domain, not shell.
  - line 27 — `bp task list --parent <TRACK>` and `bp task read <BAR> --body` → the tools. The
    resume rule (the next bar is the first whose `status` is not `done`) is unchanged; it now reads
    a field instead of a column.
  - line 29 — `bp task read <TRACK> --body` for the WHY → `mcp__bit__task_read`
  - **lines 33, 37 — the approval gate.** Today it reads *"column 4, between the title and the
    phase label … the column is empty when it isn't"*. Replace with the `approved` boolean each
    `task_list` / `task_read` row carries. All the column-position mechanics go.
  - line 35 — the refusal message telling the user `bp approve BIT-X.N`. **This stays a CLI
    command and must not become a tool call.** There is deliberately no approve tool: approval is
    the operator's act, and its absence from the surface is the point. (It does not match the
    sweep pattern, so it needs no exception note.)
  - line 41 — `bp task update <bar> -s doing` and the track roll-up to `doing` →
    `mcp__bit__task_update` with `status`
  - line 75 — the no-User-verifies unwind path, `bp task update <bar> -s doing` → the tool
  - line 91 — marking the bar done, `bp task update <bar> -s done` → the tool
  - line 93 — re-listing bars, `bp task list --parent <track>` → the tool
  - line 96 — the single roll-up call `bp task update <track> -d "<edited body>" -s <status>` →
    one `mcp__bit__task_update` call with `body` and/or `status`. The "skip the call if neither
    changed" rule stays.
  - line 98 — `` `task read <track>` `` → the tool
  - line 110 — `bp task update <track> -s done` → the tool
  - line 111 — `bp task complete <track>` → `mcp__bit__task_complete`. Keep the domain it teaches:
    every bar must be `done`, there is no override, and the relocation into `.bit/completed/` is
    what drops finished work out of the board.

**Two things the schema now enforces that the prose should stop warning about.** `status` is an
enum, so the "spelling matters / a typo silently breaks rollup" caution is obsolete. And
`task_update` returns `approved`, so approval revocation is *observable* rather than a rule the
skill has to remember — the unwind path at line 75 can state what the return shows.

## References
- `mcp-notes.md` — "Parity map" (`task_update` returns `{id, approved}` "so revocation is visible")
  and, under Decisions, *"Absence is stronger than denial — `approve` simply is not a tool"*, which
  is the reasoning behind keeping line 35 a CLI instruction.

## Migration
1. **Invoke skill-creator** with `bit/skills/do/SKILL.md` and the edits above.
2. **Rewrite the approval gate against the boolean**, not a column.
3. **Delete the obsolete cautions** (status spelling, column counting, `head -1`).
4. **Present the draft** to the user for accept/reject.

## Claude verifies
- [ ] `PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'` then
      `grep -cE "$PAT" bit/skills/do/SKILL.md` — baseline **14**, must reach **0**
- [ ] `grep -n 'bp approve' bit/skills/do/SKILL.md` still finds the refusal message — the gate
      still points at the CLI
- [ ] `claude plugin validate ./bit` passes
- [ ] `quick_validate.py bit/skills/do/` passes (kebab-case complaint is known noise)

## User verifies
- [ ] none — deterministic. The verse's end-to-end check lands on its last bar.

## Commit (user)
`refactor(skills): drive bit:do through MCP tools`