---
id: BIT-44.7
title: bit:ruler runs analyze, scope, gate, plan
status: done
approved: true
phase: 3
phase_label: Ruler
---
## **Verse 3**

The operator gets one entry point for planning. Today they have to remember scope → plan and run each skill by hand, so the agent file itself is the deliverable.

## Scope
- `bit/agents/ruler.md` (new): frontmatter `name: ruler` plus a `description` (same shape as `bot.md`; no `hooks`, `mcpServers`, or `permissionMode`, which plugin agents ignore). The body carries the flow from the track's Decisions and Visual aid:
  1. The operator describes the work. Create a **stub track** with `mcp__bit__task_create` (title from the description, body the operator's words) so research has an ID.
  2. **Analyze in a fresh subagent:** dispatch with the Agent tool, telling it to run `bit:analyze` on the track and passing only the track ID, the track body, and any open questions, never earlier notes' conclusions.
  3. Run `bit:scope` on the track straight away, with no stop between analyze and scope.
  4. **Gate:** show the scope and wait. If the operator has questions, go back to step 2 with those questions (a new fresh subagent), then refine the scope again. Only explicit approval moves on.
  5. Run `bit:plan` on the track (it keeps its own review), then **stop**. Never run `bit:do`.

## References
- https://code.claude.com/docs/en/sub-agents: `--agent` as main session, nesting depth (3 by default, so analyze can still fan out), and the plugin agent field restrictions.
- `bit/agents/bot.md`: tone and structure of an existing main-session bit agent.

## Method
- [ ] Write `ruler.md` covering steps 1–5 above.

## Claude verifies
- [ ] `claude plugin validate ./bit`

## User verifies
- none. The verse check is on the next bar.

## Commit (user)
`feat(bit): add bit:ruler planning agent`