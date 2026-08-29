---
id: BIT-43
title: Skills walk the MCP surface
status: todo
approved: true
---
## Why

Claude reaches `bp` through Bash — seven skills and two agents shell out to `bp task list`,
`bp task read`, `bp instructions`, and so on, as if the only interface were a terminal. The
MCP server (steps 1–4 done: skeleton, read surface, write surface, registration) already
exposes every one of those operations as typed, schema-validated tools — but no skill uses
them. The consequence is visible in the skills themselves: a third of each one's instruction
budget goes to shell technique (ID capture with `$( )`, body quoting, tab-column counting,
status spelling warnings) that schemas and structured returns make irrelevant. Meanwhile the
dispatch case — an unattended `bit:bot-dev` session with no operator watching — reaches for
`mv` and `sed` against `.bit/` because those paths are wide open and the typed surface is
absent from its habits. This verse closes that gap: skills call tools, the shell technique
disappears, and `bp instructions` and its 158 lines of tutorial can retire.

## Summary

Rewrite all seven skills and both agents to call MCP tools (`mcp__bit__task_read`,
`mcp__bit__task_list`, etc.) instead of shelling out, migrating in order of complexity.
Once all skills are off `bp instructions`, move its domain knowledge into the tool
descriptions and retire the command.

## Decisions

- **MCP tool calls are proven end-to-end.** `cmd/mcp_harness_test.go` exercises the full
  protocol round-trip via `mcp.NewInMemoryTransports()` — `CallTool` on `task_read`,
  `task_list`, all write tools. The stdio transport uses the same protocol path.
- **`bp instructions` stays alive through Verses 1–3.** Skills still call it during the
  migration so no in-flight session loses the domain it teaches. It retires only once every
  skill has been rewritten off it.
- **Domain rides the tool descriptions, not a `get_instructions` tool.** Each tool carries
  the concepts relevant to its own use; no separate dispatch needed.
- **Version tag cut before any skill edit.** Done — the tag is on the pre-migration commit.
- **Additive: bash path stays open until Verse 5's real cycle runs.** Step 6 (deny rules)
  is not this track's work.
- **Migrate simplest-to-hardest.** retro + feedback (read-only / one write) → scope + check
  + plan (write surface, no approval logic) → do + agents (approval gate, rollup, complete).
- **Every skill edit goes through skill-creator and ships via GitHub.** A local edit that
  hasn't been pushed doesn't count as done — the plugin installs from the default branch.

## Verses

- [ ] Verse 1 — Narrow skills migrate (bit:retro + bit:feedback): `bit:retro` is read-only
  (task_list, task_read); `bit:feedback` adds feedback_add. Both rewritten via skill-creator,
  pushed to GitHub, and the plugin refreshed before claiming done.
  Touches: `bit/skills/retro/SKILL.md`, `bit/skills/feedback/SKILL.md`.

- [ ] Verse 2 — Scope, check, plan migrate: `bit:scope` (task_create + task_update),
  `bit:check` (reads + occasional status-cleanup update), `bit:plan` (many task_create
  calls). Write surface, no approval complexity. Each rewritten via skill-creator and pushed.
  Touches: `bit/skills/scope/SKILL.md`, `bit/skills/check/SKILL.md`,
  `bit/skills/plan/SKILL.md`.

- [ ] Verse 3 — Do + agents migrate: `bit:do` (approval gate, status rollup, task_complete)
  and both agents (`bot.md`, `bot-dev.md`). Most complex surface; rewritten via skill-creator
  and pushed.
  Touches: `bit/skills/do/SKILL.md`, `bit/agents/bot.md`, `bit/agents/bot-dev.md`.

- [ ] Verse 4 — Domain lands on the tool descriptions and `bp instructions` retires: the
  description constants in `cmd/serve_mcp.go` carry track-vs-bar, approval semantics,
  rollup ownership, and ID reservation — everything the command currently teaches. Once this
  lands and `just install` runs, `bp instructions` can be removed from the CLI.
  Touches: `cmd/serve_mcp.go` (description constants, particularly `taskReadDescription`);
  `assets/bit-cli.md` and the `instructions` command retire.

- [ ] Verse 5 — Full cycle without bash: push all Verse 4 changes, install the plugin in
  `tools/example`, run a complete scope → plan → do pass. Done when no skill or agent
  contains a `bp task`, `bp feedback`, or `bp instructions` bash block, and the cycle
  produces correct output.
  Touches: GitHub push, `claude plugin update` in `tools/example`, a real pipeline run.

## References

- `mcp-notes.md` — MCP phase working notes; steps 1–4 done, step 5 is this track. Informs
  all verses, particularly the domain-enrichment list and the GitHub-coupling constraint.
- `assets/bit-cli.md` — the 158-line contract being retired in Verse 4. The "domain"
  sections (track vs. bar, approval, rollup, ID reservation, gotchas) inform Verse 4's
  description rewrites; the "shell technique" sections disappear by construction.