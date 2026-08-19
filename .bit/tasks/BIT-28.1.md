---
id: BIT-28.1
title: bp serve exists and stops when its context is cancelled
status: done
phase: 1
phase_label: serve
---
## **Verse 1**

`bp serve` exists as documented top-level surface and its body runs until its context is cancelled.
Nothing ticks yet — this is the walking skeleton the plist will later point at, and it fails cheap
if the command can't run at all.

## Scope
- `cmd/serve.go` — new; `newServeCmd()` returning the Cobra command
- `cmd/root.go` — register it on `rootCmd`
- `cmd/cmd_test.go` — add a `runWithContext` helper; the existing `runWithRunner` calls
  `root.Execute()`, which hardcodes `context.Background()` and so can never be cancelled

## TDD cycle

1. **Write test (RED):** new `cmd/serve_test.go`
   - [ ] `TestServeCmd_ReturnsWhenContextCancelled`
     - **Behavior:** the daemon body is context-driven — cancelling the command's context ends the
       run and reports success, not an error. This is the mechanism that makes launchd's `SIGTERM`
       a clean exit once BIT-28.4 wires the signal to it.
     - **Setup:** `ctx, cancel := context.WithCancel(t.Context())`; call `cancel()` *before*
       running, so the test needs no sleeping and cannot flake; then
       `runWithContext(t, ctx, serveCmdUse)`.
     - **Assertions:** returned error is `nil`; captured output is `""`.
     - **Boundary:** the context's live/cancelled state — this exercises the already-cancelled
       edge, the earliest point at which the body can be asked to stop.
   - [ ] `TestServeCmd_IsListedInHelp`
     - **Behavior:** `bp serve` is documented surface, not a hidden command — an operator can
       discover it from `bp --help`.
     - **Setup:** `run(t, "--help")`.
     - **Assertions:** output contains `"serve"`.
     - **Boundary:** the command's `Hidden` field in its false state.
   - [ ] Confirm fails: `unknown command "serve" for "bp"` on both tests.

2. **Implement (GREEN):**
   - [ ] `cmd/cmd_test.go`: add
         `runWithContext(t *testing.T, ctx context.Context, args ...string) (string, error)` —
         same body as `runWithRunner` but calling `root.ExecuteContext(ctx)` instead of
         `root.Execute()`.
   - [ ] `cmd/serve.go`: `const serveCmdUse = "serve"`; `newServeCmd() *cobra.Command` with
         `Use: serveCmdUse`, `Args: cobra.NoArgs`, a `Short` describing it as running the
         automation loop in the foreground, and a `RunE` of
         `<-cmd.Context().Done(); return nil`. Leave `Hidden` at its zero value.
   - [ ] `cmd/root.go`: `rootCmd.AddCommand(newServeCmd())`, keeping the existing alphabetical
         ordering of the `AddCommand` block (after `newInstructionsCmd()`).

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(serve): add bp serve as a context-driven foreground command`