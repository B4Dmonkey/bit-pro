# Project resolution: v1 today, and the v2 "longest registered path" design (Q4)

**Checked:** how every entry point finds its store today, re-checked against the code on `v2` @ 6a1d345, and a design for "the longest registered path that contains cwd".

## v1 today (re-verified)
- **No walk-up.** `bitdir.Resolve()` runs in root `PersistentPreRunE` (`cmd/root.go:136`). It sets `current = Canonical(wd)`, which is the relative `".bit"` unless wd contains `.claude/worktrees`. In that case `worktreeCut` returns `<main checkout>/.bit`.
- **CLI callers** of `bitdir.Current()`: `approve.go:15,26`, `feedback_add.go:19`, `init.go:39,75`, `tui.go:24`, and `cmd/task/{create,complete,list,delete,read,move,update}.go`.
- `bitdir.Root()` has one caller: the plugin-version notice at `cmd/root.go:27`, via `claude.PluginState(home, root)`, which reads `<root>/.claude/settings.json`. It still needs the repo root in v2.
- **MCP**
  - `cmd/serve_mcp.go:226` reads `root := os.Getenv("CLAUDE_PROJECT_DIR")` once, at start.
  - The 10 handlers (:294-489) each call `task.New(bitdir.ForRoot(root))` per call.
  - Tests call `runMCPServer(ctx, root, transport)` with a temp root. That is the seam.
- **opencode (new finding):** `opencode.json` registers `bp serve mcp` with **no `CLAUDE_PROJECT_DIR`**, so root is `""` and resolution uses `.bit` relative to the server's cwd. v2 must fall back to `os.Getwd()` when the env var is empty. Unverified: whether opencode spawns the server with cwd set to the project dir.
- **`.opencode/bit-list.tsx`** runs `bp task list/read` with `cwd`, so it goes through the CLI path.
- **TUI queue** `GetProjectByPath` (exact match) and **daemon** `p.Path/.bit` both disappear with [daemon-removal](daemon-removal.md). That leaves two resolvers to merge: CLI (bitdir) and MCP (env).
- **git is not used anywhere** (no `exec` of git; the only `exec`s are `claude` and `launchctl`).

## Design for v2
One package (e.g. `project/`) replaces `bitdir`: `Resolve(ctx, q, dir string) (Project, error)`.
- **`dir`:**
  - CLI: `os.Getwd()`.
  - MCP: `$CLAUDE_PROJECT_DIR`, else `os.Getwd()`.
  - The TUI uses the CLI path.
  - Resolve per call, not at server start. That lets a project registered mid-session start working.
- **Normalise:** `filepath.Abs`, then `filepath.EvalSymlinks`. macOS matters here: `/tmp` → `/private/tmp`, and `t.TempDir()` paths are symlinked. Store paths the same way at `add`. APFS is case-insensitive, but paths from `Getwd` keep their case; that is fine if both sides come from the OS.
- **Match:** `SELECT … FROM projects WHERE ? = path OR ? LIKE path || '/%' ORDER BY length(path) DESC LIMIT 1`, or filter in Go.
  - It must match at a **segment boundary**, so `acme-old` does not match `acme`.
  - `LIKE` treats `_` and `%` in paths as wildcards. Either escape them or use `substr(?,1,length(path)+1) = path || '/'`. Doing the match in Go over `ListProjects` is simpler at this scale.
- **Worktrees**
  - Claude Code's worktrees live under `<repo>/.claude/worktrees/…`, which is *inside* the registered repo path. **Longest-prefix resolves them with no special case.** The `worktreeCut` logic can go.
  - A plain `git worktree add ../x` is outside any registered path.
    - Option 1: fall back to `git rev-parse --path-format=absolute --git-common-dir` and take its parent. Verify the flag's git version before relying on it.
    - Option 2: declare that unsupported (v1 doesn't support it either).
  - **Resolution and git facts diverge in a worktree.** The project comes from the registered path, but branch and HEAD must come from the actual `dir` (the worktree). See [history-anchors](history-anchors.md).
- **Not found:**
  - CLI: an error ("not a registered project; run `bp add` or `bp migrate`"). If `<dir>/.bit` exists, give the migrate warning instead ([migrate](migrate.md)).
  - MCP: a tool error with the same text. The server still starts.
- **Returns:** `{ID, Code, Path}`. The store root is `filepath.Join(dataDir, code)` (operator: `~/.local/share/bit/<code>/`). `task.New(root)` is already root-parameterised, so every call site becomes `task.New(p.StoreDir())`.
- **Override:** `--project CODE` / `BIT_PROJECT` for sessions outside any registered path. This is not needed for the minimum scope.
- **What replaces bitdir:** the resolver plus `store.Dir()` (renamed to the `bit` dir). `bitdir_test.go` is deleted. The ~11 CLI test files that `t.Chdir` into a temp dir with `.bit/` need a helper that sets `XDG_DATA_HOME`, opens the db, registers the temp dir and then chdirs.

## Multi-repo (context only; minimum scope has no linking)
- Longest-prefix gives `acme/frontend/src` → frontend and `acme/` → acme. Nested registrations are expected, so `path UNIQUE` stays.
- `code` must become UNIQUE, because it is the directory name.
- Linking (projects↔projects, tracks↔tracks) is later. Nothing in the minimum resolver blocks it.
- ID allocation is glob max+1 with no lock (`task/store.go` ~625-675). The central store does not widen this for single-repo projects, because one project still has one store. Moving ID allocation to sqlite is optional.
