# Operator decisions

These are operator decisions, not research findings, unless an entry is marked "Claude default".

## 2026-09-30
- This track was split out of the original BIT-46. It covers the global Claude wiring (user-scope marketplace, plugin and MCP server) and removing `bp init`. (2026-10-01: deleting `bp init` moved to BIT-46 Verse 1. See below.)
- The store work doesn't depend on this track, because the dev loop (`--plugin-dir`, `--mcp-config`, `--strict-mcp-config`) bypasses the wiring.
- Cutover: the operator alone decides when v2 is ready and merges the branch. No track gates cutover. Removing each project's old per-project wiring is a manual step on the cutover checklist in `v2-sketch.md`.

## 2026-10-01
- **`bp init` is deleted in BIT-46 Verse 1,** because it depends on `bitdir` and `config.toml`, which that verse removes. BIT-46 keeps the per-project wiring helper `writeClaudeWiring` for `bp add`, and this track replaces it with the global ensure step and removes it, along with `WriteSettings`.
- Files this track owns: `claude/sync.go`, `claude/settings.go`, `claude/plugin.go`, the `writeClaudeWiring` helper, and the plugin-behind notice text in `cmd/root.go` (:119).
- **The wiring runs only when a project is first registered.** Re-running `bp add` or `bp migrate` on a registered project doesn't set up the wiring. The wiring is global and user-scope, and this track owns it.
- **Repairing lost global wiring is out of scope.** The global wiring is the user-scope plugin and MCP registration. If the operator removes it, they re-run the documented `claude` commands. There's no bp command for it.
- **Claude default:** the README documents the `claude` commands (four since the readiness pass below).
- **Claude default:** a fresh registration is saved before the wiring runs. If a wiring step fails, `bp add` says which step failed and prints the commands to run by hand, and the project stays registered.

## 2026-10-01 readiness pass (Claude defaults; the operator can veto these)
- **The ensure step runs `claude plugin marketplace update bit-pro` between `marketplace add` and `install --scope user`.** On a machine that already has the marketplace, `add` only says "already on disk" and doesn't refresh (topic BIT-46 `soundness-2`), so `install` could take a stale cached plugin. v1's `SyncPlugin` updates first too (`claude/sync.go:34`).
- **Reviving a removed project with `bp add` doesn't run the wiring.** It was registered before, and the wiring is global.
- **Dev builds never use `just install` on `v2`** (the BIT-46 rule, repeated here because the memory `bit-pro-just-install-after-changes` says otherwise).
- **The `cmd/cmd_test.go` helpers `pluginSyncCalls` and `mcpRegisterCall` (:98-111)** assert on the per-project calls and change with them, so each bar stays green.

The evidence for the wiring design is in BIT-46 topics `soundness-2`, `init-and-update` and `plugin-dir-testing`.