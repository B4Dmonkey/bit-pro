---
id: BIT-44
title: 'Deep research before scope: bit:analyze + bit:ruler'
status: doing
---
## Why
bit:scope only does light research on purpose, and bit:plan says outright that it isn't a discovery phase. So nobody does the deep codebase research, and wrong assumptions only show up during bit:do or come back as plan-to-scope hand-backs. The operator also has to remember the scope → plan sequence and run each skill by hand.

## Summary
Add a `bit:analyze` skill that does the deep research and writes its findings as notes in `.bit/research/`, and a `bit:ruler` agent that becomes the operator's entry point for planning work. The operator describes the work and the ruler runs analyze → scope. The ruler loops analyze ↔ scope until the operator approves the scope, then moves on to plan. The scope stays short because the evidence lives in the research notes, not in the track body.

## Visual aid
```
claude --agent bit:ruler
  │  operator describes the work
  ├─ stub track created (so research has a home)
  ├─ analyze  ── fresh subagent ──▶ fans out explorers ──▶ .bit/research/BIT-N/
  ├─ scope    ── reads research index, writes a short track body
  ├─ GATE     ── operator reviews scope ── questions? ──▶ back to analyze (fresh)
  └─ plan     ── bit:plan as today, then stop
```

## Decisions
- **The MCP exposes both reading and writing research notes,** so agents (analyze, scope, the ruler) write and read research through the tools and never hand-edit `.bit/`.
- **Research is MCP-only, through two tools: `research_write` and `research_read`.** There is no `bp research` CLI command. The names follow the research domain and the existing `feedback_add` naming.
- **`research_write` takes a track, a topic, and a body, and overwrites.** Rewriting a topic replaces its file, and there's no delete. Research is a scratchpad for agents, so nothing needs history. The first write creates `.bit/research/<track>/`, the way `feedback_add` creates `.bit/feedback/`. `bp init` doesn't change.
- **The index is just the topic named `index`.** It is written with the same tool. bp doesn't generate or maintain it.
- **`research_read` with only a track lists that track's topic names. With a topic, it returns that topic's body.** This follows the `view` command in Anthropic's memory tool (a directory lists its files, a file returns its contents), so an agent sees what exists before loading it. A track with no research yet returns an empty list, not an error.
- **Names that look like paths are refused; other topic names are cleaned.** A topic or track ID containing `..`, `/` or `\` is refused with an error that names the problem, and so is a topic made only of dots and spaces (including an empty one on `research_write`). Any other topic is cleaned by pathologize into a single safe file name, with leading dots stripped (`.hidden` becomes `hidden.md`). On `research_read`, an empty topic isn't refused: it lists the track's topics.
- **Research keys to a track the same way feedback does.** The track has to exist (active, completed, or archived), which catches a mistyped ID.
- **Research is one folder per track, disclosed progressively.** `.bit/research/BIT-N/index.md` has a short summary of findings with links, and each topic gets its own file next to it. A reader or a later agent opens the index first and only opens the topics it needs. This follows Anthropic's context-engineering guidance on structured note-taking and progressive disclosure.
- **Research notes are for the agent, not the operator.** Their format is loose, and the operator doesn't review them.
- **Every analyze pass runs in a fresh subagent.** The ruler dispatches it with the track body and the open questions, not the conclusions of earlier notes. It rewrites the topic files it re-checks. Subagents can nest up to three levels, so the analyze subagent can still fan out explorers.
- **The ruler runs as the main session agent** (`claude --agent bit:ruler`), the same way `bit:bot` does.
- **The ruler creates a stub track before analyzing,** so the research has a track ID from the start. bit:scope then fills in the body.
- **There is one human gate, between scope and plan.** Analyze flows straight into scope without a stop. Plan keeps its own review as it works today. The ruler stops after plan, and bit:do stays a separate step.
- **The scope stays short.** bit:scope reads the research index when there is one and cites it instead of copying the evidence into the track body.
- **`bit:analyze` is written with skill-creator** and ships through the plugin (`bit/skills/`), like the other bit skills.

## Verses
- [x] Verse 1 — The operator can run `/bit:analyze BIT-N` on a track and get deep-research notes in `.bit/research/BIT-N/`, which agents read and write through the MCP.
  Touches: `bit/skills/analyze/` (new), `cmd/serve_mcp.go`, `task/feedback.go` (the pattern `feedback_add` uses).
- [x] Verse 2 — bit:scope builds on existing research, so a scope written after analyze stays short and cites the notes.
  Touches: `bit/skills/scope/SKILL.md`.
- [x] Verse 3 — The operator runs `claude --agent bit:ruler`, describes the work, and gets analyze → scope with the analyze ↔ scope loop and a gate before plan.
  Touches: `bit/agents/ruler.md` (new), `bit/agents/bot.md` (routing table).

## References
- `rfc-plan-orchestrator.md`: the original RFC for bit:analyze and the orchestrator (all verses). Its ordering and gates are replaced by the Decisions above.
- https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents: the source for structured note-taking and progressive disclosure (Verse 1 research layout).
- https://platform.claude.com/docs/en/agents-and-tools/tool-use/memory-tool: the list-then-read shape `research_read` follows (Verse 1).
- https://code.claude.com/docs/en/sub-agents: nested subagents, the `skills` preload field, and `--agent` (Verses 1 and 3).