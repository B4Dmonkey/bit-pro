---
id: BIT-43.5
title: bit:plan mints bars through task_create
status: done
approved: true
phase: 2
phase_label: Scope, check, plan
---
## **Verse 2**

`bit:plan` is the heaviest user of `task_create` — it mints every bar in a plan, in order, each
with a multi-line body. It also owns `task_move`, which no other skill uses. This bar closes
Verse 2 by pushing all three rewritten skills and refreshing the plugin.

No RED step, for the reason recorded on Verse 1's bars. Measured count: **9** matching references
today, **1** when done — see the documented exception below.

## Scope
- `bit/skills/plan/SKILL.md`:
  - line 10 — the "run `bp instructions`" opener → names `mcp__bit__task_read`,
    `mcp__bit__task_list`, `mcp__bit__task_create`, `mcp__bit__task_update`, `mcp__bit__task_move`
  - line 21 — reading the parent track for the WHY, `bp task read <track> --body` → `mcp__bit__task_read`
  - line 73 — "read its body end to end with `bp task read <track> --body`" → `mcp__bit__task_read`
  - line 310 — "`task create --parent` prints the dotted bar ID" → the tool *returns* `{id}`
  - line 313 — the fenced `BAR=$(bp task create … -d "$(cat step-body.md)")` block → a
    `mcp__bit__task_create` call with `title`, `parent`, `phase`, `phase_label`, `body`. **The
    `step-body.md` temp file and the `$( )` ID capture both disappear.** While rewriting this
    block, add `after` — it is a real `task_create` parameter that inserts a bar mid-track at
    create time, and the retired contract never documented it, so no skill knows it exists.
  - line 410 — refine reads via `` `task read <track> --body` `` and `` `task list --parent <track>` ``
    → the tools
  - line 411 — re-listing every bar in a track with `` `bp task list --parent <track>` `` → `mcp__bit__task_list`
  - line 421 — refine's edit verbs: `` `task update <bar> -d "…"` `` → `mcp__bit__task_update` with
    `body`; `` `task create --parent …` `` → `mcp__bit__task_create`; `` `task move <bar>
    --before|--after <sibling>` `` → `mcp__bit__task_move` with `bar` and exactly one of
    `before`/`after`. The stable-identity reasoning around `task_move` stays — it is domain, not
    shell technique.

**One reference deliberately survives.** Line 246 — *"Run `bp task list` against the real records —
the 13 bars under BIT-2 stay in one column, nothing wraps"* — is an *example of a good User-verify
check*, i.e. a command the human runs in their own terminal. The CLI is still the operator's
surface, so this stays exactly as written. Line 244's `bp tui` example is the same case. That is
why this file's target is 1 and not 0.

## References
- `mcp-notes.md` — "Parity map" and its note *"`create --after` is a real flag that the contract
  never documents"*, which is the source of the `after` addition above.

## Migration
1. **Invoke skill-creator** with `bit/skills/plan/SKILL.md` and the nine edits above.
2. **Update the plan format section** so a bar body is authored as a `body` string, not built in a
   file. Every `-d "$(cat …)"` instruction goes.
3. **Present the draft** to the user for accept/reject.
4. **Close the verse:** push `bit/skills/scope/SKILL.md`, `bit/skills/check/SKILL.md` and
   `bit/skills/plan/SKILL.md` to GitHub.

## Claude verifies
- [ ] `PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'` then
      `grep -cE "$PAT" bit/skills/plan/SKILL.md` — baseline **9**, must reach **1**
- [ ] `grep -nE "$PAT" bit/skills/plan/SKILL.md` — the single remaining hit is the User-verify
      example line, not a driving instruction
- [ ] `claude plugin validate ./bit` passes
- [ ] `quick_validate.py bit/skills/plan/` passes (kebab-case complaint is known noise)

## User verifies
- [ ] Push to GitHub, then `claude plugin marketplace update bit-pro` and
      `claude plugin update bit@bit-pro --scope project`. Both succeed.
- [ ] **Whole slice:** with the refreshed plugin, run `/bit:scope` on a throwaway idea and then
      `/bit:plan` on the track it produced. A track and its bars land in `.bit/`, and the transcript
      shows `mcp__bit__task_create` calls with no `bp` Bash invocation. That is the verse's
      capability: authoring a scope and its plan without the shell.

## Commit (user)
`refactor(skills): drive bit:plan through MCP tools`