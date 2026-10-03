---
id: BIT-47.9
title: retro_write records the session's HEAD, and a re-run after a new commit appends
status: todo
phase: 3
phase_label: records show the commits they were written at
---
## **Verse 3**

BIT-49.5's `WriteRetro(name, body, head)` already appends through `appendCommit` on a re-run, but its handler passes `task.Commit{}` ("Seam for BIT-47"). A stubbed repo whose SHA must appear on the proposal's `.json`, followed by a re-run at a new SHA that must append, forces the handler to pass `sessionHead` on every write. This bar completes the verse.

## Scope
- `cmd/serve_mcp.go`: `retroWriteHandler(root string, run git.Runner)` gets `dir` from `sessionDir(root)` and calls `store.WriteRetro(name, body, sessionHead(ctx, run, dir))`. `runMCPServer` passes `run`. After this bar, no MCP write handler passes `task.Commit{}`.
- Test file: `cmd/serve_mcp_retro_test.go` (`TestRetroWriteHandler`, new in BIT-49.5 and already in shape). Reuses `mcpSessionWithGit`, `fakeGit` and `readRecord` (BIT-47.6).

## TDD cycle

0. **Restructure:** none. `cmd/serve_mcp_retro_test.go` was created in shape by BIT-49.5.

1. **Write test (RED):**
   - [ ] `TestRetroWriteHandler/records the session head`
     - **Behavior:** a retro proposal shows the commit its project was at when retro ran.
     - **Setup:** sandbox and registered `dir` as `BIT`. `fakeGit` with `rev-parse HEAD` → `9f3c2b7e4d1a0c8b6e5f4a3b2c1d0e9f8a7b6c5d` and `symbolic-ref --short -q HEAD` → `main`. `mcpSessionWithGit(t, dir, fake)`; `retro_write {name: "album-proposals", body: "## Proposal 1\n\n**Pattern:** ...\n"}`.
     - **Assertions:** `$XDG_DATA_HOME/bit/retro/BIT-album-proposals.json` has `commits` with one entry: that `sha`, `branch == "main"` and a parseable `at`.
     - **Boundary:** a new record.
   - [ ] Confirm fails: `commits` is `[]`.

2. **Implement (GREEN):**
   - [ ] Pass `run` into `retroWriteHandler` and call `sessionHead`.

3. **More tests (RED → GREEN):**
   - [ ] `TestRetroWriteHandler/a re-run after a new commit appends`: the same session. Change the fake's `rev-parse HEAD` to `b1e2…` (40 chars) and write the same name again → two entries in order, `created_at` unchanged. *Boundary:* the rewrite path, with HEAD moved.
   - [ ] `TestRetroWriteHandler/no git records no commit`: `mcpSession(t, dir)` → `commits == []`, and the write succeeds.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `grep -rn 'task.Commit{}' cmd/` shows no MCP write handler and no `cmd/feedback_add.go` (only `cmd/migrate*` paths, if any)

## User verifies
Reuse BIT-47.3's sandbox (`$SB/proj` is a git repo registered as `DEMO`) and `$SB/dev.json`, with a fresh `go build -o $SB/bin/bp .` from this commit. Run each `claude` from the shell without the `HOME` export, as `claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config --allowedTools 'mcp__bit__*' -p "<prompt>"`.
- [ ] From `$SB/proj`: prompt "call mcp__bit__research_write with track DEMO-1, topic index, body first". `$SB/data/bit/DEMO/research/DEMO-1/index.json` has one `commits` entry whose `sha` equals `git -C $SB/proj rev-parse HEAD`, with `"branch": "main"`.
- [ ] `git -C $SB/proj -c user.name=t -c user.email=t@t commit --allow-empty -m next`, then the same prompt with body `second`. `index.json` now has two entries, the second at the new HEAD.
- [ ] `git -C $SB/proj worktree add .claude/worktrees/wt -b wt-branch`. From `$SB/proj/.claude/worktrees/wt`, prompt "call mcp__bit__feedback_add with track DEMO-1 and body from the worktree". The new `$SB/data/bit/feedback/DEMO-1-NNN.json` has `"branch": "wt-branch"`.
- [ ] In the sandboxed shell, run `cd $SB/proj && bp feedback add DEMO-1 -d "from the cli"`. Its `.json` has the `$SB/proj` HEAD and `"branch": "main"`.
- [ ] From `$SB/proj`: prompt "call mcp__bit__retro_write with name album-proposals and body test". `$SB/data/bit/retro/DEMO-album-proposals.json` has one `commits` entry at HEAD.
- [ ] Whole slice: research, feedback and retro records each start with the HEAD they were written at, from whichever folder the session runs in, and a later write after a new commit appends one.

## Commit
`feat(bit): retro_write records the session's HEAD`