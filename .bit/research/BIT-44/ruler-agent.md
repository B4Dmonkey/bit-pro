# ruler-agent (Verse 3)

**Checked:** the Touches pointers `bit/agents/ruler.md` (new) and `bot.md`, the subagent-nesting Decision, how the stub track gets created, and how the gate would work.

## Agents dir and naming
- `bit/agents/` holds `bot.md` and `bot-dev.md`. Their frontmatter has only `name` and `description`: no `tools`, `model` or `skills`, so they inherit every tool.
- The plugin name `bit` comes from `bit/.claude-plugin/plugin.json:4`, and `.claude-plugin/marketplace.json` points at `./bit`.
- The docs say a plugin agent registers as `<plugin>:<name>`, and agents in subfolders add the subfolder to the name. So `bit/agents/ruler.md` with `name: ruler` becomes `bit:ruler` with no other change. `claude/dispatch.go:94` already uses `--agent bit:bot-dev`.
- The docs say plugin agents ignore the `hooks`, `mcpServers` and `permissionMode` frontmatter fields. The ruler can't bring its own MCP server; it depends on `claude mcp add bit -- bp serve mcp` (`claude/sync.go:22`).
- `automation-notes.md:206`: an unknown `--agent` only warns and starts a plain session. If the plugin copy is stale, `bit:ruler` fails without any error.

## Routing table (bot.md:47-55)
- The table routes to scope, plan, do, check, feedback, retro and learn. **There is no `bit:analyze` row either**, not just no ruler row.
- `bot.md:22` calls `task_create`/`task_update`/`task_move`/`task_complete`/`feedback_add` "the whole write surface". That's stale now that `research_write` exists.
- The Touches pointer for bot.md is **confirmed**. The edit should also add analyze and fix :22.
- Bars already exist: BIT-44.7 (ruler) and BIT-44.8 (bot routes to ruler). Both are `todo` and approved.

## Nesting Decision ("subagents can nest up to three levels"): CONFIRMED, with conditions
- The docs (code.claude.com/docs/en/sub-agents) say: "By default, a subagent can spawn subagents of its own, up to three layers below the main conversation. At the depth limit, Claude Code withholds the `Agent` tool". `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` changes the limit.
- With the ruler as the main thread, the analyze subagent is layer 1 and its explorers are layer 2, which is inside the limit.
- **Conditions the scope doesn't state:**
  - The analyze subagent must be an agent type that has the Agent tool, so not built-in `Explore`. If it's dispatched as `Explore`, the explorer fan-out is silently lost.
  - The analyze subagent also needs the Skill tool, or `skills: [analyze]` preloaded, plus the `mcp__bit__*` tools.
  - The docs say leaving out `tools` inherits built-in and MCP tools. Options:
    - a `general-purpose` subagent told to run `/bit:analyze`
    - a dedicated plugin agent (e.g. `bit/agents/analyst.md`) with `skills:` preloading analyze
  - The second option would be a new file that isn't in the Touches list. It's for scope to decide.
- No existing agent or skill dispatches a skill into an in-session subagent. `bot-dev.md:10` runs `bit:do` inline. The only fresh-context dispatch is a separate process (`claude --bg --agent bit:bot-dev`, `claude/dispatch.go:94`). The ruler would be the first to do it, so there's no pattern to copy.

## Stub track and the gate
- `task_create` (`cmd/serve_mcp.go:138-145`, handler :354-376, which calls `store.Create`, `task/store.go:205-254`) does no validation. A title with an empty body works, and so would an empty title. The new track is `todo` and unapproved. The stub-track Decision is **confirmed as feasible**.
- **No MCP tool sets `approved`.**
  - Only `bp approve`/`bp unapprove` (`cmd/approve.go:9-29`) or the TUI can set it.
  - `task_update` only *clears* it: changing the title, body, phase or phase_label, or setting status to `todo`, clears approval (serve_mcp.go:54-63).
  - The scope-then-analyze loop rewrites the body, so any approval is cleared each round.
  - The gate therefore has to be a conversational yes from the operator. The ruler can read `approved` to see whether the operator ran `bp approve`, but it can't set it.
  - That matches bit:do and bot-dev, which treat approving as the operator's act.
  - **Unknown for scope:** is the gate "the operator says go" or "the track reads approved"?

## Left unsettled by the RFC (`rfc-plan-orchestrator.md`)
The RFC's ordering has been replaced (gates after analyze, scope and plan, then hand off to do). Three things it raised have no Decision:
- **Loop stop condition.** The RFC suggests "all load-bearing unknowns resolved, all touch-points verified". The Decisions only say "operator approves".
- **What the ruler shows at the gate.** The RFC has analyze show a diff of resolved unknowns. The Decisions forbid analyze from editing the track, so the summary has to come from the analyze report plus scope's changes.
- **The form open questions take when handed to the next analyze pass.** The Decisions say "track body + open questions", with no format.

## Docs to update (not in Touches)
- `README.md:80-89` says the plugin "ships seven skills". It leaves out analyze and doesn't mention agents.
- `bit/.claude-plugin/plugin.json` description reads "scope → plan → do". Cosmetic.
