---
id: BIT-48.5
title: The per-project settings writer, SyncPlugin and RegisterMCP are deleted
status: todo
phase: 1
phase_label: bp add sets bit up for the whole machine
---
## **Verse 1**

After BIT-48.3, `WriteSettings`, `SyncPlugin` and `RegisterMCP` have no callers (exported, so lint doesn't flag them). This is a removal bar: no new tests, only deleting the code and the tests that covered it.

## Scope
- `claude/settings.go`: move `pluginKey` and `marketplaceName` (`:12-15`) into `claude/plugin.go` (used at `plugin.go:27,37,74` and by `global.go`), then delete the file (`WriteSettings`, `load`, `merge`, the `marketplace` var).
- delete `claude/settings_test.go`.
- `claude/sync.go`: delete `RegisterMCP` and `SyncPlugin`; keep `Runner` and `ExecRunner`.
- `claude/sync_test.go`: delete every `TestSyncPlugin_*` and `TestRegisterMCP_*`. Move `recorder`/`newRecorder` into `claude/global_test.go` and delete `sync_test.go` along with its now-unused consts (`scopeFlag`, `projectScope`, `bitProPlugin`, `updateSubCmd`, `mcpSubCmd`, `bitServer`, …; keep any `global_test.go` uses).

## Steps
- [ ] Move the consts, delete the files and functions above.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `grep -rn "WriteSettings\|SyncPlugin\|RegisterMCP" --include='*.go' .` finds nothing

## User verifies
- none — deterministic

## Commit (user)
`refactor(bit): delete the per-project Claude wiring`