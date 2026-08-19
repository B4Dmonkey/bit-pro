---
id: BIT-28.11
title: Contradiction forces bp start to reconcile
status: done
phase: 2
phase_label: lifecycle
---
## **Verse 2**

`bp start` reconciles the state it finds instead of assuming one: already running, loaded but not
running, and not loaded are three different jobs. Contradicts BIT-28.10, which bootstraps
unconditionally and always claims it started something — bootstrapping an already-bootstrapped job
is an error, and it would report a start that never happened.

## Scope
- `launchd/status.go` — extract an unexported `listJob(ctx, run) (loaded bool, pid int, err error)`
  and have `Status` call it, so the loaded-versus-not distinction `Status` currently collapses is
  available inside the package
- `launchd/start.go` — branch on it
- `launchd/start_test.go`, `cmd/start_test.go` — the new tests

This is where the fourth reconcile case lands; the plist-missing case is already covered by
BIT-28.8's two rows.

## TDD cycle

1. **Write test (RED):** `cmd/start_test.go`
   - [ ] `TestStartCmd_ReconcilesTheStateItFinds` (table-driven)
     - **Behavior:** `bp start` is safe to run twice. An operator who is unsure whether the daemon
       is up can just run it, and gets the truth rather than an error or a false claim.
     - **Setup:** `t.Setenv("HOME", t.TempDir())`; the recording fake runner from BIT-28.10, with
       `print-disabled` returning an empty store in every row. Three rows, keyed on what the
       *first* `list <label>` returns: (a) already running — a dict containing `"PID" = 4242;`,
       code `0`; (b) loaded but not running — the same dict with the `PID` line removed, code `0`,
       and a later `list` returning `"PID" = 4242;` once a `kickstart` has been recorded;
       (c) not loaded — `("", 113, nil)`, and a later `list` returning `"PID" = 4242;` once a
       `bootstrap` has been recorded.
     - **Assertions:** (a) output is exactly `"already running (pid 4242)\n"`, and **no** recorded
       call contains `bootstrap` or `kickstart`. (b) output is exactly `"started (pid 4242)\n"`, a
       call `launchctl kickstart gui/<uid>/com.github.b4dmonkey.bit-pro` is recorded, and **no**
       `bootstrap` call is. (c) output is exactly `"started (pid 4242)\n"`, a `bootstrap` call is
       recorded, and **no** `kickstart` call is.
     - **Boundary:** the three load states `launchctl list` can report — running, loaded-idle, and
       absent — which together with BIT-28.8's plist-missing pair complete the reconcile matrix.
       Rows (a) and (b) are the ones that fail against an unconditional bootstrap.
   - [ ] Confirm fails: row (a) fails with output `"started (pid 4242)\n"`, want
         `"already running (pid 4242)\n"`, and on the recorded `bootstrap` call that should not be
         there; row (b) fails on the missing `kickstart`.

2. **Implement (GREEN):**
   - [ ] `launchd/status.go`: extract `listJob(ctx context.Context, run Runner) (bool, int, error)` —
         run `launchctl list <Label>`, return `loaded=false` on a non-zero `code`, otherwise
         `loaded=true` plus the pid from the existing `"PID"` regexp (`0` when absent). Rewrite
         `Status` to call it, leaving its behaviour and every BIT-28.6 / BIT-28.7 assertion unchanged.
   - [ ] `launchd/start.go`: in `Start`, before enabling, call `listJob`; when it reports
         `loaded && pid != 0`, return `StateRunning` with that pid and a new
         `alreadyRunning bool` return so the command can pick its wording. Do not short-circuit on
         `StateStopped` — enabling and bootstrapping is exactly the recovery from a previous stop.
   - [ ] `launchd/start.go`: after the `enable` call, run
         `launchctl kickstart <domain>/<Label>` when `listJob` reported `loaded`, and
         `launchctl bootstrap <domain> <plistPath>` when it did not. Then return `Status(ctx, run)`
         as before.
   - [ ] `cmd/start.go`: print `already running (pid N)` when `Start` reports it, `started (pid N)`
         otherwise.

## Claude verifies
- [ ] `just test` — including that BIT-28.6's and BIT-28.7's `Status` tests still pass after the
      `listJob` extraction
- [ ] `just lint`

## User verifies
- [ ] `just install`, then run `bp start` twice in a row. The first prints `started (pid N)`; the
      second prints `already running (pid N)` with the **same** pid — the daemon was not restarted.

## Commit (user)
`feat(start): reconcile the existing launchd state before starting`