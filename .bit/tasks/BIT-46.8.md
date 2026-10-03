---
id: BIT-46.8
title: CLI commands find the project from any subfolder and use its central store
status: todo
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

The CLI's project commands switch from `bitdir.Current()` to the registry all at once. cmd tests create through `bp task create` and read back through `bp feedback add`/`approve` (`cmd/feedback_add_test.go`, `cmd/task_test.go`), so switching some commands and not others would put records in two places. Today a task created from a subfolder can't be listed (`bitdir.Canonical` returns a relative `.bit`), and the RED test below forces the switch. Records keep v1's markdown format.

## Scope
- `project/store.go`: `func OpenStore(ctx context.Context, dir string) (*task.Store, error)`. It runs `Find(ctx, dir)` → `store.ProjectDir(p.Code)` → `task.NewProject(root, p.Code)`. Resolver errors are returned so that `errors.Is` still matches them. **Seam:** this is the one entry point the CLI, the MCP server and the TUI use.
- `cmd/task/task.go`: unexported `openStore(cmd *cobra.Command) (*taskstore.Store, error)` (`os.Getwd` → `project.OpenStore(cmd.Context(), wd)`). It's used by `complete:15, create:42, delete:37, list:19, move:17, read:19, update:19`. It lives in `cmd/task` because that package's test harness builds its own root with no pre-run (`cmd/task/helpers_test.go:21-40`).
- `cmd/approve.go:15,26`, `cmd/feedback_add.go:19`, `cmd/tui.go:24`: the same switch, through a `cmd`-package `openStore(cmd)`. `approve`'s `RunE` takes `cmd` instead of `_`.
- `cmd/root.go`: delete the `PersistentPreRunE` (`:136`). `pluginState` still calls `bitdir.Root()` until a later bar.
- Fixtures: both `initProject`s drop the `SaveConfig` line.
- Test peeks at `task.New(".bit")` change to a `projectStore(t)` helper that calls `project.OpenStore(t.Context(), ".")`. That covers `cmd/task/create_test.go` ×4, `update_test.go` ×6, `helpers_test.go:69` (`approve`), `cmd/approve_test.go:15,32`, the `feedback_add_test.go` path assertions (`.bit/feedback/...` → the store dir's `feedback/...`), `cmd/task/move_test.go`, and the other `cmd/task/*_test.go` files that peek at `.bit`.
- `cmd/root_test.go` → `cmd/task_test.go`: the three `TestBitDir` subtests (converted in BIT-45.1) test `bp task list`, not a `Resolve` in `cmd`, so they move out of `root_test.go` into `TestTaskCmd` with the same expectations (outside a worktree, inside `.claude/worktrees/<name>`, nested worktrees): `TestTaskCmd/outside worktree uses the project`, `TestTaskCmd/inside claude worktree resolves to main checkout`, `TestTaskCmd/nested worktree resolves to outermost checkout`. `TestBitDir` is deleted. The worktree dirs no longer need their own `.bit/`.
- Raw-file seeds and peeks in `cmd/task/*_test.go`, `cmd/task_test.go` (`:61-66`, completed peeks) and `cmd/feedback_add_test.go` move from `.bit/...` to the store dir (`projectStore(t)`'s root). That includes the hand-written frontmatter seeds (`writeRawTask` in `complete_test.go` and `update_test.go`, the raw track in `list_test.go:101`) and the `ReadDir(".bit/...")` counts.
- `TestTaskReadCmd/contains path traversal id` and `TestTaskDeleteCmd/contains path traversal id` write their `README.md` fixture where `../../README` lands from the store's `tasks/` dir (`<data>/bit/README.md`), not in `dir`, so they still guard against an escape that would hit a real file.
- `TestTaskCreateCmd/errors without config` chdirs to a bare temp dir and would now open the real registry: it sandboxes `HOME` and `XDG_DATA_HOME` first (same assertions: an error, and no `.bit/tasks`).
- Delete `TestTaskCreateCmd/uppercases a corrupt prefix on read`: it writes a lowercase prefix into `.bit/config.toml`, which the CLI no longer reads (the code comes from the registry and is validated on `bp add`). Removal, so no replacement test.
- `add`, `list`, help and `--version` don't resolve (unchanged behavior, so no new test).

## TDD cycle

0. **Convert existing tests (pure restructure, same cases and assertions, still green).** Each flat test becomes a subtest of its unit's test, named by its suffix in lowercase words; existing tables keep their rows. `cmd/task_test.go` and `cmd/root_test.go` were already converted in BIT-45.1.
   - [ ] `cmd/approve_test.go`: `TestApproveCmd/{sets approved true, unapprove clears, errors on unknown id}`.
   - [ ] `cmd/feedback_add_test.go`: `TestFeedbackAddCmd/{writes first note, second note gets next sequence, lowercase track does not overwrite an existing note, accepts archived track, accepts completed track, note survives track rewrite, note survives track completion, errors on unknown track}`.
   - [ ] `cmd/task/complete_test.go`: `TestTaskCompleteCmd/{files track and bars under completed, lowercase track still hits the guard, lowercase track files uppercase filenames, hand edited lowercase id still hits the guard, replaces archive}`.
   - [ ] `cmd/task/create_test.go`: `TestTaskCreateCmd/{writes first task, echoes minted id, echoes second track id, echoes child id, assigns next id when tasks exist, parent mints dotted id, errors on missing parent, second child increments, appends to reordered track, after inserts mid plan, after rejects unknown anchor, lowercase parent does not destroy an existing bar, uppercases a corrupt prefix on read, errors without title, errors without config}` (the three `TestTaskCreate_*` join `TestTaskCreateCmd`; `uppercases a corrupt prefix on read` is then deleted in this bar, see Scope).
   - [ ] `cmd/task/delete_test.go`: `TestTaskDeleteCmd/{removes file with yes flag, relocates instead of destroying, force deletes unfinished, refuses unfinished without force, prompts for confirmation, keeps task when confirmation unreadable, errors on unknown id, contains path traversal id}`.
   - [ ] `cmd/task/list_test.go`: `TestTaskListCmd/{shows newest first, orders numerically not lexically, groups bars under their track, filters to parent bars, lowercase parent still lists the bars, hand edited lowercase order still ranks bars, parent with no bars, shows phase on bars, empty when no tasks, shows approved marker, unapproved shows empty field}`.
   - [ ] `cmd/task/move_test.go`: `TestTaskMoveCmd/{reorders parent list, rejects bad flags}`.
   - [ ] `cmd/task/read_test.go`: `TestTaskReadCmd/{shows full task, body only, body only empty, shows phase, omits phase when absent, errors on unknown id, contains path traversal id}`.
   - [ ] `cmd/task/update_test.go`: `TestTaskUpdateCmd` keeps its existing rows; the flat tests join it as subtests `changes phase`, `rewrites a corrupt id to canonical case`, `errors on unknown id`, `revokes approval on title change`, `no op preserves approval`, `revokes approval on body change`, `forward status move preserves approval`, `status to todo revokes approval`.

1. **Write test (RED):**
   - [ ] `TestTaskCmd/works from a subfolder against the central store` (subtest, `cmd/task_test.go`)
     - **Behavior:** a registered project's tasks are written to `~/.local/share/bit/<CODE>/` and can be read from any folder under the project.
     - **Setup:** `dir := initProject(t, "BIT")`; `createTask(t, "Track", "...")` from `dir`; `os.MkdirAll(dir+"/src/pkg")`; `t.Chdir(dir+"/src/pkg")`; `bp task list`.
     - **Assertions:** output contains `BIT-1`; `$HOME/.local/share/bit/BIT/tasks/BIT-1.md` exists; `dir/.bit/tasks` does not exist.
     - **Boundary:** depth 2 below the registered path, plus where the record lands (central, not in the repo).
   - [ ] Confirm fails: `bp task list` in the subfolder finds no tasks (relative `.bit`), or the file sits under `dir/.bit/tasks`.

2. **Implement (GREEN):**
   - [ ] `project.OpenStore`, the `cmd/task` `openStore`, all seven `cmd/task` commands, and removing the root pre-run.

3. **More tests (RED → GREEN):**
   - [ ] The existing `TestApproveCmd` and `TestFeedbackAddCmd` cases pass against the central store once `approve`, `unapprove`, `feedback add` and `tui` switch. They create through `bp task create`, which now writes centrally, and that contradiction is what forces the switch.
   - [ ] `TestTaskCmd/unregistered folder says run bp add` and `TestTaskCmd/unregistered folder with bit says run bp migrate` (two table rows in `TestTaskCmd`; a row says whether to `os.Mkdir(".bit")` and which error it wants): sandboxed `HOME`, `XDG_DATA_HOME=""`, `t.Chdir(t.TempDir())`; `bp task list` → `errors.Is(err, project.ErrNotRegistered)`; with `.bit` → `errors.Is(err, project.ErrNeedsMigrate)`. *Boundary:* no registration and no `.bit/`; `.bit/` present, unregistered.
   - [ ] The three moved `TestTaskCmd` worktree subtests pass.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`
- [ ] `grep -rn "bitdir\." cmd/task cmd/approve.go cmd/feedback_add.go cmd/tui.go` finds nothing

## User verifies
- none, deterministic (the verse's end-to-end check is on its last bar)

## Commit (user)
`feat(bit): CLI resolves the project from the registry and uses the central store`