# BIT-45 soundness pass (v2 @ 6a1d345)

**Checked:** does the delete list in `daemon-removal` hold against the code, does what's said to survive actually compile after the removal, and are there gaps no verse covers.

## Confirmed (re-verified, not carried over)
- Only `cmd/` imports `daemon` (`grep 'bit-pro/daemon"'`: serve.go, start.go, stop.go, status.go, root.go + their tests).
- `claude/dispatch.go` exports are used only by `daemon/loop.go` and `cmd/serve.go:27` (`serveRunner = claude.DirRunner(claude.ExecDirRunner)`). Safe to delete.
- `Store.Counts()` has one non-test caller, `daemon/loop.go:66`. `task/counts.go` is safe to delete.
- `store` package survives: `db/open.go:20` uses `store.Dir()` (the other caller, `cmd/start.go:59`, is deleted).
- `claude.Runner`/`SyncPlugin`/`RegisterMCP` survive through `writeClaudeWiring` (`cmd/init.go:51`), used by `add` and `init`.
- TUI enqueue/Play live in `tui/model.go` (:89-91, :174-181, :198-208, :319-374) **and** `tui/board.go` (:78-84 queued param, :227 `enqueueSelected`, :272 Play prompt text), `tui/delegate.go:19-34`. Touches list is right.
- "No drop migrations" compiles: `ListProjects` keeps returning count columns; `bp list` just stops printing them (`cmd/list.go:33-34`). `UpdateProjectCounts` stays generated but dead.
- No Justfile, `scripts/`, `update/`, plugin.json or skill references the daemon (`grep -il daemon` hits only cmd/, tui/, daemon/, db/queue*, claude/dispatch_test.go, automation-notes.md, mcp-notes.md, v2-sketch.md).

## Issues
- **should-fix — test fallout wider than Touches.** `newRootCmd(run, lc daemon.Runner)` (`cmd/root.go:128`) loses its second param, so every caller changes: `cmd/cmd_test.go` (:24-61, :128), `cmd/root_test.go:223,243`, **`cmd/task_test.go:18`** (not in `daemon-removal`). `cmd/list_test.go:67-100` (`TestListCmd_ShowsProjectCounts`) must go with the counts. `seedProject` (`cmd/list_test.go:117`) is shared with serve_test/status_test.
- **nit — `cmd/add.go:24`** Short text "Enroll a project in the registry the daemon watches"; `add.go` is not in Touches.
- **nit — generated code is gitignored.** `.gitignore` has `db/orm/`, so `db/orm/queue.sql.go` isn't a tracked delete; it disappears only when sqlc regenerates. Verify sqlc removes stale output (unverified) or delete the local file; otherwise a stale `queue.sql.go` still compiles and hides a missed caller. Plain `go build` on a fresh checkout needs `just db-gen-queries` first.
- **nit — interim db path.** Between BIT-45 and BIT-46, a v2 build still opens `~/.local/share/bit-pro/bit.db` (`store/store.go:22`, `db/open.go:25`), the same file v1 uses. Only harmful if a v2 build runs without the sandboxed `XDG_DATA_HOME`; the schema is unchanged so v1 isn't broken.

## Verdict
Sound as one vertical slice. Nothing it removes is needed by BIT-46/47.
