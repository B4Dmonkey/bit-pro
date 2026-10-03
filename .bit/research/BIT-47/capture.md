# Write-time git capture on research / feedback (Q4)

Seams only. BIT-49 owns the git helper (shells out to `git`, tests stub it; `.bit/tasks/BIT-49.md:33-34`), BIT-46 adds the empty `commits` field (`.bit/tasks/BIT-46.md:59-63`).

## Today
- No Go git helper exists. `exec.Command` only appears in `claude/sync.go`, `claude/dispatch.go`, `claude/plugin.go`, `daemon/daemon.go` (none call git). The `claude.Runner`/`DirRunner` func types (`claude/sync.go:12`, `claude/dispatch.go:16` `ExecDirRunner(ctx, dir, name, args...)`) are the existing stubbable-exec pattern; a `DirRunner` already takes a dir, which is what "session dir" needs.
- `task/feedback.go:58-79` `AddNote(track, body)`: resolve track, mkdir, next seq, `os.WriteFile` of the raw body. Always creates a new file, so capture = one `commits` entry at create.
- `task/research.go:37-57` `WriteResearch(track, topic, body)`: overwrite in place. Capture must read the existing record's `commits`, append `{sha, branch, at}` only if HEAD moved (stub rule), then write. Today the body has no frontmatter and nothing is read back before writing — so there is nowhere to keep `commits` in v1 files. That lands with BIT-46's JSON metadata sidecar.
- Retro write: doesn't exist (BIT-49).

## Where the call goes
- Session dir is known only at the entry points: MCP handlers via `root` (`CLAUDE_PROJECT_DIR`, `cmd/serve_mcp.go:231`), CLI via `os.Getwd()`. `Store` only holds the canonical bit dir (`bitdir.ForRoot` strips the worktree, `bitdir/bitdir.go:49-67`), which in a worktree points at the main checkout — wrong HEAD/branch. So either:
  - handler reads HEAD/branch via the helper and passes a git fact into the Store write (`AddNote(track, body, gitFact)` / an options struct), or
  - Store is constructed with the session dir + a git reader injected.
  Either way the Store must not shell out from `s.root`.
- Bars/tracks are not captured on write (stub); only via explicit `commit`/`branch` inputs ([mcp-git-inputs](mcp-git-inputs.md)).
- Empty-git case (`acme/`): helper must return empty, not error, so writes still succeed.
- v2 store location: BIT-46 moves records to `~/.local/share/bit/<CODE>/`, so `bitdir`/`ForRoot` may not survive; the session-dir-vs-store-dir split is the seam that does.
