---
id: BIT-39.17
title: The daemon runs the claude bp start pinned
status: done
approved: true
phase: 4
phase_label: Works with the terminal closed
---
## **Verse 4**

The daemon reads `BP_CLAUDE` — the entry `bp start` wrote into the plist two bars ago — and falls
back to a bare `"claude"` when it is unset, so a foreground `bp serve daemon` in a terminal is
unchanged. This is the bar that closes the `PATH` failure: under launchd the env var is set and
absolute, so the loop stops depending on a `PATH` it does not have.

No flag on `serve daemon`. The "no new operator surface" Decision already forbids one for the tick
interval, and an env var carried by the plist `bp start` writes needs no operator involvement at
all.

## Scope
- `cmd/serve.go` — a `claudeBin()` helper reading `BP_CLAUDE`, and `newServeDaemonCmd` passing it
  to `daemon.Loop` in place of the literal added by the previous bar.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestClaudeBin` — table-driven
     - **Behavior:** the daemon runs the path its plist pinned, and still works when nothing
       pinned one.
     - **Setup:** three subtests, each setting the environment with `t.Setenv("BP_CLAUDE", …)`:
       `"/Users/x/.local/bin/claude"`; the variable unset (`t.Setenv` to `""` — the same state
       `os.Getenv` reports for both, which is why the empty case is in the table); and `""`
       explicitly.
     - **Assertions:** `claudeBin()` returns `"/Users/x/.local/bin/claude"` for the first and
       `"claude"` for the other two.
     - **Boundary:** the env var's states — set to a real absolute path, and empty/unset, which
       `os.Getenv` cannot distinguish and which must therefore both take the fallback rather than
       returning an empty command name that `exec` would reject with a confusing error.
   - [ ] Confirm fails: `claudeBin` does not exist.

2. **Implement (GREEN):**
   - [ ] `cmd/serve.go`: `func claudeBin() string` returning `os.Getenv("BP_CLAUDE")` when
     non-empty and `"claude"` otherwise.
   - [ ] `newServeDaemonCmd`: pass `claudeBin()` to `daemon.Loop` instead of the literal.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] `just install`, then `bp stop && bp start`. Confirm the plist carries the same path your
  shell resolves:

  ```
  plutil -p ~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist | grep -A1 BP_CLAUDE
  which claude
  ```

  The two paths match.
- [ ] With the daemon stopped (`bp stop`), run `bp serve daemon -v` in a terminal with `BP_CLAUDE`
  unset. It still ticks and still lists sessions — no `executable file not found in $PATH`. This is
  the fallback path, and it is the one you use every day, so it has to stay working.

## Commit (user)
`feat(daemon): read the claude path bp start pinned`