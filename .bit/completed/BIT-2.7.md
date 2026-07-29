---
id: BIT-2.7
title: '`bit task read <id>` shows one task''s full content'
status: done
phase: 2
phase_label: List & read
---
## Step 7 (Phase 2 — List & read) — `bit task read <id>` shows one task's full content
**Status:** ✅ Done — verified 2026-07-15
Adds `bit task read`. List only ever needed `id`/`title`/`status`; a test asserting the
full multi-line body is shown forces real body-parsing that list never exercised.

**Scope:**
- `cmd/task_read.go` — new: `newTaskReadCmd()`
- `cmd/task_read_test.go` — new
- `cmd/task.go` — `taskCmd.AddCommand(newTaskReadCmd())`

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskReadCmd_ShowsFullTask` (in `cmd/task_read_test.go`)
     - **Behavior:** proves a user can review one task's full content, per the scope's
       Phase 2 goal.
     - **Setup:** `init --prefix BIT`; `task create "Full details" --description
       "Line one.\nLine two."`; fresh `rootCmd` with `SetOut(buf)`,
       `SetArgs([]string{"task", "read", "BIT-1"})`.
     - **Assertions:** `buf.String()` contains `"BIT-1"`, `"Full details"`, `"todo"`,
       `"Line one."`, and `"Line two."`.
     - **Boundary:** reading content beyond the frontmatter block — proves the `---`
       delimiter parsing correctly separates header from body across multiple lines.
   - [x] Confirm fails: `unknown command "read" for "bit task"`.

2. **Implement (GREEN):**
   - [x] `cmd/task_read.go`: `newTaskReadCmd()`, `Args: cobra.ExactArgs(1)`. `RunE`:
     read `.bit/tasks/<args[0]>.md`, split on `---`, `yaml.Unmarshal` the header,
     print id/title/status + the body to `cmd.OutOrStdout()`.

3. **More tests (RED → GREEN):**
   - [x] `TestTaskReadCmd_ErrorsOnUnknownID`
     - **Behavior:** proves a typo'd ID surfaces as a command error, not a panic or a
       raw stack trace.
     - **Setup:** `init --prefix BIT` only; run `task read BIT-99`.
     - **Assertions:** `err` is not `nil` (this should already pass — `os.ReadFile`'s
       error propagates through `RunE` — but assert it explicitly as a regression
       guard).
     - **Boundary:** an ID with no corresponding file — the not-found edge.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] the read output format (fields then body) is readable enough for MVP — full
  `glow`-style rendering is explicitly out of scope for the CLI (that's TUI-era work)

**Commit (user):** `feat(task-crud): add task read command`