---
id: BIT-46.10
title: The plugin-behind notice reads the resolved project's root
status: todo
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

`pluginState` runs after every command and calls `bitdir.Root()` (`cmd/root.go:21-28`), the last production caller of `bitdir` outside the package. The track decides its root is the resolved project's path, else the current folder, never an error. A test from a subfolder of a registered project forces it: `bitdir.Root()` returns the subfolder.

## Scope
- `cmd/root.go`:
  - unexported `pluginRoot() string`: `wd, err := os.Getwd()` (error → `"."`, as today); `p, err := project.Find(context.Background(), wd)`; on success return `p.Path`, otherwise `wd`.
  - `pluginState` calls `claude.PluginState(home, pluginRoot())`; drop the `bitdir` import.
  - **Seam for BIT-48:** BIT-48 later drops the root argument when the notice reads the user-scope install; `pluginRoot` is what it deletes.
- `cmd/root_test.go` (converted in BIT-45.1) — no sandbox changes. Only `execute` calls `pluginState`: `run`/`mustRun` call `root.Execute()` and never reach it (so `TestRootCmd/help` and `TestRootCmd/version` don't), `TestExecute/suppressed command writes no notice` returns before it and overrides it, and the other `runSplit` callers already run `initProject`, which sandboxes `HOME`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestPluginRoot` (table, `cmd/root_test.go`)
     - **Behavior:** the notice checks the plugin install of the project bp is running in, even from deep inside it.
     - **Rows / Assertions:**
       - `TestPluginRoot/registered subfolder`: `dir := initProject(t, "BIT")`; `mkdir dir/src`; `t.Chdir(dir+"/src")` → `pluginRoot()` equals the path `project.CanonicalPath(dir)` returns (it returns `(string, error)`; fail on the error).
       - `TestPluginRoot/unregistered folder`: sandboxed `HOME`/`XDG_DATA_HOME`, `t.Chdir(t.TempDir())` → `pluginRoot() == os.Getwd()`.
     - **Boundary:** registered-subfolder vs unregistered — both branches; the unregistered one must not error.
   - [ ] Confirm fails: the `registered subfolder` row returns `dir/src` (`bitdir.Root` only cuts at `.claude/worktrees`).

2. **Implement (GREEN):**
   - [ ] `pluginRoot` and the `pluginState` change.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): plugin notice uses the resolved project root`