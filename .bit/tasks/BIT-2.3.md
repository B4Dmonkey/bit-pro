---
id: BIT-2.3
title: Contradiction forces empty-prefix validation
status: done
phase: 1
phase_label: Init wizard + create
---
## Step 3 (Phase 1 — Init wizard + create) — Contradiction forces empty-prefix validation
**Status:** ✅ Done — verified 2026-07-15

Neither Step 1 nor Step 2 rejects an empty result — a blank `--prefix ""` or a blank
line at the prompt currently writes `Config{Prefix: ""}`, which would later produce
malformed task filenames like `-1.md`. A test asserting this is refused can't pass
against the current silently-accepting implementation.

**Scope:**
- `cmd/init.go` — return an error when the resolved prefix (flag or prompt) is empty
  after trimming

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestInitCmd_ErrorsOnEmptyPrefix` (in `cmd/init_test.go`)
     - **Behavior:** proves the wizard can't leave a project in a state where task IDs
       would be malformed.
     - **Setup:** `dir := t.TempDir()`; `t.Chdir(dir)`; `rootCmd := NewRootCmd()`;
       `rootCmd.SetIn(strings.NewReader("\n"))`; `rootCmd.SetArgs([]string{"init"})`;
       `rootCmd.Execute()`.
     - **Assertions:** `err` is not `nil`; `config.toml` does **not** exist (validation
       happens before the write).
     - **Boundary:** trimmed prefix length == 0 — the lower bound, whether reached via
       an empty flag or a blank prompt response.
   - [x] Confirm fails: `err` is `nil` and `config.toml` exists with `Prefix: ""` — Step
     2's implementation writes whatever it resolved, empty or not.

2. **Implement (GREEN):**
   - [x] In `cmd/init.go`'s `RunE`, after resolving `prefix` (flag or prompt) and
     trimming, if it's still `""`, `return fmt.Errorf("task ID prefix cannot be empty")`
     before calling `saveConfig`.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] none — pure validation, matches the scope's implicit requirement that the wizard
  produce a usable config

**Commit (user):** `feat(task-crud): init rejects an empty task ID prefix`