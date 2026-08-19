---
id: BIT-28.10
title: bp start enables before it bootstraps
status: done
phase: 2
phase_label: lifecycle
---
## **Verse 2**

`bp start` enables the label, bootstraps the job, and reports `started (pid N)`. Contradicts
BIT-28.8, which writes the plist and then does nothing and says nothing — a plist on disk that
launchd has never been told about is not a running daemon.

## Scope
- `daemon/start.go` — new: `Start(ctx context.Context, run Runner, plistPath string) (State, int, error)`
- `daemon/daemon.go` — a `domain()` helper returning `"gui/" + strconv.Itoa(os.Getuid())`, now
  used by both `Status` and `Start`
- `cmd/start.go` — render the result
- `daemon/start_test.go`, `cmd/start_test.go` — the new tests

## References
- `automation-notes.md` (repo root, untracked) — "Daemons on macOS": `bootstrap` takes a **path to
  the plist**, while `bootout` and `kickstart` take a **`<domain>/<label>`** and `list` takes a
  **bare label**. `launchctl load` / `unload` are the deprecated forms that appear in older blog
  posts — do not use them.

## TDD cycle

1. **Write test (RED):** `cmd/start_test.go`
   - [ ] `TestStartCmd_EnablesBeforeBootstrapping`
     - **Behavior:** `bp start` after any previous `bp stop` actually restarts, rather than
       hard-failing on a label launchd still has in its disabled store.
     - **Setup:** `t.Setenv("HOME", t.TempDir())`; a fake `daemon.Runner` that appends every call
       to an ordered slice and answers: `print-disabled` with an empty store; `list <label>` with
       `("", 113, nil)` until a `bootstrap` call has been recorded and with a dict containing
       `"PID" = 4242;` and code `0` afterwards; `enable` and `bootstrap` with `("", 0, nil)`.
       Run `runWithDaemon(t, lc, startCmdUse)`.
     - **Assertions:** output is exactly `"started (pid 4242)\n"`. In the recorded calls, the
       index of `launchctl enable gui/<uid>/com.github.b4dmonkey.bit-pro` is **less than** the index
       of `launchctl bootstrap gui/<uid> <plistPath>`, where `<uid>` is `os.Getuid()` and
       `<plistPath>` is `$HOME/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist`. Assert on
       the ordering, not merely on both being present.
     - **Boundary:** the ordering constraint itself. It is measured, not stylistic: with the label
       in the disabled store, `launchctl bootstrap` fails with `Bootstrap failed: 5: Input/output
       error` and exit 5, so the `enable` in the earlier position is the whole point.
   - [ ] Confirm fails: output is `""`, want `"started (pid 4242)\n"` — BIT-28.8 makes no
         `launchctl` calls at all, so the recorded slice is empty and both index lookups miss.

2. **Implement (GREEN):**
   - [ ] `daemon/daemon.go`: `func domain() string` returning `"gui/" + strconv.Itoa(os.Getuid())`;
         rewrite `Status`'s `print-disabled` argument to use it.
   - [ ] `daemon/start.go`: `Start(ctx, run, plistPath)` — `run(ctx, "launchctl", "enable", domain()+"/"+Label)`,
         then `run(ctx, "launchctl", "bootstrap", domain(), plistPath)`, then return
         `Status(ctx, run)`. Propagate a non-nil `err` from either call; a non-zero `code` from
         `enable` is not fatal on its own, so surface it only via the state `Status` reports.
   - [ ] `cmd/start.go`: after the plist write, call `daemon.Start`; print
         `started (pid N)` using the returned pid. Keep it to the terminal state — no narration of
         the plist write or the individual `launchctl` calls.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] `just install`, then — on a machine where the daemon has never been started — run `bp start`.
      It prints `started (pid N)` with a real pid, not `pid 0`. (A fake runner cannot prove launchd
      has populated the pid by the time `bp start` asks; only a real run can.) Then `bp status`
      prints `running (pid N)` with that same pid.
- [ ] Tear back down by hand for now, since `bp stop` does not exist until BIT-28.12:
      `launchctl bootout gui/$UID/com.github.b4dmonkey.bit-pro`.

## Commit (user)
`feat(start): enable and bootstrap the LaunchAgent`