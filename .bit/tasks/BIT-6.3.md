---
id: BIT-6.3
title: '`--body` emits the body only'
status: done
phase: 2
phase_label: Round-trip a body cleanly
---
## Step 3 (Phase 2 — Round-trip a body cleanly) — `--body` emits the body only
**Status:** ✅ Done — verified 2026-07-20

A read-modify-write refine (edit a scope body, tick a phase box) has to peel the summary
header off `read`'s output first. This adds a `--body` flag that emits just `t.Body`, so a
caller can `read --body` → edit → `update -d`. Additive: the default output is unchanged.

**Scope:**
- `cmd/task_read.go` — add a `bool` `--body` flag; when set, print only `t.Body` and return before the header block.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskReadCmd_BodyOnly` in `cmd/task_read_test.go`
     - **Behavior:** `--body` yields the exact stored body with no summary header and no leading blank line, so the output round-trips back through `update -d`.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Full details", "Line one.\nLine two.")`; `out := mustRun(t, "task", "read", "BIT-1", "--body")`.
     - **Assertions:** `out == "Line one.\nLine two."` (no `BIT-1\ttodo\t...` prefix, no `\n\n`).
     - **Boundary:** a non-empty multi-line body — the normal case the refine loop runs on.
   - [x] `TestTaskReadCmd_BodyOnlyEmpty` in `cmd/task_read_test.go`
     - **Behavior:** an empty body yields empty output, not a stray header or newline.
     - **Setup:** `initProject(t, "BIT")`; `mustRun(t, "task", "create", "No body", "-d", "")`; `out := mustRun(t, "task", "read", "BIT-1", "--body")`.
     - **Assertions:** `out == ""`.
     - **Boundary:** empty-string body — the lower bound; proves `--body` never emits the header fallback.
   - [x] Confirm fails: `--body` is an unknown flag, so `run` returns a non-null error and `mustRun` fails the test.

2. **Implement (GREEN):**
   - [x] Convert `newTaskReadCmd` to a local `var bodyOnly bool` + `cmd.Flags().BoolVar(&bodyOnly, "body", false, "print only the task body, without the summary header")`.
   - [x] In `RunE`, after `Load`, if `bodyOnly { fmt.Fprint(out, t.Body); return nil }` before the existing header block.

**Claude verifies:**
- [x] `just test` — new tests pass; `TestTaskReadCmd_ShowsFullTask` / `ShowsPhase` / `OmitsPhaseWhenAbsent` stay green (default output unchanged).
- [x] `just lint`

**User verifies:**
- [x] `--body` is the right flag name (vs. `--raw`); body-with-no-trailing-newline matches what `update -d "$(...)"` expects to write back.

**Commit (user):** `feat(task): add task read --body for header-free output`