---
id: BIT-23.3
title: task create and update gain the --check flag
status: todo
phase: 2
phase_label: Runnable checks
---
## **Verse 2**

`bp check <ID>` executes the declared check, but there is no way to declare one yet from the CLI. This bar adds `--check` to `task create` and `task update` so the criterion can be written without hand-editing the frontmatter.

## Scope
- `cmd/task_create.go` — add `var check string` + `cmd.Flags().StringVar(&check, "check", "", "shell command that proves this bar is done")`; pass it to `task.Task{Check: check}`
- `cmd/task_update.go` — add `var check string` + flag; apply when `cmd.Flags().Changed("check")`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCreateCmd_WritesCheckField`
     - **Behavior:** `bp task create --check "make test"` persists the check field on the new task
     - **Setup:** `initProject(t, "BIT")`; `mustRun(t, "task", "create", "Track", "--check", "make test")`
     - **Assertions:** `task.New(".bit").Load("BIT-1")` → `got.Check == "make test"`
     - **Boundary:** non-empty check string — proves the flag is wired through to the struct
   - [ ] Confirm fails: `--check` flag does not exist on `task create`; Execute returns "unknown flag"

2. **Implement (GREEN):**
   - [ ] Add `--check` flag to `newTaskCreateCmd()` in `cmd/task_create.go`; include `Check: check` in the `task.Task{}` literal
   - [ ] Add `--check` flag to `newTaskUpdateCmd()` in `cmd/task_update.go`; apply with `if cmd.Flags().Changed("check") { t.Check = check }`

3. **More tests (RED → GREEN):**
   - [ ] `TestTaskUpdateCmd_UpdatesCheckField`
     - **Behavior:** `bp task update --check "go test ./..."` changes an existing check
     - **Setup:** `initProject(t, "BIT")`; create task with `--check "make test"`; `mustRun(t, "task", "update", "BIT-1", "--check", "go test ./...")`
     - **Assertions:** loaded `BIT-1.Check == "go test ./..."`
     - **Boundary:** update changes an existing value — the replacement path; contradicts a hypothetical hardcoded write from the create step
   - [ ] `TestTaskCreateCmd_CheckDefaultsEmpty`
     - **Behavior:** creating without `--check` leaves Check field empty
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Track", "...")`
     - **Assertions:** `got.Check == ""`
     - **Boundary:** flag absent — lower bound; proves omitempty behavior in the struct

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] `bp task create "Prove it" --check "just test" && bp task list` — new task appears in list; `bp task read` shows the body; `bp check <ID>` runs `just test` and exits 0

## Commit (user)
`feat(task): add --check flag to task create and update`