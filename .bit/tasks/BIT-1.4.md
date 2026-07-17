---
id: BIT-1.4
title: '`bit init` creates `.bit/`'
status: done
phase: 2
phase_label: Bootstrap
---
## Step 4 (Phase 2 — Bootstrap) — `bit init` creates `.bit/`
**Status:** ✅ Done — verified 2026-07-15
Adds `init` as a subcommand on the root command built in Steps 1–3. Forced because no test
so far has demanded an `init` subcommand exist — running `bit init` today returns Cobra's
"unknown command" error. Both the fresh-directory and already-initialized cases are written
up front as one table-driven test, since the real implementation (`os.MkdirAll`) is
idempotent by construction and satisfies both without contradiction between them.

**Scope:**
- `cmd/init.go` — new, `newInitCmd() *cobra.Command`
- `cmd/init_test.go` — new
- `cmd/root.go` — register `newInitCmd()` on the root command

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestInitCmd` (table-driven, two subtests) in `cmd/init_test.go`
     - **Behavior:** proves `bit init` creates the `.bit/` directory — the Phase 2 finish
       line and the on-disk home future commands (task CRUD) will read and write — and
       that it's safe to run more than once.
     - **Setup:** subtest `"fresh directory"`: `dir := t.TempDir()`; `t.Chdir(dir)`;
       `rootCmd := NewRootCmd()`; `rootCmd.SetArgs([]string{"init"})`;
       `err := rootCmd.Execute()`. Subtest `"already initialized"`: identical setup, but
       call `rootCmd.Execute()` a second time before asserting.
     - **Assertions:** both subtests — `err` is `nil`; `os.Stat(filepath.Join(dir, ".bit"))`
       succeeds and `.IsDir()` is `true`.
     - **Boundary:** whether `.bit/` exists before the run — 0 (doesn't exist) vs. 1
       (already created by a prior `init`) — proves re-running `init` doesn't error, not
       just that the first run works.
   - [x] Confirm fails (both subtests) — for a slightly different reason than predicted:
     `rootCmd.Execute()` returns `nil` (not `unknown command "init" for "bit"`) and prints
     help text, because the root command has no `RunE` and Cobra falls back to printing
     help for unrecognized positional args rather than erroring. `.bit/` is still never
     created either way — no `init` subcommand is registered yet — so the test fails for
     the right underlying reason.

2. **Implement (GREEN):**
   - [x] `cmd/init.go`:
     ```go
     package cmd

     import (
         "os"

         "github.com/spf13/cobra"
     )

     func newInitCmd() *cobra.Command {
         return &cobra.Command{
             Use:   "init",
             Short: "Create the .bit/ directory bit uses to track this project",
             RunE: func(cmd *cobra.Command, args []string) error {
                 return os.MkdirAll(".bit", 0o755)
             },
         }
     }
     ```
   - [x] In `NewRootCmd()` (`cmd/root.go`), add `rootCmd.AddCommand(newInitCmd())` before
     returning.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes
- [x] `just build`, then running `./bin/bit init` from a scratch temp directory creates
  `.bit/` there too — confirms the built binary matches the in-process test behavior, same
  as Step 3 did for `--help`/`--version`

**User verifies:**
- [x] none — matches the scope's Phase 2 acceptance bar exactly (create `.bit/`, nothing
  about what goes inside it, which is deliberately deferred to the CRUD scope)

**Commit (user):** `feat(bootstrap): add bit init command to create .bit/`

---

Once Step 4 lands, both phases of the bootstrap scope are done — `bit` is a real,
installable command, and `bit init` gives future scopes (task CRUD, list UI, board UI) a
real `.bit/` directory to build against.