---
id: BIT-1.1
title: Root command answers `--help`
status: done
phase: 1
phase_label: Bootstrap
---
## Step 1 (Phase 1 — Bootstrap) — Root command answers `--help`
**Status:** ✅ Done — verified 2026-07-15
Stands up the Go module and a Cobra root command named `bit`; this is the walking skeleton
everything else hangs off of. Forced by nothing prior — it's the first test.

**Scope:**
- `go.mod` — new module, `github.com/B4Dmonkey/bit-pro`
- `cmd/root.go` — new, `NewRootCmd() *cobra.Command`
- `cmd/root_test.go` — new
- `main.go` — new, calls `cmd.NewRootCmd().Execute()`

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestRootCmd_Help` (in `cmd/root_test.go`)
     - **Behavior:** proves `bit` is wired as a real Cobra command that answers `--help`
       with usage text — the minimum needed for `bit --help` to work per the scope.
     - **Setup:** `rootCmd := NewRootCmd()`; `buf := &bytes.Buffer{}`;
       `rootCmd.SetOut(buf)`; `rootCmd.SetArgs([]string{"--help"})`; call `rootCmd.Execute()`.
     - **Assertions:** `Execute()` returns `nil`; `buf.String()` contains `"bit"` (the
       command's `Short` text). No `"Usage:"` assertion: Cobra's default help template only
       emits the `Usage:` block when `Runnable()` or `HasSubCommands()` is true (verified
       against `cobra@v1.10.2/command.go`), and this step's root command has neither a
       `RunE` nor subcommands yet — asserting on it would force scope creep (a no-op `RunE`
       or a dummy subcommand) the step doesn't otherwise need.
     - **Boundary:** the `--help` flag path — the only supported invocation at this step,
       since no subcommands are registered yet.
   - [x] Confirm fails: compile error — `undefined: NewRootCmd` (package `cmd` doesn't
     exist yet).

2. **Implement (GREEN):**
   - [x] `go mod init github.com/B4Dmonkey/bit-pro` (go directive: `go 1.26.5`, matching the
     installed toolchain exactly)
   - [x] `go get github.com/spf13/cobra@v1.10.2`
   - [x] `cmd/root.go`:
     ```go
     package cmd

     import "github.com/spf13/cobra"

     func NewRootCmd() *cobra.Command {
         return &cobra.Command{
             Use:   "bit",
             Short: "bit is a project-management CLI for LLM-driven development workflows",
         }
     }
     ```
   - [x] `main.go`:
     ```go
     package main

     import (
         "os"

         "github.com/B4Dmonkey/bit-pro/cmd"
     )

     func main() {
         if err := cmd.NewRootCmd().Execute(); err != nil {
             os.Exit(1)
         }
     }
     ```

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] none — pure wiring, nothing judgment-based yet

**Commit (user):** `feat(bootstrap): wire root cobra command with help text`