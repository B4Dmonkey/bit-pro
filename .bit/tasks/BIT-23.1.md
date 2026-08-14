---
id: BIT-23.1
title: BIT_DIR env var routes all commands to canonical .bit/
status: todo
phase: 1
phase_label: Worktree routing
---
## **Verse 1**

`bp` resolves `.bit` relative to cwd. A dispatched session's cwd is always a worktree branch, so any `bp` call inside it writes to the branch's checked-out `.bit/`, tangling task state into code commits. A `BIT_DIR` env var (or `--bit-dir` flag) lets any invocation be pointed at the canonical `.bit/` instead.

## Scope
- `cmd/root.go` — change `const bitDir = ".bit"` to `var bitDir = ".bit"`; add `--bit-dir` persistent string flag bound to a local `bitDirFlag`; add `PersistentPreRunE` that resets `bitDir = ".bit"` then applies flag (if non-empty) else `BIT_DIR` env var (flag > env > default)

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestBitDirEnvVar_RoutesListToCanonicalDir`
     - **Behavior:** `bp task list` with `BIT_DIR` set reads from that directory, not from the current directory's `.bit/`
     - **Setup:** `dir1 := initProject(t, "BIT")`; `createTask(t, "Track", "...")` (mints BIT-1 in dir1's `.bit/`); `t.Chdir(t.TempDir())` (simulate a worktree cwd with no local `.bit/`); `t.Setenv("BIT_DIR", filepath.Join(dir1, ".bit"))`
     - **Assertions:** `mustRun(t, "task", "list")` output contains `BIT-1`
     - **Boundary:** env var points to a non-cwd path; cwd has no `.bit/` so the default would fail
   - [ ] Confirm fails: `task.New(bitDir)` resolves to `".bit"` in the new cwd; no tasks exist there; list returns empty
   - [ ] `TestBitDirEnvVar_DefaultIsRelativeDotBit`
     - **Behavior:** with no `BIT_DIR` or flag, commands resolve `.bit` relative to cwd as before
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Track", "...")`
     - **Assertions:** `mustRun(t, "task", "list")` output contains `BIT-1`
     - **Boundary:** no env var, no flag — lower bound; proves the default path is unchanged

2. **Implement (GREEN):**
   - [ ] Change `const bitDir = ".bit"` to `var bitDir = ".bit"` in `cmd/root.go`
   - [ ] Add `var bitDirFlag string` local to `newRootCmd`; bind with `rootCmd.PersistentFlags().StringVar(&bitDirFlag, "bit-dir", "", "canonical .bit directory (overrides BIT_DIR)")`
   - [ ] Add `PersistentPreRunE: func(cmd *cobra.Command, args []string) error { bitDir = ".bit"; if bitDirFlag != "" { bitDir = bitDirFlag } else if v := os.Getenv("BIT_DIR"); v != "" { bitDir = v }; return nil }`

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] From inside `.claude/worktrees/any-slug/`, run `BIT_DIR=$(git rev-parse --show-toplevel)/.bit bp task list` — the list shows the main checkout's tasks, not an empty result

## Commit (user)
`feat(cmd): resolve .bit dir from --bit-dir flag or BIT_DIR env var`