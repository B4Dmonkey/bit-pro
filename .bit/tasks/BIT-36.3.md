---
id: BIT-36.3
title: Contradiction forces stale-plist repair in bp start
status: todo
approved: true
phase: 2
phase_label: stale plist repair
---
## **Verse 2**

`enrollDaemon` currently returns early when the plist file exists, so an already-enrolled machine's plist keeps pointing at `serve` (the parent, no RunE) after upgrading. A new test proves the stale case needs repair; the current skip-if-exists logic can't satisfy both it and the no-plist case without real comparison logic.

This is the last bar of Verse 2, so it carries the integration check.

## Scope
- `daemon/stop.go` — add exported `func Bootout(ctx context.Context, run Runner)` that calls `launchctl bootout` and ignores the exit code (not loaded is not an error for a repair)
- `cmd/start.go` — add `ctx context.Context, lc daemon.Runner` to `enrollDaemon`; compare canonical plist to on-disk; call `daemon.Bootout` + rewrite if they differ; update the `RunE` call site to pass `cmd.Context()` and `lc`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStartCmd_RepairsAStalePlist` in `cmd/start_test.go`
     - **Behavior:** when the on-disk plist differs from the canonical form, `bp start` does bootout then bootstrap and rewrites the plist to canonical
     - **Setup:** write `"<string>serve</string>"` (stale, no `daemon`) to the LaunchAgents path; use `recordingLaunchctl(&calls, bootstrapCall(home), "", 113)` as the runner so bootstrap triggers the "now loaded" transition
     - **Assertions:**
       - `bootoutCall()` appears in `calls` before `bootstrapCall(home)` (use `slices.Index`)
       - `out == startedPID`
       - plist on disk passes `assertEnrollsTheDaemon`
     - **Boundary:** plist exists but differs — the single condition that distinguishes "repair needed" from "no action"; stale content is `"<string>serve</string>"` alone (the exact state prior to Bar 2's template fix)
   - [ ] Confirm fails: `enrollDaemon` returns `path, nil` on any existing file; bootout is never called

2. **Implement (GREEN):**
   - [ ] In `daemon/stop.go`, add:
     ```go
     func Bootout(ctx context.Context, run Runner) {
         run(ctx, "launchctl", "bootout", domain()+"/"+Label) //nolint:errcheck
     }
     ```
     (Error and exit code are intentionally ignored — the daemon may not be loaded.)
   - [ ] In `cmd/start.go`, change `enrollDaemon()` to `enrollDaemon(ctx context.Context, lc daemon.Runner) (string, error)`
   - [ ] Inside `enrollDaemon`, after computing `canonical := daemon.Plist(exe, filepath.Join(dir, "daemon.log"))`:
     - Read the existing file; if `os.IsNotExist` → write canonical and return (existing branch stays)
     - If other read error → return error
     - If `bytes.Equal(existing, canonical)` → return path (no-op)
     - Otherwise: `daemon.Bootout(ctx, lc)`, then `daemon.WritePlist(path, canonical)`, return path
   - [ ] Update the `RunE` call site: `path, err := enrollDaemon(cmd.Context(), lc)`

3. **More tests (RED → GREEN):**
   - [ ] `TestStartCmd_WritesThePlistOnlyWhenMissing`, `"a plist the operator edited"` case: change `check` from asserting `plist == editedPlist` to `assertEnrollsTheDaemon` — any differing plist is now repaired, not preserved. Consider renaming the test to `TestStartCmd_EnrollsOrRepairsThePlist`.
   - [ ] `TestStartCmd_WritesThePlistOnlyWhenMissing`, `"no plist on disk"` case: unchanged (still calls `assertEnrollsTheDaemon`).

## Claude verifies
- [ ] `just test ./cmd/... ./daemon/...` passes
- [ ] `just lint` passes

## User verifies
- [ ] Verse 2 integration (stale-plist repair): run `just install` to get the new binary; `bp stop` if the daemon is running; manually edit `~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist` to remove the `<string>daemon</string>` line; then `bp start` — confirm the plist is rewritten to contain `<string>daemon</string>` and `bp status` shows running

## Commit (user)
`fix(start): repair stale plist on start (bootout → rewrite → bootstrap)`