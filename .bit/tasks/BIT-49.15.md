---
id: BIT-49.15
title: Migrated records carry HEAD and the branch of the session dir
status: todo
phase: 2
phase_label: bp migrate
---
## **Verse 2**

The scope fixes what migrate writes for git info, per record kind. With a stubbed repo, records must carry that SHA and branch. In a folder with no git, they must be empty. Those two cases force the git helper into `migrate.Run`.

## Scope
- `migrate/migrate.go`:
  - `Options` gains `Git git.Runner` and `Now func() time.Time`.
  - At the start of `Run`: `h := git.ReadHead(ctx, opts.Git, opts.Dir)`, the session dir (scope: git facts come from the session dir). Then `head := task.Commit{SHA: h.SHA, Branch: h.Branch, At: opts.Now()}`. `appendCommit` already normalizes `At` to UTC seconds (BIT-46.16).
  - Tasks, in every place: set `t.Branch = h.Branch` and `t.Commit = h.SHA` before `SaveTo` (BIT-46.14 writes them as given).
  - Feedback, research and retro: pass `head` instead of `task.Commit{}`. `appendCommit` turns an empty SHA into `"commits": []` (BIT-46.16), so a folder with no git leaves every git field empty.
  - Record timestamps come from the store's clock at write time, which is the migration time.
- `cmd/migrate.go`: passes `Git: git.ExecRunner, Now: time.Now`.
- Test files touched: new `migrate/migrate_test.go` (a `TestRun` function, sandboxed `HOME`/`XDG_DATA_HOME`, `db.Open`, and the `fakeGit` pattern from BIT-49.9) and `cmd/migrate_test.go`. From this bar on, `bp migrate` runs the real `git` through `git.ExecRunner`, so every `TestMigrateCmd` case calls `gittest.Isolate(t)` (BIT-49.9) beside `mcpSandbox(t)`. The pre-commit hook's `GIT_*` variables then never point the command's git at bit-pro's own repo or index.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestRun/stamps every record with the session head`
     - **Behavior:** every migrated record points at the commit and branch the project was migrated from.
     - **Setup:** a v1 store with `tasks/BIT-1.md`, `completed/BIT-1.1.md`, `feedback/BIT-1-001.md`, `research/BIT-1/index.md` and `retro/album-proposals.md`. `fakeGit`: `rev-parse HEAD` → `9f3c…` (40 characters) and `symbolic-ref --short -q HEAD` → `v2`. `Now` is fixed at `2026-10-02T10:00:00-04:00`. Call `migrate.Run(ctx, q, Options{Dir: dir, Git: fake, Now: fixed})`.
     - **Assertions:**
       - `tasks/BIT-1.json` and `completed/BIT-1.1.json` have `"branch": "v2"` and `"commit": "9f3c…"`.
       - The `feedback/BIT-1-001.json`, `research/BIT-1/index.json` and `retro/BIT-album-proposals.json` records each have `commits == [{"sha": "9f3c…", "branch": "v2", "at": "2026-10-02T14:00:00Z"}]`.
       - The fake saw `dir` as the git dir.
     - **Boundary:** every record kind gets a git field.
   - [ ] Confirm fails: `Options` has no `Git` field (compile), then the fields are empty.

2. **Implement (GREEN):**
   - [ ] `ReadHead` in `Run`, the task fields and the `head` argument. Wire `cmd/migrate.go`.

3. **More tests (RED → GREEN):**
   - [ ] `TestRun/a folder with no git leaves git fields empty`: `fakeGit` errors on every call → tasks have `"branch": ""` and `"commit": ""`, and the other records have `"commits": []`. *Boundary:* no git.
   - [ ] `TestMigrateCmd/migrates a folder outside git`: the existing cmd fixture (a `t.TempDir()`, not a repo, now with `gittest.Isolate(t)`) still migrates with empty git fields through the real `git.ExecRunner`. *Boundary:* a real git error never fails the command.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrated records carry the session's HEAD and branch`