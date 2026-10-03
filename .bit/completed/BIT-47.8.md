---
id: BIT-47.8
title: bp feedback add records the current folder's HEAD on a new note
status: done
approved: true
phase: 3
phase_label: records show the commits they were written at
---
## **Verse 3**

The CLI is the other way a note gets in, and the scope wants its `commits` to match MCP's (Decision "The CLI's `bp feedback add` captures the same way, from the current folder"). Today it passes `task.Commit{}`. A note added inside a real git repo, whose `.json` must carry that repo's `git rev-parse HEAD`, forces the capture.

## Scope
- `cmd/feedback_add.go`: `wd, err := os.Getwd()`. Then `AddNote(args[0], description, sessionHead(cmd.Context(), git.ExecRunner, wd))`, reusing BIT-47.6's helper. A `Getwd` error is returned, as the store resolution already does (BIT-46.8's `os.Getwd` path). No flags (scope: no `--commit`/`--branch`).
- Test file: `cmd/feedback_add_test.go` (`TestFeedbackAddCmd`, converted by BIT-46.8; BIT-49.1 added `writes to the shared feedback folder`). Tests use the real `git` binary through BIT-49.9's `gittest` helper, as its `TestExecRunner` does: both subtests call `gittest.Isolate(t)` first, so `git.ExecRunner` in the command doesn't inherit the pre-commit hook's `GIT_*` variables, and run their own git commands with `gittest.Run` (which skips when `git` is missing).

## TDD cycle

0. **Restructure:** none. Already converted by BIT-46.8.

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd/records the current folder's head`
     - **Behavior:** a note added from the command line shows the commit and branch of the folder it was added from, the same as through MCP.
     - **Setup:** `gittest.Isolate(t)`. `dir := initProject(t, "BIT")` (cwd is `dir`); `createTask(t, "Track", "...")`. `gittest.Run(t, dir, "init", "-b", "main")`, then `gittest.Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "--allow-empty", "-m", "init")`. `want := gittest.Run(t, dir, "rev-parse", "HEAD")`. Then `bp feedback add BIT-1 -d firstNote`.
     - **Assertions:** the `.json` beside the printed path has `commits` with one entry: `sha == want` (40 characters), `branch == "main"`, and an `at` that parses as RFC 3339.
     - **Boundary:** a real repo with one commit.
   - [ ] Confirm fails: `commits` is `[]`.

2. **Implement (GREEN):**
   - [ ] `os.Getwd` plus `sessionHead` in `RunE`.

3. **More tests (RED → GREEN):**
   - [ ] `TestFeedbackAddCmd/a folder outside git records no commit`: `gittest.Isolate(t)`, then `initProject` without `git init` (a `t.TempDir()`) → the note is written and `commits == []`. *Boundary:* the real `git` fails, and the command still succeeds.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): bp feedback add records the current folder's HEAD`