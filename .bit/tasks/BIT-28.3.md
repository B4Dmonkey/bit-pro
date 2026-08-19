---
id: BIT-28.3
title: Log encoding follows the output writer
status: done
phase: 1
phase_label: serve
---
## **Verse 1**

The log handler is chosen from the writer: text when it is a character device (a terminal),
JSON otherwise. Contradicts BIT-28.2, which hardcoded the text handler — under launchd stdout is
`daemon.log`, a regular file, and the same binary has to write JSON there.

## Scope
- `cmd/serve.go` — extract `newHandler(w io.Writer, level slog.Level) slog.Handler`
- `cmd/serve_test.go` — the new tests

Stdlib only: check `w.(*os.File)` and then `f.Stat()` for `info.Mode()&os.ModeCharDevice != 0`.
`golang.org/x/term` is not a direct dependency of this module and is not worth adding for one
predicate.

The rule is "character device", so `/dev/null` tests the true branch without needing a pty. That is
a harmless false positive of the predicate — nothing writes a daemon log to `/dev/null` — and it is
what makes the branch testable at all. A real terminal is the user-verify below.

## TDD cycle

1. **Write test (RED):** `cmd/serve_test.go`
   - [ ] `TestServeCmd_LogsJSONWhenOutputIsNotATerminal`
     - **Behavior:** under launchd, stdout is redirected to `daemon.log`, so the daemon's own log
       is machine-readable for anything that later reads it back.
     - **Setup:** same fast-`serveTick` harness as BIT-28.2, run with `"-v"` so ticks are emitted
       into the test's `bytes.Buffer`.
     - **Assertions:** every non-empty line of the output unmarshals with `json.Unmarshal` into a
       `map[string]any`; at least one has `["msg"] == "tick"` and `["level"] == "DEBUG"`.
     - **Boundary:** a writer that is not an `*os.File` at all — the far end of the predicate,
       where there is no file descriptor to ask about.
   - [ ] `TestNewHandler_PicksEncodingFromTheWriter` (table-driven, calling `newHandler` directly)
     - **Behavior:** the encoding follows the destination, not a flag — so the same binary is
       readable live and parseable in a log file with nothing to configure.
     - **Setup:** three writers — a `*bytes.Buffer`; a regular file from `os.CreateTemp(t.TempDir(), "")`;
       and `os.Open("/dev/null")` (a character device). Close the files with `t.Cleanup`.
     - **Assertions:** type-assert the result — `*slog.JSONHandler` for the buffer and the regular
       file, `*slog.TextHandler` for `/dev/null`.
     - **Boundary:** the `os.ModeCharDevice` bit in both states, plus the non-`*os.File` case where
       `Stat` cannot be called at all.
   - [ ] Confirm fails: `undefined: newHandler`, and the command-level test fails on
         `json.Unmarshal` rejecting `time=… level=DEBUG msg=tick`.

2. **Implement (GREEN):**
   - [ ] `cmd/serve.go`: add `newHandler(w io.Writer, level slog.Level) slog.Handler` — if
         `f, ok := w.(*os.File); ok`, and `info, err := f.Stat()` succeeds, and
         `info.Mode()&os.ModeCharDevice != 0`, return `slog.NewTextHandler(w, opts)`; otherwise
         return `slog.NewJSONHandler(w, opts)`, where `opts` is `&slog.HandlerOptions{Level: level}`.
   - [ ] `cmd/serve.go`: `RunE` builds its logger with `slog.New(newHandler(cmd.OutOrStdout(), level))`
         instead of calling `slog.NewTextHandler` directly.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] `just install`, then run `bp serve -v` in your terminal: `msg=started` at `INFO`, then
      `time=… level=DEBUG msg=tick` every 10 seconds — text, not JSON. `Ctrl-C` still leaves the
      shell usable. (The terminal branch is the one case no automated test can reach without a pty.)
      No `msg=stopped` on Ctrl-C yet — nothing cancels the root context until BIT-28.4.

## Commit (user)
`feat(serve): pick the log encoding from the output writer`