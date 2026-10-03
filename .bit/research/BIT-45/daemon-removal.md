# Removing the daemon/queue (Q1)

**Checked:** exactly what goes when the daemon/queue feature is removed, what depends on it, and what survives of `bp add`/`list`/`status`. Paths are relative to the repo root.

## Delete whole
- `daemon/` (all 8 files). Only `cmd/` imports it.
- `claude/dispatch.go` and `dispatch_test.go`. `DirRunner`, `ExecDirRunner`, `WorktreeName`/`slug`, `Agent`/`Under`, `Agents` and `Spawn` are used only by `daemon/loop.go` and `cmd/serve.go:27`.
- `cmd/start.go`, `cmd/stop.go`, `cmd/status.go` and their tests.
- `task/counts.go` and `counts_test.go`. The only caller of `Store.Counts()` is `daemon/loop.go:66`.
- `db/queries/queue.sql`, the generated `db/orm/queue.sql.go` and `db/queue_test.go`.
- `clear-queue.sh`. `automation-notes.md` covers the daemon throughout. `mcp-notes.md` has daemon sections (roughly 219-232 and 334-447).

## Partial edits
- **`cmd/serve.go`**
  - Remove `newServeDaemonCmd`, `BP_CLAUDE`/`claudeBinEnv`, `serveTick`, `serveRunner`, `claudePath` and the slog `newHandler`.
  - `serve` keeps only `serve mcp`.
  - Almost all of `cmd/serve_test.go` goes. Keep `TestServeMCPCmd_IsListedInServeHelp` (~:158).
- **`cmd/root.go`**
  - Remove the `daemon` import, `daemon.ExecRunner` (:125), the `lc daemon.Runner` param (:128) and the start/status/stop wiring (:143-145).
  - `claude.Runner` **stays**. `init`/`add` use it through `writeClaudeWiring` (`cmd/init.go:51`) for `SyncPlugin` + `RegisterMCP`.
  - In `cmd/cmd_test.go`, remove `runWithDaemon` and `nothingLoaded`. In `cmd/root_test.go:216`, remove the "serve daemon" row.
- **`cmd/tui.go`**: remove `queueFuncs` and the `db.Open` block (:31-94). The file then needs no `db`/`orm`/`os`. It has no test file.
- **`tui/`**
  - Remove enqueue/listQueue/queued from `model.go`: fields :89-91, `WithEnqueue`/`WithListQueue` :174-181, the reload branch, `applyQueued`, `enqueueSelected` and the "e" key.
  - Remove the `queued` param from `board.go` and `queuedColor` from `delegate.go`.
  - **Product call:** the "Play X? (y/n)" prompt, which comes after approve, has enqueue as its only "y" action (`model.go:319-336`). Remove it or repurpose it.
  - Tests to cut: `tui/model_test.go` 38-160 and 1456-1697, `board_test.go:877`, `delegate_test.go:278-300`.
  - The TUI keeps browse, reload and approve.

## db
- The old migrations (projects, queue, count columns, queue unique) are moot. **The v2 db is fresh at a new path** (operator decision). v2 can therefore replace the four migrations with a new initial set instead of adding drop migrations.
- `db/open_test.go:42` asserts exactly 4 migrations, so update it with the new set.
- `UpdateProjectCounts` becomes dead: callers are daemon plus `cmd/list_test.go:90-100` and `status_test`. So do `GetProjectByPath` (exact match, TUI only) and the count columns in `ListProjects`.
- `store.Dir()` stays, because `db/open.go:20` uses it. Rename its dir `bit-pro` → `bit`.

## What survives as project registration
- **`bp add`** (`cmd/add.go`) is the registration command. It checks `ProjectExists`, reads `.bit/config.toml` for a default code, prompts for a code, runs `writeClaudeWiring` (settings.json, plugin sync, `claude mcp add bit`) if `.bit/` is missing, and then calls `CreateProject`.
  - v2 changes: drop the `.bit` checks, and make the code UNIQUE and validated.
  - The Short text at :24 says "the daemon watches". Reword it.
  - Consider merging `init` and `add`. Both do Claude wiring, `init` writes config.toml, and `add` writes the db row. In v2 both reduce to "register path+code, wire Claude".
- **`bp list`** survives without counts (`cmd/list.go:32-33`). `cmd/list_test.go:117 seedProject` is shared with the serve/status tests.
- **`bp status`**: nothing survives. It is launchd status plus counts.

## Machine state at cutover
- `~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist` **exists on this machine**, but is not loaded (`launchctl list` shows nothing). Delete it by hand at cutover, next to deleting `~/.local/share/bit-pro/`. After v2 there is no code left to `bootout` it.
- No Justfile or `scripts/` target references the daemon.

## Keep (shared, not daemon-only)
`signalContext` (root.go), `task.ParentID`/`barParent`, `claude.ExecRunner`/`RegisterMCP`/`SyncPlugin`, and `bit/agents/bot-dev.md`, which is still usable interactively.

## Ordering note
Do this first. It deletes about half of `db/` and the only exact-path project lookup, so the new resolver and registry are built on a smaller surface.
