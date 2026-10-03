# Testing v2 skills + a dev MCP server with --plugin-dir (verified 2026-09-30, Claude Code 2.1.285)

## What was checked
1. Does `claude --plugin-dir ./bit` duplicate/conflict with the installed `bit@bit-pro` in a project that enables it? Which copy wins?
2. Does `--plugin-dir` load cleanly where the plugin isn't enabled?
3. How does it combine with `--mcp-config <file> --strict-mcp-config` for a dev `bp` MCP server?

## Setup facts
- `bit/.claude-plugin/plugin.json`: name `bit`, version `1.3.0`, **no `mcpServers`**; `bit/` has no `.mcp.json`. The plugin ships skills + agents only.
- `bit@bit-pro` is enabled in bit-pro's `.claude/settings.json` (project scope), not in `~/.claude/settings.json`.
- The `bit` MCP server is a **local-scope** entry in `~/.claude.json` (projects bit-pro and a client project): `{"command":"bp","args":["serve","mcp"]}`, resolving `bp` on PATH (`~/go/bin/bp`).
- The v2 branch's `bit/skills/` has `complete`, which installed 1.3.0 (`~/.claude/plugins/cache/bit-pro/bit/1.3.0`) lacks — a natural marker.

## Method
Copied `bit/` to scratchpad, prefixed the `analyze` skill and `bot` agent descriptions with `MARKERV2`, ran `claude -p --output-format stream-json --verbose` and read the `init` event (`plugins`, `slash_commands`, `agents`, `mcp_servers`). Dev bp built with `go build -o <scratch>/bp-dev .`, wrapped in a logging shell script referenced by an `--mcp-config` JSON.

## Results
**Q1 — in bit-pro (plugin enabled): no duplicates; `--plugin-dir` wins.** The `plugins` list held exactly one `bit`, `source: bit@inline`, path = the scratch copy; the cached `bit@bit-pro` was not loaded. `bit:complete` present, `bit:bot` showed `MARKERV2`, and each `bit:*` skill/agent appeared once. (The same-named inline plugin replaces the marketplace one for the session.) Note: `version` is reported from plugin.json (1.3.0), so don't use version to tell copies apart — use the `path`/`source` in the init event.
- Unrelated: legacy user-level skills `bit-check`, `bit-plan`, `bit-retro`, `bit-run`, `bit-scope` also load (non-plugin, hyphen names). They don't collide with `bit:*` but are noise.

**Q2 — in a dir without the plugin: loads cleanly.** Same single `bit@inline` plugin, all 9 `bit:*` skills + 3 agents.

**Q3 — MCP: works, no interference.** With `--mcp-config dev.json --strict-mcp-config`, `mcp_servers` was just `[{name: bit, source: dynamic}]` — the local-scope `bp` server was dropped, and since the plugin declares no MCP there's nothing to conflict. The wrapper log proved the dev binary was spawned, with `cwd` and `CLAUDE_PROJECT_DIR` both = the launch dir, in both bit-pro and the empty dir. A `task_list` call succeeded (`{"tasks":[]}`) once `--allowedTools mcp__bit__task_list` was given. **In `-p` mode the dynamic server's tools were not pre-allowed, even in bit-pro**, so headless probes need `--allowedTools` (interactive sessions just prompt). `env` in the config JSON gets passed to the server (used for `XDG_DATA_HOME`). Nothing was written to the sandbox XDG dir because the current branch's bp is still v1 (it reads `.bit/`).

## Verdict
All three confirmed empirically. Caveat: if a future plugin.json adds `mcpServers`, `--strict-mcp-config` should also drop it (the help text says "ignoring all other MCP configurations"), but that was not tested.

## Recommended pre-cutover test recipe
```
# once per change
go build -o /tmp/bitdev/bp .            # or ./bin/bp; never `just install` on v2
cat > /tmp/bitdev/mcp.json <<EOF
{"mcpServers":{"bit":{"type":"stdio","command":"/tmp/bitdev/bp","args":["serve","mcp"],
  "env":{"XDG_DATA_HOME":"/tmp/bitdev/xdg"}}}}
EOF
# session, from a sandbox or throwaway project dir
claude --plugin-dir <bit-pro-v2-worktree>/bit \
       --mcp-config /tmp/bitdev/mcp.json --strict-mcp-config
```
- Keep the server name `bit` so the skills' `mcp__bit__*` tool names still match.
- Runs from any dir, including bit-pro itself: the flags override the installed plugin and local MCP for that session only. No change to `~/.claude*`, installed plugins or project `.claude/`, so the daily v1 setup stays intact.
- Confirm in-session via `/plugin` or `/mcp` (or the `init` event in `-p --output-format stream-json --verbose`): `bit@inline` and `bit` MCP `source: dynamic`.
- Point `--plugin-dir` at a v2 worktree so the main checkout can stay on `main`.
