---
id: BIT-37.1
title: Worktree root resolves to the main checkout's store
status: done
approved: true
phase: 1
phase_label: Read surface
---
## **Verse 1**

The MCP handlers resolve `.bit/` by joining `CLAUDE_PROJECT_DIR` directly, so a worktree
session's tools read a store `bp` never writes. This bar puts the handlers on the same path cut
the CLI already uses — before `task_list` is built on top of it, so the new handler is never
written against the wrong resolver.

## Scope
- `bitdir/bitdir.go` — add a root-taking resolver: cut a `.claude/worktrees/<name>/` root back
  to the main checkout's `.bit`, otherwise `filepath.Join(root, ".bit")`. `Canonical` returns a
  *relative* `.bit` for a non-worktree path, which is why a bare `Canonical(root)` is wrong
  here; factor its segment scan so both entry points share it rather than string-comparing
  `Canonical`'s result.
- `cmd/serve_mcp.go` — `taskReadHandler` builds its store from the new resolver instead of
  `filepath.Join(root, ".bit")`.
- `cmd/serve_mcp_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_ResolvesWorktreeRootToMainCheckout`
     - **Behavior:** a tool call made from a worktree path reads the main checkout's store, so
       the server and `bp` can never disagree about which `.bit/` is live.
     - **Setup:** `dir := t.TempDir()`; save `&task.Task{ID: "FOO-1", Title: "mcp test track",
       Status: task.StatusTodo, Body: "the body"}` into `task.New(filepath.Join(dir, ".bit"))`.
       Start the server via `runMCPServer` with root
       `filepath.Join(dir, ".claude", "worktrees", "wt")` — the directory need not exist, only
       the path shape matters. Call `task_read` with `{"id": "FOO-1"}`.
     - **Assertions:** `result.IsError` is false; `structuredContent.title` is
       `mcp test track`.
     - **Boundary:** the root carries a `.claude/worktrees/<name>` segment pair — the one shape
       that must be cut. The no-worktree root is the other end of that condition and is already
       pinned by the two existing `task_read` tests, so both states are covered.
   - [ ] Confirm fails: the call comes back `IsError` with `loading task FOO-1: … no such file
     or directory`, because the handler looks in
     `<dir>/.claude/worktrees/wt/.bit/tasks/FOO-1.md`.

2. **Implement (GREEN):**
   - [ ] `bitdir`: add the root-taking resolver — the worktree-cut `.bit` when the root has a
     `.claude/worktrees/<name>` segment pair, else `filepath.Join(root, defaultDir)`.
   - [ ] `cmd/serve_mcp.go`: `taskReadHandler` builds its store from it.

## Claude verifies
- [ ] `just test` passes — including the two existing `task_read` tests, which pin the
  non-worktree root
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(mcp): resolve the store through bitdir so a worktree session shares one .bit`