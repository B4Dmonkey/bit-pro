---
id: BIT-2.2
title: Contradiction forces interactive prompting
status: done
phase: 1
phase_label: Init wizard + create
---
## Step 2 (Phase 1 — Init wizard + create) — Contradiction forces interactive prompting
**Status:** ✅ Done — verified 2026-07-15

Step 1 only handles the flag-given path — with `--prefix` omitted, `prefix == ""` and no
config is written at all. A test that omits the flag and still expects a written config
can't pass without real prompting, forcing the wizard's interactive half.

**Scope:**
- `cmd/init.go` — prompt for prefix via stdin when `--prefix` is empty
- `cmd/init_test.go` — new test; **also update the existing `TestInitCmd` table test**
  to pass `--prefix "BIT"` in its `SetArgs` (see note below)

**Why the existing test needs updating:** once `RunE` reads from `cmd.InOrStdin()` when
`--prefix` is empty, `TestInitCmd`'s two subtests (which never call `SetIn`) would fall
into the prompt path and block on `bufio.Reader.ReadString` reading real `os.Stdin` during
`go test` — a hang, not a failure. `TestInitCmd` was only ever testing `.bit/` directory
creation, so keeping it on the flag path with `--prefix "BIT"` preserves its original
intent without exercising the new prompt logic.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestInitCmd_PromptsForPrefixWhenFlagOmitted` (in `cmd/init_test.go`)
     - **Behavior:** proves the interactive half of the wizard — a human running `bit
       init` with no flags still gets prompted and ends up with a valid config, not a
       silently-empty one.
     - **Setup:** `dir := t.TempDir()`; `t.Chdir(dir)`; `rootCmd := NewRootCmd()`;
       `rootCmd.SetIn(strings.NewReader("BIT\n"))`; `rootCmd.SetArgs([]string{"init"})`;
       `rootCmd.Execute()`.
     - **Assertions:** `err` is `nil`; `config.toml` exists and decodes to
       `Config{Prefix: "BIT"}` — same end state as Step 1, reached via stdin instead of
       a flag.
     - **Boundary:** `--prefix` omitted (flag's zero value `""`) — the prompt-fallback
       path.
   - [x] Update `TestInitCmd`'s `SetArgs` to `[]string{"init", "--prefix", "BIT"}` in both
     subtests.
   - [x] Confirm fails: `os.Stat(config.toml)` returns "no such file" — Step 1's `RunE`
     skips `saveConfig` entirely when `prefix == ""`, so nothing is written.

2. **Implement (GREEN):**
   - [x] In `cmd/init.go`'s `RunE`, when `prefix == ""`: `fmt.Fprint(cmd.OutOrStdout(),
     "Task ID prefix: ")`; read one line from `bufio.NewReader(cmd.InOrStdin())`; trim
     whitespace into `prefix`.
   - [x] Write config using the resulting `prefix` (same `saveConfig` call as Step 1, now
     reachable from either path).

**Claude verifies:**
- [x] `go test ./...` passes (including the updated `TestInitCmd`)
- [x] `go vet ./...` passes

**User verifies:**
- [x] the prompt text ("Task ID prefix: ") reads fine for a first-time user

**Commit (user):** `feat(task-crud): init prompts for prefix when --prefix omitted`