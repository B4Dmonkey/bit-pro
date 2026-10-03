# Testing a v2 build before cutover (Q7)

**Checked:** how to exercise v2 without `just install` replacing the daily `bp`, and how MCP sessions can point at a dev build. This supersedes risks 1-3 of [coexistence](coexistence.md), which assumed side-by-side use. The operator chose a hard cutover.

## Facts
- `scripts/install.sh` always builds `$(go env GOBIN || GOPATH/bin)/bp`, which is `~/go/bin/bp` here. **Never run `just install` on `v2` before cutover.**
- The `Justfile` already has `run *ARGS: db-gen-queries` → `go run . {{ARGS}}`.
- **v1 and v2 data dirs already differ:** v1 uses `~/.local/share/bit-pro/`, v2 uses `~/.local/share/bit/` (which does not exist yet). A v2 dev build therefore can't corrupt v1's registry, even without a sandbox.
- v2 migrate *copies* from a project's `.bit/` (see [migrate](migrate.md)), so testing migrate on a live project leaves v1 working. That includes bit-pro itself.
- The one real collision is **`.bit/` writes by a v2 build in a project**. v2 shouldn't write there at all, so this is safe by design.
- `store.Dir()` honours `XDG_DATA_HOME`, and the Go tests already sandbox with it.

## Options
1. **`go run` / `just run` with a sandbox:** `XDG_DATA_HOME=$PWD/.dev-data just run add` etc. No install, and nothing is touched. Good for CLI slices.
2. **Alternate binary:** `go build -o ./bin/bp .` (add `bin/` to .gitignore), or a `just build-dev` target. Call it by absolute path. Don't put it on PATH as `bp`.
3. **MCP pointed at the dev build, per project** (`claude mcp add` syntax verified from `--help`: `-e KEY=value`, `-s local` default):
   `claude mcp add bit -e XDG_DATA_HOME=<sandbox> -- <abs>/bin/bp serve mcp`, run inside a scratch project such as `tools/example`.
   - Local scope is keyed to that project path in `~/.claude.json`, so other projects keep v1.
   - Keeping the server **name `bit`** keeps the tool names `mcp__bit__*`, so the skills work unchanged.
   - Caveat: `bp init`/`add` call `RegisterMCP`, which is a no-op when `claude mcp get bit` exists. It won't overwrite this entry, which is good.
4. **Fully ephemeral session**, which leaves `~/.claude.json` untouched: `claude --mcp-config dev-mcp.json --strict-mcp-config --plugin-dir ./bit`. Both flags were confirmed in `claude --help`.
   - `--plugin-dir` loads the **v2 branch's skills** without pushing to GitHub `main`. That solves "skills ship from main".
   - Unknown: how it interacts with a project that also has `bit@bit-pro` enabled in `.claude/settings.json` (duplicate skills?). Test it in a project without the plugin enabled, or verify.
5. For opencode: point `opencode.json`'s `command` at `<abs>/bin/bp` in a scratch dir. Remember there is no `CLAUDE_PROJECT_DIR` there.

## MCP in Go tests
`runMCPServer(ctx, root, transport)` (`cmd/serve_mcp.go`) plus `cmd/mcp_harness_test.go` already test the server in process. v2 tests set `XDG_DATA_HOME` to `t.TempDir()`, register the temp project in the db and pass `root`. No binary is needed.

## Hazard found
The rehearsal fixture `tools/example` relies on **git-tracked `.bit/`**. `reset.sh:8,129` resets task state through git checkouts and tags, and `checkpoint.sh` works the same way. With central storage, reset/checkpoint no longer restore task state. It needs a v2 rework, e.g. snapshotting `~/.local/share/bit/ex/` next to the git tag, or using a sandboxed `XDG_DATA_HOME` per run. That is a candidate **separate track**.

## Cutover checklist (facts gathered)
1. Merge `v2` → `main`, then `just install`.
2. `bp migrate` in each project: bit-pro, example, a client project.
3. `git rm -r .bit` in bit-pro.
4. Delete `~/.local/share/bit-pro/` and `~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist` (present on this machine, not loaded).
5. The plugin update reaches projects from `main`.
