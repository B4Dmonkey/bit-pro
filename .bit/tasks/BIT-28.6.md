---
id: BIT-28.6
title: Contradiction forces the real launchctl list parse
status: todo
approved: true
phase: 2
phase_label: lifecycle
---
## **Verse 2**

A loaded job with a pid reports `running (pid N)`. Contradicts BIT-28.5's hardcoded answer, which
forces the real `launchctl list <label>` call, its output parse, and the exec seam that lets a test
stand in for `launchctl`.

## Scope
- `launchd/launchd.go` — new package: `Label`, `Runner`, `ExecRunner`, `State`
- `launchd/status.go` — new: `Status(ctx, run Runner) (State, int, error)`
- `cmd/status.go` — take a `launchd.Runner` and render its answer
- `cmd/root.go` — `newRootCmd` gains a `launchd.Runner` parameter; `NewRootCmd()` passes
  `launchd.ExecRunner`
- `cmd/cmd_test.go` — thread the new parameter through the helpers
- `launchd/status_test.go`, `cmd/status_test.go` — the new tests

**The exec seam.** `claude.Runner` returns only an `error` and discards output, so it cannot serve
here — `bp status` has to read what `launchctl` printed. Define a second, output-carrying runner in
the new package rather than widening `claude.Runner`, whose two existing call sites do not want it:

```go
type Runner func(ctx context.Context, name string, args ...string) (out string, code int, err error)
```

`err` is non-nil only when the process could not be run at all (that is also, deliberately, what a
non-macOS machine gets — no platform guard in this track); `code` carries the exit status of a
process that ran and refused, which is the case `launchctl list` uses to say "not loaded". Both are
trivial to fake, which an `*exec.ExitError` is not.

**Whole-module green.** Changing `newRootCmd`'s signature is a shared-signature change, so check it
holds: the only callers are `NewRootCmd()` in `cmd/root.go` and `runWithRunner` in `cmd/cmd_test.go`.
Both are updated in this bar, so the module still builds and every existing test still passes.

## References
- `automation-notes.md` (repo root, untracked) — "Daemons on macOS" and "Measured facts" carry the
  observed `launchctl list <label>` output shape this bar parses: a plist dict including
  `"PID" = N;` when running, a non-zero exit when the label is not loaded. Use it for the fixture
  text rather than inventing a plausible dict.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStatusCmd_ReportsWhatLaunchctlSays` (table-driven, in `cmd/status_test.go`)
     - **Behavior:** `bp status` is a faithful reading of launchd's own answer — a live daemon is
       reported with the pid launchd owns, so the operator can find the process.
     - **Setup:** a fake `launchd.Runner` that records its calls and returns a canned
       `(out, code, err)` per subtest. Three rows, each keyed on `launchctl list <label>`:
       (a) loaded with a pid — the dict text from `automation-notes.md` containing `"PID" = 4242;`,
       code `0`; (b) loaded without a pid — the same dict with the `PID` line removed, code `0`;
       (c) not loaded — empty output, code `113`. Run `runWithLaunchd(t, lc, statusCmdUse)`.
     - **Assertions:** (a) output is exactly `"running (pid 4242)\n"`; (b) and (c) are exactly
       `"not running\n"`; every row returns a `nil` error; the recorded call is
       `launchctl list com.github.b4dmonkey.bit-pro`.
     - **Boundary:** the `PID` key present versus absent inside a dict that parses either way, and
       the exit code at `0` versus the `113` launchctl uses for an unloaded label — the two axes
       that independently mean "not running".
   - [ ] Confirm fails: row (a) fails with output `"not running\n"`, want `"running (pid 4242)\n"` —
         BIT-28.5's hardcoded string cannot vary with the fake's answer.

2. **Implement (GREEN):**
   - [ ] `launchd/launchd.go`: `const Label = "com.github.b4dmonkey.bit-pro"`; the `Runner` type
         above; `ExecRunner` built on `exec.CommandContext(ctx, name, args...).CombinedOutput()`,
         returning the trimmed output plus `exitErr.ExitCode()` via `errors.As(err, &exitErr)` for
         an `*exec.ExitError` and a wrapped error for anything else.
   - [ ] `launchd/launchd.go`: `type State int` with `StateRunning`, `StateNotRunning`,
         `StateStopped` and a `String()` returning `"running"`, `"not running"`, `"stopped"` — the
         scope's three-word vocabulary lives in one place.
   - [ ] `launchd/status.go`: `Status(ctx context.Context, run Runner) (State, int, error)` —
         call `run(ctx, "launchctl", "list", Label)`; propagate a non-nil `err`; return
         `StateNotRunning, 0, nil` when `code != 0`; otherwise match the
         output against a package-level regexp compiled from the Go raw string
         `"PID"\s*=\s*(\d+)` — RE2 supports `\s`, `\d`, and the capture group as written —
         returning `StateRunning` with the parsed pid on a match and `StateNotRunning` otherwise.
   - [ ] `cmd/status.go`: `newStatusCmd(lc launchd.Runner)`; call `launchd.Status`; print
         `running (pid N)` when the state is `StateRunning`, otherwise the state's `String()`.
   - [ ] `cmd/root.go`: `newRootCmd(run claude.Runner, lc launchd.Runner)`; pass `lc` to
         `newStatusCmd`; `NewRootCmd()` calls `newRootCmd(claude.ExecRunner, launchd.ExecRunner)`.
   - [ ] `cmd/cmd_test.go`: give `runWithRunner` a default no-op `launchd.Runner` (returns `"", 113, nil`,
         i.e. nothing loaded) so every existing test keeps compiling and passing unchanged, and add
         `runWithLaunchd(t *testing.T, lc launchd.Runner, args ...string) (string, error)` for the
         tests that need to drive it.

## Claude verifies
- [ ] `just test` — including that the pre-existing `cmd` tests still pass after the
      `newRootCmd` signature change
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. The real `launchctl` is exercised on BIT-28.12.

## Commit (user)
`feat(status): report the running state and pid from launchctl list`