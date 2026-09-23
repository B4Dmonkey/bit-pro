# BIT-44 research index

First pass, 2026-09-23. No open questions were handed over, so the questions came from the scope's Decisions and Touches.

## Findings that change the scope
- **Path escape (bug):** the research tools let a track ID escape `.bit/research/`. `researchDir` joins the raw track with `filepath.Join`, but the existence check sanitizes it, so `../../BIT-44` passes the check and then writes to `<project>/BIT-44/`. → [mcp-research-tools](mcp-research-tools.md)
- **Decision contradicts the code:** "Topic names have no required format … instead of refusing it" doesn't match the code, which refuses `/`, `\`, `..`, and topics that are only dots. Topics that differ can also collide on one file (`a:b`/`ab`, and case on macOS). → [mcp-research-tools](mcp-research-tools.md)
- **A bar ID is accepted as a track** by the research tools. → [mcp-research-tools](mcp-research-tools.md)
- **The Why is wrong about which skill says what:** "not a discovery phase" is in the *scope* skill (:100), not the plan skill. Plan :97 says "go deeper here", and scope :116 hands the deep dive to plan. → [scope-skill](scope-skill.md)
- **The nesting Decision holds only under conditions:** three levels is confirmed by the docs, but the analyze subagent must not be `Explore`, which has no Agent tool. It needs the Skill tool (or `skills:` preload) and the MCP tools. No existing pattern dispatches a skill into a subagent. → [ruler-agent](ruler-agent.md)
- **No MCP tool can approve**, and scope rewrites clear approval. The gate has to be conversational. → [ruler-agent](ruler-agent.md)

## Pointer corrections
- **Verse 1:** `task/research.go` is the real implementation; the Touches list only names `task/feedback.go`, which supplies `trackExists` and nothing else. → [mcp-research-tools](mcp-research-tools.md)
- **Verse 2:** the scope skill's References section (:186-191) only allows external artifacts, so it needs a slot for citing research. → [scope-skill](scope-skill.md)
- **Verse 3:** bot.md also has no `bit:analyze` row, and its :22 "whole write surface" leaves out `research_write`. `README.md:80-89` ("seven skills") is stale. → [ruler-agent](ruler-agent.md)

## New unknowns (for scope)
- Should bit:plan (:97) read research too? No verse touches `plan/SKILL.md`. → [scope-skill](scope-skill.md)
- There's no stop condition for the analyze↔scope loop, no format for passing questions to the next analyze pass, and nothing says what the gate shows. → [ruler-agent](ruler-agent.md)
- Should analyze run in a dedicated agent file (with `skills:` preload) or in a `general-purpose` subagent? → [ruler-agent](ruler-agent.md)

## Confirmed
- Track existence covers active, completed and archived tracks.
- Writes overwrite; there's no delete; the folder is created on the first write.
- There's no `bp research`, and `bp init` is unchanged.
- The plugin picks up skills and agents from their directories, and `ruler.md` would register as `bit:ruler`.
- Evidence: [mcp-research-tools](mcp-research-tools.md), [ruler-agent](ruler-agent.md).

## Topics
- [mcp-research-tools](mcp-research-tools.md): Verse 1 Go code. Where it lives, each Decision checked, the bugs, and test gaps.
- [scope-skill](scope-skill.md): Verse 2. Which lines of the scope skill to change, and the scope/plan contradiction.
- [ruler-agent](ruler-agent.md): Verse 3. Agent naming, nesting conditions, the stub track, the gate, and gaps left by the RFC.
