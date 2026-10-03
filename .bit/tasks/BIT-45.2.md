---
id: BIT-45.2
title: Orphaned daemon, dispatch and counts packages are deleted
status: done
approved: true
phase: 1
phase_label: No daemon in a v2 build
---
## **Verse 1**

Deletes the code BIT-45.1 left with no importer: `daemon/`, `claude/dispatch.go`, and the task counts that only the daemon read. This is a pure deletion. No test can demand that code go away, so there's no RED test. What forces the bar is that BIT-45.1 removed the last caller, and `unused` will flag `listCompleted` once `counts.go` goes, so it has to go in the same bar.

## Scope
- `daemon/`: delete the whole package (8 files: `daemon.go`, `loop.go`, `loop_test.go`, `plist.go`, `plist_test.go`, `start.go`, `status.go`, `stop.go`).
- `claude/dispatch.go`, `claude/dispatch_test.go`: delete. `DirRunner`/`ExecDirRunner`, `WorktreeName`, `Agents`, `Spawn` and the rest go with them. `claude.Runner`, `ExecRunner` (`claude/sync.go`), `SyncPlugin`, `RegisterMCP` and `plugin.go`/`settings.go` stay, because `init`/`add` use them.
- `claude/testdata/agents.json`: delete. Its only reader is `dispatch_test.go:11`. If `claude/testdata/` is then empty, the directory goes too.
- `task/counts.go`, `task/counts_test.go`: delete. `Store.Counts()` had one caller, `daemon/loop.go:66`.
- `task/store.go`: delete `(*Store).listCompleted` (:442-465, only caller `counts.go:38`) and `ParentID` (:548, only caller `daemon/loop.go:211`, no test). Keep `barParent` and `highestReserved`'s scan of `completed/`, which is separate code.

## TDD cycle

1. **No new test.** This is a deletion. The existing suites for `claude/` (sync, plugin, settings) and `task/` (store, task) are the regression net.
   - [ ] Before deleting, confirm nothing outside the deleted set references the targets: `grep -rn 'daemon\.\|DirRunner\|ExecDirRunner\|\.Counts()\|listCompleted\|ParentID(' --include='*.go' . | grep -v '^./daemon/'` returns only lines inside `claude/dispatch*.go`, `task/counts*.go` and the definitions in `task/store.go`.

2. **Delete (GREEN):**
   - [ ] Delete the files and functions in Scope.
   - [ ] `just lint`. Remove any import, const or helper that `unused`/the compiler now flags (e.g. a `task/testconst_test.go` const only `counts_test.go` used).

## Claude verifies
- [ ] `just test` passes.
- [ ] `just lint` reports `0 issues`.
- [ ] `test ! -e daemon && test ! -e claude/dispatch.go && test ! -e task/counts.go` exits 0.

## User verifies
- [ ] none (deterministic)

## Commit (user)
`refactor: delete the daemon, dispatch and task counts packages`