---
id: BIT-2.13
title: Contradiction forces interactive confirmation
status: done
phase: 4
phase_label: Delete
---
## Step 13 (Phase 4 — Delete) — Contradiction forces interactive confirmation
**Status:** ✅ Done — verified 2026-07-16
Step 12 returns a plain error when `--yes` is omitted — it never reads a response. A
test that omits `--yes`, feeds a confirmation via stdin, and expects the file removed
can't pass against that error-only stub, forcing real prompting. A second subtest
(declining) proves the "no" branch is a clean no-op, not an error — this is the scope's
actual "warns/confirms before removing anything" requirement.

**Scope:**
- `cmd/task_delete.go` — replace the placeholder error with an interactive prompt

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskDeleteCmd_PromptsForConfirmation` (table-driven, two subtests —
     matching the style of `cmd/init_test.go`'s `TestInitCmd`)
     - **Behavior:** proves the delete confirmation actually gates on the user's
       answer, not just on whether a prompt was shown.
     - **Setup:** subtest `"confirms"`: `init --prefix BIT`; `task create "X"
       --description "..."`; fresh `rootCmd` with `SetIn(strings.NewReader("y\n"))`,
       `SetArgs([]string{"task", "delete", "BIT-1"})`. Subtest `"declines"`: identical
       setup but `SetIn(strings.NewReader("n\n"))`.
     - **Assertions:** `"confirms"` — `err` is `nil`; `.bit/tasks/BIT-1.md` no longer
       exists. `"declines"` — `err` is `nil` (declining isn't an error condition);
       `.bit/tasks/BIT-1.md` still exists.
     - **Boundary:** the user's response — `"y"` vs. any other input — proves both
       branches of the confirm gate, not just that a prompt was printed.
   - [x] Confirm fails: both subtests get the placeholder error from Step 12
     (`"confirmation required..."`) since nothing reads `cmd.InOrStdin()` yet.

2. **Implement (GREEN):**
   - [x] In `cmd/task_delete.go`'s `RunE`, when `yes` is false: print a confirmation
     prompt naming the task ID to `cmd.OutOrStdout()`; read one line from
     `bufio.NewReader(cmd.InOrStdin())`; trim and lowercase it; if it equals `"y"` or
     `"yes"`, `os.Remove(taskPath(args[0]))`; otherwise print a "cancelled" message and
     `return nil` (declining is not an error).

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] the confirmation prompt's wording is clear enough that a typo'd ID doesn't get
  destroyed by accident — this is the scope's explicit reason for requiring
  confirmation at all

**Commit (user):** `feat(task-crud): task delete prompts for interactive confirmation`

---

Once Step 13 lands, Phases 1–4 are done: `bit init` sets up a project end to end, and
`bit task create/list/read/update/delete` all work against real files in `.bit/tasks/`.
That satisfies every phase's acceptance bar except Phase 5, which stays deliberately
unplanned until a fresh bit_plan pass can design its frontmatter against real task
records these phases produced.