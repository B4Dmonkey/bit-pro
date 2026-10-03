# XDG data dir

**Checked:** Q7 — existing helper? Is `$XDG_DATA_HOME` straightforward on macOS?

## What the code does
`store/store.go` `Dir()` already implements it: `$XDG_DATA_HOME` if set, else `~/.local/share`, then `/bit-pro`, `MkdirAll` 0755. Used by `db.Open()` and `cmd/start.go` (daemon log). No third-party xdg lib in go.mod. Tests (`store_test`, `db/*_test`, `cmd/{add,list,serve,start,status}_test`, `daemon/loop_test`) set `XDG_DATA_HOME` to temp dirs — so the override is already the test seam.

Go's stdlib has no `UserDataDir`; `os.UserConfigDir()` on macOS returns `~/Library/Application Support`, so hand-rolling (as done) is the right call if `~/.local/share` is wanted.

## Gotchas
- **launchd doesn't inherit shell env.** The daemon plist sets only `BP_CLAUDE`. If the operator exports `XDG_DATA_HOME` in their shell, CLI and daemon resolve different dirs. Fix: bake the resolved dir into the plist `EnvironmentVariables` (or pass a flag) at `bp start`.
- MCP servers inherit Claude Code's environment (usually the login shell's), so usually consistent with the CLI — unverified for GUI-launched Claude.
- XDG spec says relative `XDG_DATA_HOME` must be ignored; current code accepts it (`filepath.Clean` only). Minor.
- The dir name changes from `bit-pro` to `bit` in the plan — v2 should parameterise it (a constant per build) rather than edit `store.Dir()` in place, to keep v1 pointing at `bit-pro`.

**Verdict:** straightforward; helper exists; one real gotcha (launchd env).
