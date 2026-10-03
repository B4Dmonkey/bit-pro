---
id: BIT-46.7
title: bp init is gone and the test fixtures register their project
status: done
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

`bp init` depends on `bitdir` and `config.toml` (`cmd/init.go:10,39,75`), which this verse removes, and the shared fixture `initProject` runs it (`cmd/cmd_test.go:93`). The track puts both changes in one bar. This is a removal bar, so it adds no new tests (operator rule for v2): it deletes the command, removes the tests that only covered it, and rewrites the fixtures so the existing suite stays green. The fixtures start registering their project in a sandboxed registry, which the next bar's CLI switch needs.

## Scope
- `cmd/init.go` — delete `newInitCmd`, `initCmdUse`, `readInteractivePrefix` and the imports only they used. **Keep `writeClaudeWiring(cmd, run, dir)` in this file** (BIT-48's Touches point at it here; `bp add` is its only caller). After this bar `cmd/init.go` holds only `writeClaudeWiring` and the imports it uses (`fmt`, `path/filepath`, `claude`, `cobra`): no other function, const or var. BIT-48.3 relies on that and deletes the file whole.
- `cmd/root.go` — drop `rootCmd.AddCommand(newInitCmd(run))`.
- delete `cmd/init_test.go`; delete `mcpRegisterCall` (`cmd/cmd_test.go:98`, used only by init_test) and `prefixFlag` (`cmd/testconst_test.go:4`).
- `cmd/cmd_test.go` `initProject(t, prefix)` becomes: `t.Setenv("HOME", t.TempDir())`, `t.Setenv("XDG_DATA_HOME", "")`, `dir := t.TempDir()`, `t.Chdir(dir)`, `path, err := project.CanonicalPath(dir)` (it returns `(string, error)`, BIT-46.4), then register `orm.CreateProjectParams{Path: path, Code: prefix}` with `seedProject` from `cmd/list_test.go` (kept as is: `seedProject(t, orm.CreateProjectParams)`, which BIT-46.9, BIT-46.20, BIT-49.3 and BIT-50.1 also call), then `task.New(".bit").SaveConfig(&task.Config{Prefix: prefix})`. The `SaveConfig` keeps the still-`bitdir`-based CLI working for this bar; the next bar drops it. No cmd test uses `t.Parallel`, so `t.Setenv` is safe.
- `cmd/task/helpers_test.go` `initProject`: the same sandbox and registration (its own copy, package `task_test`), keeping its `SaveConfig`.
- `scripts/install.sh:22` — `echo "Run 'bp add <path>' in a project to register it."`.

## Steps
- [ ] Delete the command, its registration, `init_test.go`, `mcpRegisterCall` and `prefixFlag`.
- [ ] Rewrite both fixtures as above.
- [ ] Edit `install.sh`.

## Claude verifies
- [ ] `just lint` and `just test` pass. Lint catches any helper that only `init_test.go` used.
- [ ] `test ! -e ~/.local/share/bit/main.db`. The fixtures now open the db at all 78 call sites, so this proves they're sandboxed.
- [ ] `grep -rn "newInitCmd\|initCmdUse\|prefixFlag\|readInteractivePrefix" --include='*.go' .` finds nothing, and `grep -n '^func\|^const\|^var' cmd/init.go` shows only `func writeClaudeWiring`; `grep -n "bp init" scripts/install.sh` finds nothing.

## User verifies
- none (deterministic)

## Commit (user)
`feat(bit): delete bp init; fixtures register a sandboxed project`