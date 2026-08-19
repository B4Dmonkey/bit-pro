---
id: BIT-28.5
title: bp status answers not running
status: done
phase: 2
phase_label: lifecycle
---
## **Verse 2**

`bp status` exists and answers `not running`. The answer is hardcoded — BIT-28.6 contradicts it.
This is the cheapest possible first bar of the launchd slice: it pins the command's name, its
output vocabulary, and the fact that it never errors just because nothing is loaded.

## Scope
- `cmd/status.go` — new; `newStatusCmd()`
- `cmd/root.go` — register it
- `cmd/status_test.go` — the new test

No `launchd` package and no exec seam yet — nothing here needs to run a subprocess, so nothing
here should grow the wiring to allow one. BIT-28.6 is the test that demands it.

## TDD cycle

1. **Write test (RED):** `cmd/status_test.go`
   - [ ] `TestStatusCmd_ReportsNotRunning`
     - **Behavior:** asking about a daemon that was never started is an ordinary answer, not an
       error — an operator running `bp status` on a fresh machine gets a word, not a stack trace.
     - **Setup:** `run(t, statusCmdUse)`.
     - **Assertions:** returned error is `nil`; output is exactly `"not running\n"`.
     - **Boundary:** the empty end of the daemon's state range — nothing loaded and nothing
       disabled, the state every machine starts in.
   - [ ] Confirm fails: `unknown command "status" for "bp"`.

2. **Implement (GREEN):**
   - [ ] `cmd/status.go`: `const statusCmdUse = "status"`; `newStatusCmd() *cobra.Command` with
         `Args: cobra.NoArgs` and a `RunE` that does
         `fmt.Fprintln(cmd.OutOrStdout(), "not running"); return nil`.
   - [ ] `cmd/root.go`: `rootCmd.AddCommand(newStatusCmd())`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(status): add bp status reporting the not-running state`