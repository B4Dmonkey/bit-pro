---
id: BIT-28.4
title: A termination signal cancels the root context
status: todo
approved: true
phase: 1
phase_label: serve
---
## **Verse 1**

`Ctrl-C` and `SIGTERM` cancel the root context, so `bp serve` exits cleanly instead of being killed.
Forced by BIT-28.1: the body stops when its context is cancelled, but in a real run nothing ever
cancels it — `main.go` calls `Execute()`, which hardcodes `context.Background()`.

## Scope
- `cmd/root.go` — add `signalContext()` and an exported `Execute() error`
- `main.go` — call `cmd.Execute()` instead of `cmd.NewRootCmd().Execute()`
- `cmd/root_test.go` — the new test

`SIGTERM` matters because it is exactly what `launchctl bootout` sends the daemon in Verse 2;
`os.Interrupt` is the `Ctrl-C` half that Verse 1's done-when names.

## TDD cycle

1. **Write test (RED):** `cmd/root_test.go`
   - [ ] `TestSignalContext_CancelsOnTerminationSignals` (table-driven over
         `syscall.SIGTERM` and `syscall.SIGINT`)
     - **Behavior:** an interrupt or a termination request reaches the running command as a
       context cancellation, which is the one thing `bp serve` already knows how to act on.
     - **Setup:** `ctx, stop := signalContext()`; `defer stop()` — `signal.NotifyContext` returns
       only after the handler is registered, so sending the signal next cannot race it. Then
       `syscall.Kill(syscall.Getpid(), tt.sig)`.
     - **Assertions:** `select` on `<-ctx.Done()` versus `<-time.After(2 * time.Second)`; the
       timeout branch calls `t.Fatal`.
     - **Boundary:** both signals in the notified set — `SIGTERM` is what launchd sends,
       `SIGINT` is `Ctrl-C`, and missing either leaves half of Verse 1's done-when unmet.
   - [ ] Confirm fails: `undefined: signalContext`. Note the other failure shape: if
         `signalContext` is later written without registering one of these signals, that signal's
         default action kills the test binary and the whole package run reports
         `signal: terminated`. That crash *is* the red for this test — it is not an unrelated
         breakage.

2. **Implement (GREEN):**
   - [ ] `cmd/root.go`: `func signalContext() (context.Context, context.CancelFunc)` returning
         `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`.
   - [ ] `cmd/root.go`: `func Execute() error` — `ctx, stop := signalContext()`, `defer stop()`,
         `return NewRootCmd().ExecuteContext(ctx)`.
   - [ ] `main.go`: replace `cmd.NewRootCmd().Execute()` with `cmd.Execute()`, keeping the existing
         `Error:` printing and `os.Exit(1)`. `NewRootCmd` stays exported and unchanged, so nothing
         else that builds a root command is affected.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `go build ./...` — `main.go` is not covered by any test, so the build is what proves the
      call-site change compiles.

## User verifies
- [ ] Whole slice: `just install`, run `bp serve -v`, watch at least one `tick` line appear, then
      press `Ctrl-C` — it returns to the prompt immediately, prints no panic or stack trace, and
      `echo $?` is `0`. This is Verse 1's capability end to end: a daemon an operator can run in
      their terminal and watch.

## Commit (user)
`feat(cli): cancel the root context on SIGINT and SIGTERM`