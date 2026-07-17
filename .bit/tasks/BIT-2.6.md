---
id: BIT-2.6
title: '`bit task list` shows all tasks'
status: done
phase: 2
phase_label: List & read
---
## Step 6 (Phase 2 — List & read) — `bit task list` shows all tasks
**Status:** ✅ Done — verified 2026-07-15
Adds `bit task list`. Unlike Step 4, there's no honest hardcoded stand-in here — the
first test seeds two distinct tasks and asserts both appear, which requires real
iteration and parsing from the start.

**Scope:**
- `cmd/task_list.go` — new: `newTaskListCmd()`
- `cmd/task_list_test.go` — new
- `cmd/task.go` — `taskCmd.AddCommand(newTaskListCmd())`

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskListCmd_ShowsAllTasks` (in `cmd/task_list_test.go`)
     - **Behavior:** proves a user can see every task at a glance, per the scope's
       Phase 2 goal.
     - **Setup:** `init --prefix BIT`; `task create "First" --description "..."`;
       `task create "Second" --description "..."`; fresh `rootCmd` with
       `SetOut(buf)`, `SetArgs([]string{"task", "list"})`.
     - **Assertions:** `buf.String()` contains `"BIT-1"`, `"First"`, `"BIT-2"`, and
       `"Second"`; `BIT-1`'s line appears before `BIT-2`'s (ascending ID order).
     - **Boundary:** two existing tasks (N=2) — proves iteration over multiple files,
       not a single-entry special case.
   - [x] Confirm fails: `unknown command "list" for "bit task"`.

2. **Implement (GREEN):**
   - [x] `cmd/task_list.go`: `newTaskListCmd()`, no args. `RunE`: `filepath.Glob(".bit/tasks/*.md")`,
     `slices.Sort(matches)` (stdlib `slices` package, Go 1.21+ — not `sort.Slice`) so
     output is in ascending ID order, for each file read + split on `---` +
     `yaml.Unmarshal` the header into a minimal local struct (`ID`, `Title`, `Status`),
     print one line per task (e.g. `fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", id,
     status, title)`).

3. **More tests (RED → GREEN):**
   - [x] `TestTaskListCmd_EmptyWhenNoTasks`
     - **Behavior:** proves an empty project doesn't error or panic.
     - **Setup:** `init --prefix BIT` only, no tasks created; run `task list`.
     - **Assertions:** `err` is `nil`; no panic (this may already pass under the Step 6
       implementation — `filepath.Glob` on a directory with no matches returns an empty
       slice, not an error — but it's worth asserting explicitly as a regression guard).
     - **Boundary:** zero tasks — the lower bound for the iteration loop.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] the list's column layout (`id`, `status`, `title`) is a reasonable MVP shape —
  matches the terse tabular style noted as a reference point for `bit`'s CLI output

**Commit (user):** `feat(task-crud): add task list command`