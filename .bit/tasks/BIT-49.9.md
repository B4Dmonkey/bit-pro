---
id: BIT-49.9
title: The git helper reads HEAD and the branch, and gives empty values when git fails
status: todo
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

migrate stamps records with HEAD and the branch, and BIT-47 and BIT-50 extend the same helper. This bar adds the seam: a stub-friendly runner plus `ReadHead`. A stubbed repo and a stubbed failure can't both pass with a hardcoded value, so the test forces real parsing and the empty-on-error rule.

## Scope
- new package `git` (`git/git.go`). **Seam for BIT-47/BIT-50:**
  - `type Runner func(ctx context.Context, dir string, args ...string) (string, error)`. It runs `git <args>` in `dir` and returns stdout with surrounding whitespace trimmed. A non-zero exit returns an error that includes the args and stderr. Copy the shape of `ExecDirRunner` from `git show 6a1d345:claude/dispatch.go`, but bind the name to `git`, keep stdout and stderr separate (stderr goes only in the error), and fold the exit code into `err`.
  - `func ExecRunner(ctx context.Context, dir string, args ...string) (string, error)` uses `exec.CommandContext(ctx, "git", args...)` with `cmd.Dir = dir`.
  - `type Head struct { SHA, Branch string }`.
  - `func ReadHead(ctx context.Context, run Runner, dir string) Head`:
    - `SHA` comes from `run(ctx, dir, "rev-parse", "HEAD")`, the full 40-character SHA.
    - `Branch` comes from `run(ctx, dir, "symbolic-ref", "--short", "-q", "HEAD")`.
    - Each field falls back to `""` on its own error, so no error is returned. A detached HEAD gives a SHA and an empty branch. A repo with no commits gives a branch and an empty SHA. No git gives both empty.
  - Later tracks add functions next to `ReadHead` in this package. They take the same `Runner` as their second parameter, and tests pass a fake `Runner`.
- new package `gittest` (`git/gittest/gittest.go`). This is test support in the style of `httptest`, not a `_test.go` file, so the `git`, `cmd`, `migrate` and `landing` tests can all import it. **Every test that runs the real `git` goes through it.** The reason: this repo's pre-commit hook runs `just test` inside `git commit`, and git hands the hook `GIT_INDEX_FILE`, `GIT_PREFIX`, `GIT_CONFIG_PARAMETERS`, `GIT_AUTHOR_*`, `GIT_EDITOR` and `GIT_EXEC_PATH`, and can also export `GIT_DIR` and `GIT_WORK_TREE`. A test that inherits them runs its temp repo's git against bit-pro's own index.
  - `func Env() []string`: `os.Environ()` without any entry whose key starts with `GIT_`. `PATH`, `HOME` and everything else stay.
  - `func Run(t testing.TB, dir string, args ...string) string`: `exec.LookPath("git")`, else `t.Skip`. Runs `git <args>` in `dir` with `cmd.Env = Env()`, returns trimmed stdout, and calls `t.Fatalf` with the args and stderr on failure. Tests use it for their own setup and inspection commands.
  - `func Isolate(t testing.TB)`: for each `GIT_*` key in `os.Environ()`, `t.Setenv(k, "")` then `os.Unsetenv(k)`. `t.Setenv` restores them at cleanup. A test calls it when the code under test runs `git.ExecRunner`, which inherits the process environment.
  - BIT-50.1 extends this package with `New`, `Repo`, `Repo.Git` and `Repo.Commit`.
- **How tests stub it:** a `fakeGit` in the test, `map[string]struct{out string; err error}` keyed by `strings.Join(args, " ")`, returned as a `git.Runner` closure that also records the `dir`. BIT-49.15 and later bars reuse that pattern in `migrate`'s tests.
- Test files touched: new `git/git_test.go` and new `git/gittest/gittest_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestReadHead` (table, `git/git_test.go`), row `on a branch`
     - **Behavior:** the helper reports the commit and branch of the folder it's pointed at.
     - **Setup:** a fake runner: `rev-parse HEAD` → `"6a1d3459c0ffee000000000000000000000000ab\n"`, `symbolic-ref --short -q HEAD` → `"v2\n"`. `dir = "/repo/.claude/worktrees/wt"`.
     - **Assertions:** `Head{SHA: "6a1d3459c0ffee000000000000000000000000ab", Branch: "v2"}`, and the runner saw that `dir` both times.
     - **Boundary:** a normal checkout, with trailing newlines trimmed.
   - [ ] Confirm fails: package `git` doesn't exist.

2. **Implement (GREEN):**
   - [ ] `Runner`, `Head`, `ReadHead`.

3. **More tests (RED → GREEN):** (rows of `TestReadHead`)
   - [ ] `detached head`: `symbolic-ref` errors → `Head{SHA: "<sha>", Branch: ""}`. *Boundary:* each field fails on its own.
   - [ ] `no commits yet`: `rev-parse` errors, `symbolic-ref` gives `main` → `Head{"", "main"}`.
   - [ ] `not a git folder`: both error → `Head{}`. *Boundary:* empty values and no failure.
   - [ ] `TestRun/ignores an inherited index file` (`git/gittest/gittest_test.go`): `bogus := filepath.Join(t.TempDir(), "missing", "index")` and `t.Setenv("GIT_INDEX_FILE", bogus)`. In `dir := t.TempDir()`: `Run(t, dir, "init", "-b", "main")`, write `a.txt`, `Run(t, dir, "add", "a.txt")`, then `Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-m", "x")`. → `Run(t, dir, "ls-files") == "a.txt"`, and `os.Stat(bogus)` gives `fs.ErrNotExist`. *Boundary:* the hook's environment. Without the stripping, `git add` fails because `bogus`'s folder doesn't exist.
   - [ ] `TestIsolate/unsets inherited git variables`: `t.Setenv("GIT_INDEX_FILE", bogus)` and `t.Setenv("GIT_DIR", bogus)`, then `Isolate(t)` → `os.LookupEnv` reports both unset. *Boundary:* the code under test sees no `GIT_*`.
   - [ ] `TestExecRunner/reads a real repo`: `gittest.Isolate(t)`. In `dir := t.TempDir()`: `gittest.Run(t, dir, "init", "-b", "main")`, then `gittest.Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "--allow-empty", "-m", "x")`. `ReadHead(ctx, ExecRunner, dir)` → a 40-character SHA equal to `gittest.Run(t, dir, "rev-parse", "HEAD")`, and `Branch == "main"`. *Boundary:* the real binary (`gittest.Run` skips without it).
   - [ ] `TestExecRunner/a failing command returns an error`: `gittest.Isolate(t)`, then `ExecRunner(ctx, t.TempDir(), "rev-parse", "HEAD")` outside a repo → non-nil error that contains `rev-parse`. *Boundary:* an inherited `GIT_DIR` would make this pass against bit-pro's repo.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): git helper reads HEAD and branch with empty values on failure`