---
id: BIT-40
title: bp init registers the MCP server
status: done
---
## Why

The MCP server works and is invisible. Steps 1–3 of the MCP phase built eight tools behind
`bp serve mcp`, but nothing tells Claude Code that the server exists — reaching it today means
hand-registering it per project. So a project that `bp init` has set up still gets a Claude that
falls back to `mv`, `cat`, and `sed` against `.bit/`, which is the exact behaviour the whole phase
exists to remove. `bp init` is already the one command an operator runs to wire a project for bit;
registration belongs there, and until it lands the tools have no delivery path at all.

## Summary

`bp init` gains a third wiring step: after writing `.claude/settings.json` and syncing the plugin,
it registers the MCP server at local scope by shelling out to
`claude mcp add bit -- bp serve mcp`. The operator's next session in that project lists
`mcp__bit__*` and Claude can drive the ledger through typed tools instead of a shell.

## Visual aid

```
bp init
  ├─ task.SaveConfig            .bit/config      (exists)
  └─ writeClaudeWiring
       ├─ claude.WriteSettings  .claude/settings.json    (exists)
       ├─ claude.SyncPlugin     claude plugin update…    (exists)
       └─ register the server   claude mcp add bit …     (this scope)
                                     ↓
                            ~/.claude.json
                            projects."<abs path>".mcpServers.bit
                                     ↓
                            next session lists mcp__bit__task_read …
```

## Decisions

- **The registered command is the bare word `bp`, not an absolute path.** So the command written
  is exactly `claude mcp add bit -- bp serve mcp`, and `bp` resolves against the `PATH` of whatever
  process Claude Code launches the server from. This follows the binary wherever `just install`
  puts it rather than pinning to where it sat at init time. The cost is that a spawning process
  with a minimal `PATH` — a GUI launcher that does not inherit the shell's — produces an entry that
  looks correct and never connects, with nothing pointing at `PATH`. Accepted knowingly; if that
  turns up, the fix is to revisit this decision, not to work around it downstream.
- **`claude mcp add` is idempotent, so registration is a bare add.** The operator's call: a second
  add over an existing `bit` entry overwrites rather than failing, so `bp init` needs no
  check-then-add and no "already registered" branch. Verse 2 is what proves it — re-running init
  leaves exactly one entry.
- **Local scope, not project or user.** The entry is the nested
  `projects."<abs path>".mcpServers` key in `~/.claude.json`. Project `.mcp.json` was rejected
  because it writes into the project, reversing the direction `bp init` deliberately holds
  (`TestInitCmd_WritesNoSkills`); user scope was rejected because it loads the tools in every
  project on the machine, including ones with no `.bit/`.
- **Register by shelling out to `claude mcp add`, never by editing `~/.claude.json`.** That file
  holds all of Claude Code's per-project state, so `bp` must not read-modify-write it. `claude mcp
  add` defaults to local scope and to stdio transport, so the command is exactly
  `claude mcp add bit -- bp serve mcp`.
- **This adds no new dependency.** `bp init` already requires `claude` on `PATH` —
  `claude.SyncPlugin` shells out to `claude plugin` today — so the stated cost in `mcp-notes.md`
  is already paid. The existing `claude.Runner` seam is the injection point, and it is already
  faked in tests.
- **The server is named bare `bit`.** Claude Code namespaces tools as `mcp__<server>__<tool>`, so
  `bit` yields `mcp__bit__task_read`. These are the names step 6's deny rules will have to spell,
  and a plugin-declared server would have forced the scoped `plugin:bit:bit` form instead.
- **Declaring the server in the plugin manifest stays rejected.** The inline `mcpServers` key is
  silently stripped during manifest parsing (`anthropics/claude-code` #16143, open), and a
  plugin-declared server registers under a scoped name. Recorded as why the route is closed, not
  as an open question.
- **Dispatched sessions are out of scope.** Local scope covers an operator working in a checkout.
  Getting the server in front of a session the daemon spawns elsewhere is a property of dispatch
  and is owned in `automation-notes.md`.
- **Registration is additive and undone by not using it.** Nothing existing changes behaviour, so
  this step sits before the step-5 version tag rather than behind it.

## Verses

- [x] Verse 1 — An operator who runs `bp init` gets the bit tools in their next session: init
  registers the server at local scope, and the tools appear without anyone hand-editing config.
  Touches: `cmd/init.go` (`writeClaudeWiring`) and the `claude` package alongside `sync.go` —
  where to look to verify.
- [x] Verse 2 — Re-running `bp init` stays safe: the routine `just install` → `bp init` loop leaves
  exactly one correct `bit` entry and the rest of `~/.claude.json` untouched, rather than erroring
  or duplicating.
  Touches: the same registration path in the `claude` package — where to look to verify.

## References

- `mcp-notes.md` — the MCP phase's working notes; step 4 "Registration" is this scope, and its
  Decisions and Measured facts sections are the source for the decisions above.
- `automation-notes.md` — owns dispatch, and therefore owns registering the server for a
  daemon-spawned session.