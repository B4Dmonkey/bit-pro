# BIT-45 soundness pass 2 (v2 @ 6a1d345)

**Checked:** every claim and Touches pointer in the revised body against the code.

## Confirmed
- The launchd plist is at `~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist`, and `launchctl list` shows no bit-pro label (not loaded).
- `cmd/add.go:24` Short is "Enroll a project in the registry the daemon watches".
- `cmd/task_test.go:18` calls `newRootCmd(..., nothingLoaded)`, the daemon runner argument.
- `cmd/list_test.go` `TestListCmd_ShowsProjectCounts` is around line 67.
- `db/orm/` is gitignored (`.gitignore:3`), so a stale `queue.sql.go` is real.
- `cmd/serve.go:74,89` holds the daemon subcommand.
- The TUI enqueue and "Play?" live in `tui/board.go:227,272`, `tui/model.go:89,174,319`, and `cmd/tui.go:32-65` (`queueFuncs`, `listQueue`).
- `claude.Runner` is still needed by `add`/`init` (`cmd/init.go:51`).

## Should-fix: Touches misses files the removal forces
- `db/queries/projects.sql`: `ListProjects` selects `backlog, todo, done, completed` and `UpdateProjectCounts` exists. The "list loses its counts" decision means `cmd/list.go:33` changes. Either these queries are trimmed now, or the count columns stay until BIT-46's fresh schema. Pick one; they aren't in Touches.
- Tests outside `cmd/`: `db/queue_test.go` (it breaks once `queue.sql` goes), `db/queries_test.go`, `claude/dispatch_test.go`, `task/counts_test.go`, `tui/{board,delegate,model}_test.go` (e.g. `tui/board_test.go:877` `TestUpdateBoard_EKey_EnqueuesTrackBars`), and `cmd/{start,stop,status}_test.go`. The Tests line lists only `cmd/` files.

## Nit
- The TUI also shows queued state through `WithListQueue` (`cmd/tui.go:33,46`). "Browse, reload, approve work as before" implies it goes too, but it isn't named.
- Until BIT-46, a dev `bp add`/`bp list` without `XDG_DATA_HOME` still hits v1's live `~/.local/share/bit-pro/bit.db` (`store/store.go:22`, `db/open.go:25`). BIT-46's sandbox rule covers this. BIT-45 doesn't restate it.

No ordering problem: the single verse is a usable slice, and BIT-46 depends on it.
