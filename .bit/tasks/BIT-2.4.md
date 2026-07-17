---
id: BIT-2.4
title: '`bit task create` writes the first task'
status: done
phase: 1
phase_label: Init wizard + create
---
## Step 4 (Phase 1 — Init wizard + create) — `bit task create` writes the first task
**Status:** ✅ Done — verified 2026-07-15
Adds the `task` command group and `create` subcommand. This is the other half of the
walking skeleton — the scope's Phase 1 finish line is a durably created task. ID
assignment is hardcoded to `1` for now (nothing yet demands otherwise); Step 5
contradicts that.

**Scope:**
- `go.mod` / `go.sum` — add `gopkg.in/yaml.v3@v3.0.1`
- `cmd/task.go` — new: `newTaskCmd()` (parent `task` command, no `RunE`, just groups
  subcommands)
- `cmd/task_create.go` — new: `newTaskCreateCmd()`
- `cmd/task_create_test.go` — new
- `cmd/root.go` — `rootCmd.AddCommand(newTaskCmd())`

**Task file format:** `.bit/tasks/<id>.md`, YAML frontmatter delimited by `---` lines,
then the body:
```
---
id: BIT-1
title: Set up init wizard
status: todo
---
Add flags for prefix capture.
```
Default `status` is `"todo"` — the scope explicitly defers validating status
*transitions* to the next scope, but the field itself has to exist and start somewhere.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskCreateCmd_WritesFirstTask` (in `cmd/task_create_test.go`)
     - **Behavior:** proves a task can be durably created and read back from disk — the
       scope's stated walking-skeleton finish line ("nothing else in this scope has
       anything to operate on until a task can be durably created").
     - **Setup:** `dir := t.TempDir()`; `t.Chdir(dir)`; run `init --prefix BIT` via a
       fresh `rootCmd` (real end-to-end setup, matching how `cli-bootstrap-plan.md`'s
       Step 4 chained real commands); then a second fresh `rootCmd` with
       `SetArgs([]string{"task", "create", "Set up init wizard", "--description", "Add
       flags for prefix capture."})`.
     - **Assertions:** `err` is `nil`; `.bit/tasks/BIT-1.md` exists; its content, run
       through the same `---`-split parsing Step 7 will formalize, gives
       `id: BIT-1`, `title: Set up init wizard`, `status: todo`, and a body containing
       `"Add flags for prefix capture."`.
     - **Boundary:** first task in a fresh `.bit/tasks/` — zero existing tasks, the
       lower bound for ID assignment.
   - [x] Confirm fails: `unknown command "task" for "bit"`.

2. **Implement (GREEN):**
   - [x] `cmd/task.go`:
     ```go
     package cmd

     import "github.com/spf13/cobra"

     const tasksDir = ".bit/tasks"

     func newTaskCmd() *cobra.Command {
         taskCmd := &cobra.Command{
             Use:   "task",
             Short: "Manage tasks",
         }
         taskCmd.AddCommand(newTaskCreateCmd())
         return taskCmd
     }
     ```
   - [x] `cmd/task_create.go`: `newTaskCreateCmd()` with `Args: cobra.ExactArgs(1)` and a
     `--description`/`-d` string flag. `RunE`: `loadConfig()` for the prefix;
     `os.MkdirAll(tasksDir, 0o755)`; hardcode `id := cfg.Prefix + "-1"`; write
     `.bit/tasks/<id>.md` with frontmatter (`id`, `title` from `args[0]`,
     `status: "todo"`) + body from the description flag.
   - [x] `cmd/root.go`: `rootCmd.AddCommand(newTaskCmd())`.

3. **More tests (RED → GREEN):**
   - [x] `TestTaskCreateCmd_ErrorsWithoutTitle`
     - **Behavior:** proves the title argument is required, not silently optional.
     - **Setup:** after `init --prefix BIT`, run `task create` with no positional args.
     - **Assertions:** `err` is not `nil` (Cobra's `ExactArgs(1)` usage error).
     - **Boundary:** zero positional args — below the required minimum of one.
   - [x] `TestTaskCreateCmd_ErrorsWithoutConfig`
     - **Behavior:** proves `task create` fails clearly when run before `init`, instead
       of writing a task with an empty/garbage ID.
     - **Setup:** fresh temp dir, no `init` run; run `task create "Foo"`.
     - **Assertions:** `err` is not `nil`; `.bit/tasks/` is not created.
     - **Boundary:** `config.toml` absent — the zero-setup case.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] the frontmatter shape (`id`/`title`/`status` + body) is an acceptable minimal
  starting point

**Commit (user):** `feat(task-crud): add task create command`