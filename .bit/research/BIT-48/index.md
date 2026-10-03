# BIT-48 research index

- [claim-audit-2026-10-02](claim-audit-2026-10-02.md): per-claim CONFIRMED/WRONG/STALE audit, every call site and test that changes, a green bar order for Verses 1-2, what BIT-46 leaves in `cmd/{add,init,root}.go`, and gaps (wiring-failure exit status, root tests reading the real HOME). **Open before planning.**
- [review-2026-10-02](review-2026-10-02.md): independent review against the code and real `~/.claude*` (read only). **Open before planning.** Findings that change the plan:
  - Deleting `claude/settings.go` whole breaks the build, because `pluginKey`/`marketplaceName` (`:12-15`) are used by `claude/plugin.go:27,37,74`.
  - User-scope MCP lives at the top-level `mcpServers` of `$HOME/.claude.json` (sandbox-probed). The ensure step reads it, and on a read or parse error it attempts the add anyway.
  - `claude.Runner` returns only `error` (no output or exit code), and `DirRunner` dies in BIT-45, so an "already exists" check can't be done by inspecting the output.
  - Missing Touches: `cmd/init.go`, `cmd/add_test.go:85-111`, `claude/*_test.go`, `cmd/root_test.go:153`, `cmd/root.go:32`. The `cmd/cmd_test.go` helper range is really `:98-112`, plus `mcpLookupCall`.
  - Placement: `claude.EnsureGlobal(ctx, run, home)` with a thin cmd helper shared by `add` and BIT-49's `cmd/migrate.go`.
- [decisions](decisions.md): the operator decisions of 2026-09-30 and 2026-10-01, and the readiness-pass Claude defaults. Read first.
- BIT-46 topic `soundness-2`: sandbox-verified `claude` CLI behaviour for the user-scope marketplace, plugin and MCP server (`marketplace add` is required, `mcp get` is scope-blind, `mcp add -s user` fails when the entry exists), and the per-project plugin-behind notice. Its "Verse 1" is this track's Verses 1 and 2. Open it before planning.
- BIT-46 topic `init-and-update`: what `init` and `add` do today, with file:line. Its per-project recommendation is superseded by the global decision.
- BIT-46 topic `plugin-dir-testing`: the dev loop that bypasses the wiring.
- BIT-45 topic `testing`: why dev runs sandbox `HOME`.
