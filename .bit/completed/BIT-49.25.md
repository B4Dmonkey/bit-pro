---
id: BIT-49.25
title: migrate prints the cleanup step for a tracked or untracked .bit/ and never runs it
status: done
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

After a migration, the operator removes `.bit/` as a cutover step. The step differs: `git rm -r .bit` where it's tracked, plus deleting the untracked files that leaves behind (bit-pro has its untracked planning records, e.g. BIT-45..50). Otherwise the folder is just removed. Testing both a tracked and an untracked source forces a git check the helper doesn't have yet, which is the first extension of BIT-49.9's seam. This bar completes the verse.

## Scope
- `git/git.go`: `func Tracks(ctx context.Context, run Runner, dir, path string) bool`, which runs `run(ctx, dir, "ls-files", "--", path)`. Non-empty output means `true`. An error or empty output means `false`.
- `migrate/migrate.go`: `Result` gains `Holder string` (the folder holding `.bit/`) and `Tracked bool`, from `git.Tracks(ctx, opts.Git, holder, ".bit")` (asked in the holder, which is the main checkout).
- `cmd/migrate.go`: between the `migrated` line and `ensureGlobalWiring`, print:
  - tracked: `cleanup: in <holder>, run \`git rm -r .bit\` and commit, then \`rm -rf .bit\` to delete the untracked files git leaves behind`;
  - untracked: `cleanup: remove the v1 folder with \`rm -rf <holder>/.bit\``.

  migrate never runs either command.
- Test files touched: `git/git_test.go`, `migrate/migrate_test.go` and `cmd/migrate_test.go`. `TestMigrateCmd` cases already call `gittest.Isolate(t)` (BIT-49.15). That matters most here, because `ls-files` reads the index, and an inherited `GIT_INDEX_FILE` would answer from bit-pro's.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/prints the cleanup step for an untracked bit folder`
     - **Behavior:** the operator is told how to remove the v1 folder, and it's still there.
     - **Setup:** sandbox and `gittest.Isolate(t)`. A valid store in a `t.TempDir()` (not a repo). Then `bp migrate`.
     - **Assertions:** the output contains `rm -rf <canonical dir>/.bit` and doesn't contain `git rm`. `dir/.bit` still exists, byte-unchanged.
     - **Boundary:** an untracked `.bit/`.
   - [ ] Confirm fails: no cleanup line.

2. **Implement (GREEN):**
   - [ ] `Tracks`, the `Result` fields, and both messages.

3. **More tests (RED → GREEN):**
   - [ ] `TestTracks` (table, fake runner): output `.bit/config.toml\n` → true; empty → false; error → false. *Boundary:* each outcome.
   - [ ] `TestRun/reports a tracked bit folder`: fake `ls-files -- .bit` → `.bit/config.toml`, with the dir == the holder → `res.Tracked == true`. *Boundary:* the check runs in the holder, not the session dir (use a worktree session as in BIT-49.23).
   - [ ] `TestMigrateCmd/prints git rm for a tracked bit folder`: sandbox and `gittest.Isolate(t)`. In `dir`: `gittest.Run(t, dir, "init", "-b", "main")`, `gittest.Run(t, dir, "add", ".bit")`, then `gittest.Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-m", "v1")` (`gittest.Run` skips without `git`). Then `bp migrate` → the output contains `git rm -r .bit` and `rm -rf .bit`, and `gittest.Run(t, dir, "status", "--porcelain")` is empty. *Boundary:* a real tracked `.bit/`, with the repo left untouched.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
This checks the whole verse in a sandbox (never `just install`). BIT-46.11's sandbox recipe uses a fake `claude` on `PATH`. A client project is never opened or copied.
- [ ] Set up as in BIT-46.11. Then `cp -R <v2 checkout> $SB/bitpro` and `cp -R <repo>/tools/example $SB/example` (copies, so `.bit/` keeps its tracked state in `$SB/bitpro`).
- [ ] `cd $SB/bitpro && bp migrate` prints `migrated BIT …/bitpro`, then the `git rm -r .bit` cleanup line. `find $SB/data/bit/BIT/tasks $SB/data/bit/BIT/completed $SB/data/bit/BIT/archive/tasks -type f | wc -l` is twice `find .bit/tasks .bit/completed .bit/archive/tasks -name '*.md' | wc -l`. `ls $SB/data/bit/feedback | grep -c '\.md$'` matches `ls .bit/feedback | wc -l`. `git -C $SB/bitpro status --porcelain .bit` prints the same as a capture taken just before `bp migrate` (bit-pro's untracked `.bit/` files show as `??` both times).
- [ ] `cd $SB/bitpro && bp task list` lists the same active tracks as `ls .bit/tasks` (without `.md`), and `bp task create "Probe"` mints the next track ID after the highest across `.bit/tasks`, `.bit/completed` and `.bit/archive/tasks`.
- [ ] `cd $SB/bitpro/.claude/worktrees/x` (`mkdir -p` it) `&& bp migrate` prints `already migrated`.
- [ ] `cd $SB/example && bp migrate` prints `migrated` with example's code and the `rm -rf …/.bit` cleanup line, if example's `.bit/` is untracked there.
- [ ] Whole slice: a live v1 project's state can be trialled in v2 from a sandbox copy, with the source untouched and v1 still working on it.

## Commit (user)
`feat(bit): migrate prints the .bit/ cleanup step`