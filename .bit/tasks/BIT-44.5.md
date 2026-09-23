---
id: BIT-44.5
title: /bit:analyze writes deep research for a track
status: todo
approved: true
phase: 1
phase_label: Analyze
---
## **Verse 1**

The operator-facing half of the verse: `/bit:analyze BIT-N` does the deep research the scope skips and leaves it in `.bit/research/BIT-N/` through the tools from bars 1–4. This is skill text, not Go, so skill-creator authoring replaces the TDD cycle.

## Scope
- `bit/skills/analyze/SKILL.md` (new): frontmatter `name: bit_analyze`, matching the other bit skills' underscore names. The `analyze/` directory is what makes it `/bit:analyze`.
- `bit/skills/analyze/evals/evals.json` (new) if skill-creator's process produces evals, following the `bit/skills/feedback/evals/` precedent.

## References
- `rfc-plan-orchestrator.md`, "The analysis skill": what analyze does (fan out explorers, validate the scope's assumptions, surface missed unknowns, confirm or correct touches pointers). Its "updates the scope" part is superseded: analyze writes notes, not the track body.
- https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents: structured note-taking and progressive disclosure for the index-plus-topics layout.
- `bit/skills/feedback/SKILL.md`: how an existing bit skill names its `mcp__bit__*` tools and forbids hand-editing `.bit/`.

## Method
- [ ] Author the skill with `skill-creator:skill-creator`. The body has to carry these, each from the track's Decisions:
  - Input: a track ID, plus any open questions the caller hands over. Read the track with `mcp__bit__task_read`.
  - Start with `mcp__bit__research_read` (track only) to see which topics exist. Re-check and **rewrite** the topics the open questions touch; don't carry forward earlier conclusions unchecked.
  - Fan out explorer subagents across the code areas the verses touch. Validate the assumptions the scope rests on, and confirm or correct each touches pointer.
  - Write one topic per area with `mcp__bit__research_write`, then write `index` last: a short summary of findings with a link to each topic.
  - Notes are for agents; the format is loose. Never read or hand-edit `.bit/research/` directly; the tools are the only way in.
  - Don't edit the track body. Shaping the scope is bit:scope's job.
- [ ] Description triggers: "analyze", "research this track", "deep dive before scoping", and `/bit:analyze BIT-N`.

## Claude verifies
- [ ] `uv run --quiet --with pyyaml python ~/.claude/plugins/cache/claude-plugins-official/skill-creator/unknown/skills/skill-creator/scripts/quick_validate.py bit/skills/analyze/` (the kebab-case `name:` failure is known noise)
- [ ] `claude plugin validate ./bit`

## User verifies
- [ ] With bars 1–4 installed (`just install`), start `claude --plugin-dir ./bit` in this repo and run `/bit:analyze BIT-44`. Check that `.bit/research/BIT-44/` holds `index.md` plus at least one topic file, that the index links to those topics, and that `bp task read BIT-44 --body` is unchanged. **Whole slice:** a track now has deep research an agent can list and read through the MCP.

## Commit (user)
`feat(bit): add bit:analyze skill for deep track research`