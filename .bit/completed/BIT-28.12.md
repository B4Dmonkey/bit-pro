---
id: BIT-28.12
title: bp stop brings the daemon down durably
status: done
phase: 2
phase_label: lifecycle
---
## **Verse 2**

`bp stop` boots the job out and disables the label, in that order, and reports `stopped`. This is
the bar that makes a stop durable — `bootout` alone is session-scoped, and launchd re-walks
`~/Library/LaunchAgents/` at login, so without the `disable` a daemon the operator stopped would
silently resurrect itself after a reboot.

## Scope
- `daemon/stop.go` — new: `Stop(ctx context.Context, run Runner) error`
- `cmd/stop.go` — new; `newStopCmd(lc daemon.Runner)`
- `cmd/root.go` — register it
- `daemon/stop_test.go`, `cmd/stop_test.go` — the new tests

## References
- `automation-notes.md` (repo root, untracked) — "Daemons on macOS" for `bootout` being
  session-scoped and `disable` durable. The ordering within `bp stop` is settled on this track's
  scope and supersedes the note's line 187, which still asks for it to be measured; the probe was
  run during the scope pass, so do not re-run it.

## TDD cycle

1. **Write test (RED):** `cmd/stop_test.go`
   - [ ] `TestStopCmd_BootsOutThenDisables` (table-driven)
     - **Behavior:** stopping is an explicit operator intent that outlives the login session, and it
       reports one terminal state regardless of what it had to do to get there.
     - **Setup:** the recording fake `daemon.Runner`. Two rows: (a) the job is loaded — every call
       returns `("", 0, nil)`; (b) the job is not loaded — the `bootout` call returns
       `("Boot-out failed: 3: No such process", 3, nil)` and the rest return `("", 0, nil)`.
       Run `runWithDaemon(t, lc, stopCmdUse)`.
     - **Assertions:** both rows — output is exactly `"stopped\n"` and the returned error is `nil`;
       the recorded calls contain `launchctl bootout gui/<uid>/com.github.b4dmonkey.bit-pro` at an
       index **less than** `launchctl disable gui/<uid>/com.github.b4dmonkey.bit-pro`, with
       `<uid>` from `os.Getuid()`. Row (b) additionally asserts the `disable` call is recorded even
       though `bootout` failed.
     - **Boundary:** `bootout`'s success and failure states, and the ordering of the pair. The order
       is measured, not reasoned: `disable` does not kill a running job, so disabling first would
       leave the daemon alive while marked disabled — and since `bp status` checks the disabled
       store first, it would report `stopped` about a live daemon. Booting out first fails honestly
       instead. Row (b) is the row that would break if a failing `bootout` short-circuited the
       `disable`, which is the case that silently un-does durability.
   - [ ] Confirm fails: `unknown command "stop" for "bp"`.

2. **Implement (GREEN):**
   - [ ] `daemon/stop.go`: `Stop(ctx, run)` — `run(ctx, "launchctl", "bootout", domain()+"/"+Label)`,
         ignoring a non-zero `code` (a job that is not loaded is already booted out) but propagating
         a non-nil `err`; then `run(ctx, "launchctl", "disable", domain()+"/"+Label)`, propagating
         both a non-nil `err` and a non-zero `code` from this one — a `disable` that did not take is
         a stop that will not survive a reboot, and must not be reported as success.
   - [ ] `cmd/stop.go`: `const stopCmdUse = "stop"`; `newStopCmd(lc daemon.Runner)` with
         `Args: cobra.NoArgs`; `RunE` calls `daemon.Stop` and prints `stopped`.
   - [ ] `cmd/root.go`: `rootCmd.AddCommand(newStopCmd(lc))`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `go build ./...`

## User verifies
Whole slice — Verse 2's capability is a supervised daemon an operator can start, watch, and stop,
and the reboot half is the part no automated test can reach. Run these in order after `just install`:

- [ ] `bp start` → `running (pid N)` from `bp status`. Close that terminal window entirely, open a
      new one: `bp status` still prints `running (pid N)` with the same pid.
- [ ] `ls -l ~/.local/share/bit-pro/daemon.log` — the file exists. It is expected to be zero bytes:
      the stub loop logs ticks at `debug` and the plist passes no `-v`.
- [ ] `bp stop` → `stopped`, and `bp status` agrees. Confirm it is genuinely down, not just marked:
      `launchctl list com.github.b4dmonkey.bit-pro` exits non-zero.
- [ ] **Reboot the machine**, then `bp status` — still `stopped`. This is the durability the
      `disable` buys, and the only way to see it.
- [ ] `bp start` after that reboot → `started (pid N)`. This is the case that fails outright if the
      `enable` were dropped, since the label is now in the disabled store.

## Commit (user)
`feat(stop): bring the daemon down durably with bootout and disable`