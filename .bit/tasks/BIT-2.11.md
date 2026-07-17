---
id: BIT-2.11
title: Contradiction forces `--description` and `--status`
status: done
phase: 3
phase_label: Update
---
## Step 11 (Phase 3 — Update) — Contradiction forces `--description` and `--status`
**Status:** ✅ Done — verified 2026-07-16
Step 10's command only registers a `--title` flag — passing `--description` or
`--status` today is `unknown flag`, not a silent no-op. A test using both at once can't
pass without wiring them in independently.

**Scope:**
- `cmd/task_update.go` — add `--description`/`-d` and `--status`/`-s` flags

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskUpdateCmd_ChangesDescriptionAndStatus`
     - **Behavior:** proves each field flag is wired independently, not as a single
       hardcoded title-only path.
     - **Setup:** `init --prefix BIT`; `task create "Keep this title" --description
       "Old body."`; run `task update BIT-1 --description "New body." --status "doing"`.
     - **Assertions:** `err` is `nil`; `loadTask("BIT-1")` gives `Title: "Keep this
       title"` (unchanged), `Body` contains `"New body."`, `Status: "doing"`.
     - **Boundary:** two field flags changed simultaneously, title flag absent — proves
       the three `Changed()` checks are independent of each other.
   - [x] Confirm fails: `unknown flag: --description`.

2. **Implement (GREEN):**
   - [x] Add `--description`/`-d` and `--status`/`-s` string flags to
     `newTaskUpdateCmd()`; in `RunE`, mirror the title check for each:
     `if cmd.Flags().Changed("description") { t.Body = descriptionFlag }`, same for
     status.

3. **More tests (RED → GREEN):**
   - [x] `TestTaskUpdateCmd_NoFlagsIsANoOp`
     - **Behavior:** proves calling `update` with no field flags doesn't error or
       corrupt the file.
     - **Setup:** `init --prefix BIT`; `task create "Title" --description "Body."`; run
       `task update BIT-1` with no flags.
     - **Assertions:** `err` is `nil`; `loadTask("BIT-1")` is unchanged from creation.
     - **Boundary:** zero fields changed — the lower bound for the `Changed()` checks.
   - [x] `TestTaskUpdateCmd_ErrorsOnUnknownID`
     - **Behavior:** proves updating a nonexistent task fails clearly (this should
       already pass via `loadTask`'s propagated `os.ReadFile` error — asserted as a
       regression guard).
     - **Setup:** `init --prefix BIT` only; run `task update BIT-99 --title "X"`.
     - **Assertions:** `err` is not `nil`.
     - **Boundary:** an ID with no corresponding file.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] allowing free-form `--status` values here (no enum/validation) is acceptable for
  this scope — the state-machine validation is explicitly deferred to the next scope

**Commit (user):** `feat(task-crud): task update supports description and status`