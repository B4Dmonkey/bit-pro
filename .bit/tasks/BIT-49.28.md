---
id: BIT-49.28
title: The agents and the plugin description no longer name .bit/
status: todo
phase: 3
phase_label: no .bit/ in sight
---
## **Verse 3**

bot, ruler and the plugin manifest still describe bit as "a project with a `.bit/` directory" reached "through the bp CLI". This is a wording sweep. bot-dev's staging lines are BIT-49.29's.

## Scope
Re-grep first: `grep -n '\.bit\|bp CLI' bit/agents/*.md bit/.claude-plugin/plugin.json`.
- `bit/agents/bot.md`:
  - `:3` (description): "any project with a `.bit/` directory" → "any project registered with bit (`bp add` or `bp migrate`)".
  - `:8` and `:16`: name the `mcp__bit__*` tools instead of `.bit/`.
  - `:22`: "Never hand-edit `.bit/tasks/*.md`" becomes the scope's sentence ("the `mcp__bit__*` tools are the only way in"). Keep the write-surface list, which already has `retro_write` from BIT-49.8.
- `bit/agents/ruler.md:10`, `:12`: the same rule sentence. Research "lives in the store under the track", reached through `mcp__bit__research_*`.
- `bit/.claude-plugin/plugin.json:5`: drop `.bit/` and "through the bp CLI". The skills work through the `bit` MCP tools.
- `bit/agents/bot-dev.md` is left alone (BIT-49.29 for `:27-28`; BIT-47 rewrites the rest).
- Text only, with no test files.

## Change checklist
- [ ] Apply the edits.
- [ ] `grep -n '\.bit' bit/agents/bot.md bit/agents/ruler.md bit/.claude-plugin/plugin.json` finds nothing.
- [ ] `grep -n 'bp CLI' bit/.claude-plugin/plugin.json` finds nothing.

## Claude verifies
- [ ] `claude plugin validate ./bit` passes
- [ ] `python3 -m json.tool bit/.claude-plugin/plugin.json > /dev/null`
- [ ] `just lint` and `just test` pass

## User verifies
- none. The verse's end-to-end run is on BIT-49.30.

## Commit (user)
`docs(bit): agents and plugin description drop .bit/`