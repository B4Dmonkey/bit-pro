---
id: BIT-1.2
title: Contradiction forces `--version` support
status: done
phase: 1
phase_label: Bootstrap
---
## Step 2 (Phase 1 — Bootstrap) — Contradiction forces `--version` support
**Status:** ✅ Done — verified 2026-07-15
Step 1's `rootCmd` never sets `Version`, so Cobra never registers a `--version` flag —
running it returns an `unknown flag` error. A test asserting real version output can't
pass against Step 1's command, forcing a real `Version` field.

**Scope:**
- `cmd/root.go` — add `Version` to the command literal

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestRootCmd_Version` (in `cmd/root_test.go`)
     - **Behavior:** proves `bit --version` (the other half of the scope's Phase 1
       acceptance criteria) actually works, not just `--help`.
     - **Setup:** `rootCmd := NewRootCmd()`; `buf := &bytes.Buffer{}`;
       `rootCmd.SetOut(buf)`; `rootCmd.SetArgs([]string{"--version"})`; call
       `rootCmd.Execute()`.
     - **Assertions:** `Execute()` returns `nil`; `buf.String() == "bit version 0.1.0-dev\n"`
       — this is Cobra's verified default version template
       (`{{.DisplayName}} version {{.Version}}`), not a guess: confirmed by reading
       `defaultVersionTemplate` in `cobra@v1.10.2/command.go`.
     - **Boundary:** the `--version` flag path, distinct from `--help` — proves a second,
       independently-registered flag works, not just the first one Cobra sets up for free.
   - [x] Confirm fails: `Execute()` returns a non-nil error (`unknown flag: --version`),
     because Step 1's command has no `Version` set, so Cobra never registers the flag.

2. **Implement (GREEN):**
   - [x] In `cmd/root.go`, add a `const version = "0.1.0-dev"` and set `Version: version`
     on the `cobra.Command` literal in `NewRootCmd()`.

**Claude verifies:**
- [x] `go test ./...` passes

**User verifies:**
- [x] the version string (`0.1.0-dev`) is an acceptable placeholder for now

**Commit (user):** `feat(bootstrap): add --version output to root command`