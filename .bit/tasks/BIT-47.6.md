---
id: BIT-47.6
title: research_write records the session's HEAD, read per call from the session dir
status: todo
approved: true
phase: 3
phase_label: records show the commits they were written at
---
## **Verse 3**

BIT-46.16's `WriteResearch(track, topic, body, head)` keeps a `commits` history, but the handler passes `task.Commit{}`. A stubbed git repo whose SHA must show up in the topic's `.json` forces the MCP server to read HEAD. That needs BIT-49's git helper threaded into `runMCPServer` (scope Decision "Git facts are read where the request comes in").

## Scope
- new `cmd/session_head.go`:
  - unexported `sessionDir(root string) (string, error)`: `root` when it's set, else `os.Getwd()` (the opencode fallback). `mcpStore` (BIT-46.9) switches to it, so the store and the git read agree on the session dir. This is a small extraction with no behaviour change.
  - unexported `sessionHead(ctx context.Context, run git.Runner, dir string) task.Commit`: `h := git.ReadHead(ctx, run, dir)` and `return task.Commit{SHA: h.SHA, Branch: h.Branch, At: time.Now()}`. It never errors: an empty SHA means no git, and `appendCommit` then adds nothing (BIT-46.16). The store normalizes `At`.
- `cmd/serve_mcp.go`:
  - `runMCPServer(ctx context.Context, root string, run git.Runner, transport mcp.Transport)`. `newServeMCPCmd` passes `git.ExecRunner`.
  - `researchWriteHandler(root string, run git.Runner)`: `dir, err := sessionDir(root)`. On error it returns the error, as `mcpStore` does. Then it calls `store.WriteResearch(track, topic, body, sessionHead(ctx, run, dir))`. HEAD is read on every call, never once at startup, because a session commits between writes.
  - Git is read in the session dir (`CLAUDE_PROJECT_DIR`, the worktree in a worktree session), not the resolved project root.
- `cmd/mcp_harness_test.go`:
  - `mcpSessionWithGit(t, root string, run git.Runner) *mcp.ClientSession` holds today's `mcpSession` body, including BIT-46.9's `mcpSandbox(t)` call, and starts `runMCPServer(ctx, root, run, serverT)`. `mcpSession(t, root)` calls it with `noGit`, a runner that returns an error for every call, so every existing session (BIT-46.9's and BIT-49's) stays hermetic and record no commits.
  - `fakeGit`: BIT-49.9's pattern in package `cmd`. It's a `map[string]struct{ out string; err error }` keyed by `strings.Join(args, " ")`, and it returns a `git.Runner` that also appends each `dir` it sees to a slice the test reads. A missing key returns an error.
  - `readRecord(t, mdPath string) map[string]any`: reads the `.json` beside a returned `.md` path.
- Test file: `cmd/serve_mcp_research_test.go` (`TestResearchWriteHandler`, converted by BIT-46.9).
- **Seam for BIT-50:** `runMCPServer`'s `git.Runner` parameter and `mcpSessionWithGit`/`fakeGit`. `task_landing`'s handler takes the same `run`.

## TDD cycle

0. **Restructure:** none. `cmd/serve_mcp_research_test.go` was converted by BIT-46.9, and `cmd/mcp_harness_test.go` holds no tests.

1. **Write test (RED):**
   - [ ] `TestResearchWriteHandler/records the session head`
     - **Behavior:** a research topic remembers the commit and branch it was written at.
     - **Setup:** sandbox and registered `dir` with track `testTrackID`. `fakeGit{"rev-parse HEAD": {out: "9f3c2b7e4d1a0c8b6e5f4a3b2c1d0e9f8a7b6c5d\n"}, "symbolic-ref --short -q HEAD": {out: "v2\n"}}`. `mcpSessionWithGit(t, dir, fake)`; `research_write {track: testTrackID, topic: "index", body: testResearchBody}`.
     - **Assertions:** `readRecord(path)["commits"]` has one entry with `sha == "9f3c2b7e4d1a0c8b6e5f4a3b2c1d0e9f8a7b6c5d"`, `branch == "v2"` and an `at` that parses as RFC 3339, falling between the test's start and end (truncated to seconds). The fake saw `dir`.
     - **Boundary:** a new topic, HEAD on a branch.
   - [ ] Confirm fails: `mcpSessionWithGit` and the `runMCPServer` parameter don't exist (compile). Once added, `commits` is `[]` because the handler still passes `task.Commit{}`.

2. **Implement (GREEN):**
   - [ ] `sessionDir`, `sessionHead`, the `runMCPServer` parameter, the harness helpers and the handler change.

3. **More tests (RED → GREEN):**
   - [ ] `TestResearchWriteHandler/a later write after a new commit appends`: the same session. Write once, then change the fake's `rev-parse HEAD` to `b1e2…` (40 chars), and write the same topic again. `commits` has two entries in order. *Boundary:* HEAD is read per call, so caching it at startup fails this.
   - [ ] `TestResearchWriteHandler/reads git in the session dir, not the project root`: `wt := dir/.claude/worktrees/wt` (`os.MkdirAll`). `mcpSessionWithGit(t, wt, fake)`; one write. Every dir the fake saw equals `wt`, not `dir`, while the topic lands in `dir`'s store. *Boundary:* a worktree session.
   - [ ] `TestResearchWriteHandler/no git records no commit`: `mcpSession(t, dir)` (the `noGit` runner) → `commits == []` and the write succeeds. *Boundary:* git errors never fail a write.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `grep -n 'task.Commit{}' cmd/serve_mcp.go` no longer shows the research handler

## User verifies
- none, deterministic (the verse's end-to-end check is on BIT-47.9)

## Commit
`feat(bit): research_write records the session's HEAD`