---
id: BIT-43.1
title: bit:retro reads through tools, not Bash
status: todo
phase: 1
phase_label: Narrow skills
---
## **Verse 1**

`bit:retro` is the read-only end of the pipeline — it lists tracks, reads bodies, and writes its
proposals as a plain file. Migrating it first proves the read tools carry everything a skill needs
before any write path is touched.

**There is no RED step in this bar, deliberately.** Skill files are markdown with no test harness,
and the decision on this track is to verify migration by grep plus skill-creator validation rather
than to build one. What stands in for red-green is the measured count below: this file carries
three `bp` driving references today and must carry zero when the bar is done.

## Scope
- `bit/skills/retro/SKILL.md` — three driving references become tool calls:
  - line 19 — `bp task read <id> --body` → `mcp__bit__task_read`, whose `body` field *is* the prose
  - line 27 — the "before you drive the CLI, run `bp instructions`" paragraph → names
    `mcp__bit__task_list` and `mcp__bit__task_read` directly. The asymmetry the paragraph already
    teaches (this skill *reads* through the surface but *writes* its proposals as a plain file,
    because no retro tool exists) stays, and gains the reason it is now self-evident: the tool list
    is enumerable, so a missing tool is visible rather than something the reader has to be warned
    about.
  - line 29 — `bp task read <track> --body` → `mcp__bit__task_read`

**Feedback notes stay a filesystem read.** There is no feedback list or read tool — `feedback_add`
is create-only — so `.bit/feedback/*.md` is still globbed and read directly. That is unchanged
behaviour, not an oversight, and the skill already says so. Don't invent a tool call for it.

## References
- `mcp-notes.md` — the "Parity map" table gives each tool's params and return shape. `task_read`
  returns `{id, title, status, approved, phase, phase_label, parent, body}`, so `--body` has no
  analogue and needs none.

## Migration
1. **Invoke skill-creator** (Skill tool) with the target `bit/skills/retro/SKILL.md` and the three
   edits above. Skill prose is its craft — don't hand-edit around it.
2. **Rewrite the three references**, and delete the shell technique that goes with them: no ID
   capture with `$( )`, no `--body` flag, no tab-column counting.
3. **Present the draft to the user** for an explicit accept/reject before it counts as applied.

## Claude verifies
- [ ] The reference sweep reaches zero. This pattern catches both the `bp task …` form and the
      bare backticked `` `task read …` `` form the skills also use:
      ```
      PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'
      grep -cE "$PAT" bit/skills/retro/SKILL.md
      ```
      Baseline today is **3**; this bar must bring it to **0**.
- [ ] `claude plugin validate ./bit` passes
- [ ] `SC=~/.claude/plugins/cache/claude-plugins-official/skill-creator/unknown/skills/skill-creator`
      then `uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/retro/`
      passes. Its "not kebab-case" complaint about the `bit_*` name is known noise — `claude plugin
      validate` is the check that counts.

## User verifies
- [ ] none — deterministic. The rewritten skill isn't exercisable until this verse's last bar
      pushes and refreshes the plugin.

## Commit (user)
`refactor(skills): drive bit:retro through MCP tools`