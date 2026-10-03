---
id: BIT-49.23
title: From a subfolder or a Claude worktree, migrate reads and registers the main checkout's .bit/
status: todo
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

Today migrate reads only `<cwd>/.bit`. In a Claude worktree, that's the worktree's own tracked snapshot, not the live state v1 keeps in the main checkout. From a subfolder, there's no `.bit` at all. A worktree test whose snapshot differs from the main checkout's forces `project.MainCheckout`, and a subfolder test forces the walk-up. Git facts still come from the session dir.

## Scope
- `migrate/locate.go` (new): unexported `locate(dir string) (string, error)`, which returns the folder that holds the `.bit/` to read:
  1. `if root, ok := project.MainCheckout(dir); ok`: return `root` if `root/.bit` is a dir, else an error (`ErrNoBitDir`).
  2. Otherwise walk from `dir` up to `/` and return the first folder that has a `.bit` dir.
  3. If none is found: `var ErrNoBitDir = errors.New("no .bit/ here or above")`.
- `migrate.Run`: `holder, err := locate(opts.Dir)`, then `src := filepath.Join(holder, ".bit")` and `path, err := project.CanonicalPath(holder)` (it returns `(string, error)`, BIT-46.4). `git.ReadHead` keeps `opts.Dir`, the session dir. BIT-49.25's tracked check uses `holder`.
- Test files touched: `migrate/migrate_test.go` (stub git) and `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestRun/in a claude worktree reads the main checkout`
     - **Behavior:** migrating from a worktree session copies the live v1 state and registers the main checkout, while recording the worktree's own HEAD and branch.
     - **Setup:** sandbox. `dir/.bit` is a valid store with `tasks/BIT-1.md` and `tasks/BIT-2.md`. `wt := dir/.claude/worktrees/wt` holds its own `.bit` with only `tasks/BIT-1.md` (a stale snapshot). `fakeGit` returns branch `worktree-wt`. Call `Run(Options{Dir: wt, ...})`.
     - **Assertions:** the registered path is `CanonicalPath(dir)`, not `wt`. The store has `BIT-1` and `BIT-2`. `tasks/BIT-2.json` has `"branch": "worktree-wt"`. The fake saw `wt` as the git dir.
     - **Boundary:** a session dir inside `.claude/worktrees/<name>`.
   - [ ] Confirm fails: the worktree's snapshot is migrated and `wt` is registered.

2. **Implement (GREEN):**
   - [ ] `locate`, and its use in `Run`.

3. **More tests (RED → GREEN):**
   - [ ] `TestMigrateCmd/walks up from a subfolder`: run `bp migrate` from `dir/src/pkg` → registers `dir`, and the output path is `dir`. *Boundary:* depth 2 below the holder.
   - [ ] `TestMigrateCmd/no bit folder here or above`: from a fresh `t.TempDir()` → `errors.Is(err, migrate.ErrNoBitDir)`. *Boundary:* nothing to migrate.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate finds the main checkout's .bit/ from subfolders and worktrees`