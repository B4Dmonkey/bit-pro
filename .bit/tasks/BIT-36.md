---
id: BIT-36
title: Rename bp serve → bp serve daemon
status: doing
---
## Why
`bp serve` is today a leaf command whose `RunE` *is* the daemon loop. The MCP phase (`mcp-notes.md`) needs `bp serve mcp` to live beside the daemon as a sibling — which requires `serve` to become a parent first. Without the rename, the two servers can't share a grouping, and `bp serve` would have to be both a parent and the daemon itself, which Cobra cannot express cleanly.

## Summary
Turn `bp serve` from a leaf command into a parent with two children. Move the current daemon loop body to `bp serve daemon`. Update the plist template to emit `serve daemon` in `ProgramArguments`. Fix `bp start` to detect and repair a stale plist (one that still says `serve`) by doing bootout → rewrite → bootstrap rather than skipping.

## Risks & unknowns

None — all questions are settled. See Decisions.

## Decisions

- **`serve` becomes a parent with no `RunE`; Cobra prints help on bare `bp serve`.** Settled in automation-notes.md.
- **The daemon body moves to `bp serve daemon`.** The plist's `ProgramArguments` follows. Settled.
- **`bp start` must rewrite a stale plist.** `enrollDaemon` currently writes the plist only when the file is absent (`cmd/start.go:60-63`). An already-enrolled machine keeps a plist pointing at bare `bp serve` — which, as a parent with no `RunE`, exits 0. `KeepAlive {SuccessfulExit: false}` therefore does not restart it. The fix is `bp start` comparing the on-disk plist against what it would write and, on a difference, doing bootout → rewrite → bootstrap. Settled in automation-notes.md.
- **Tests reference `serveCmdUse` throughout.** The constant moves to `"serve daemon"`; the test helper that invokes `runWithContext(t, ctx, serveCmdUse)` stays valid because Cobra routes two-word commands from the args slice. The IsListedInHelp test changes to assert `bp serve --help` lists `daemon`.
- **No MCP subcommand yet.** This scope is the rename only. `bp serve mcp` is the MCP phase's bar to add.

## Verses

- [x] Verse 1 — `bp serve daemon` runs the loop: `serve` becomes a parent, the daemon body moves to `serve daemon`, the plist template emits `serve daemon`, and tests pass against the two-word form.
  Touches: `cmd/serve.go`, `cmd/root.go`, `daemon/plist.go`, `cmd/serve_test.go`, `daemon/plist_test.go`.

- [ ] Verse 2 — `bp start` repairs a stale plist: on first start after an upgrade, if the on-disk plist differs from what `bp start` would write, it does bootout → rewrite → bootstrap instead of skipping. An already-enrolled machine running a new binary ends up with a daemon on `bp serve daemon`.
  Touches: `cmd/start.go`, `daemon/start.go`, `daemon/plist.go`, `cmd/start_test.go`.

## References

- `automation-notes.md` — "Pending rename" and "Two staleness cases" sections; the decisions this scope is built on.
- `mcp-notes.md` — the phase that requires `serve` to become a parent.