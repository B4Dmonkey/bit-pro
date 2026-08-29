---
id: BIT-39.15
title: bp start resolves claude and pins it in the plist
status: done
approved: true
phase: 4
phase_label: Works with the terminal closed
---
## **Verse 4**

`bp start` resolves `claude` in the operator's shell and pins the absolute path into the plist it
already writes. This exists because the `PATH` assumption was measured and failed: a LaunchAgent
inherits launchd's `/usr/bin:/bin:/usr/sbin:/sbin`, `claude` lives in `~/.local/bin`, and the
daemon logged `exec: "claude": executable file not found in $PATH` on every tick. `bp start` is
the only place in the chain that can still see the operator's own `PATH`.

Nothing reads `BP_CLAUDE` yet — that is the third bar. This one makes the failure loud where an
operator is watching, and puts the path where the daemon will be able to find it.

## Scope
- `cmd/start.go` — `enrollDaemon` looks `claude` up with `exec.LookPath` and returns a wrapped
  error if it is absent, before it touches the plist.
- `daemon/plist.go` — `Plist(exe, logPath, claudeBin string)` renders an `EnvironmentVariables`
  dict carrying `BP_CLAUDE`.
- `daemon/plist_test.go` — the two existing `Plist(...)` calls gain the third argument.
- `cmd/start_test.go` — the existing start tests get a `PATH` holding a stub `claude`, so they
  stop depending on whatever `PATH` the developer's shell happens to have.

`enrollDaemon` already compares the canonical bytes against what is on disk and boots the agent
out before rewriting, so a plist that gains this key is repaired on the next `bp start` with no
extra work here.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStartCmd_PinsTheResolvedClaudeInThePlist`
     - **Behavior:** the plist `bp start` writes names the absolute `claude` the operator's
       `PATH` resolves to, so the daemon is never asked to search a `PATH` it does not have.
     - **Setup:** `home := t.TempDir()`; `t.Setenv("HOME", home)`; `t.Setenv("XDG_DATA_HOME", "")`.
       A second `t.TempDir()` holding an executable file named `claude` (mode `0o755`), with
       `t.Setenv("PATH", thatDir)`. Then `runWithDaemon(t, nothingLoaded, startCmdUse)`.
     - **Assertions:** the plist at
       `<home>/Library/LaunchAgents/<daemon.Label>.plist` matches
       `<key>EnvironmentVariables</key>\s*<dict>\s*<key>BP_CLAUDE</key>\s*<string>` +
       `regexp.QuoteMeta(filepath.Join(dir, "claude"))` + `</string>\s*</dict>`.
     - **Boundary:** `PATH` holds exactly one `claude` — the found branch of `exec.LookPath`, and
       the resolved value is an absolute path rather than the bare name, which is the only shape
       launchd can use.
   - [ ] Confirm fails: `Plist` takes two arguments, so the package does not compile; once the
     signature is widened, the plist has no `EnvironmentVariables` key and the regex does not match.

2. **Implement (GREEN):**
   - [ ] `daemon/plist.go`: add `ClaudeBin` to the template's field struct and an
     `EnvironmentVariables` dict to `plistTemplate` rendering `BP_CLAUDE`; widen `Plist` to
     `Plist(exe, logPath, claudeBin string)`.
   - [ ] `daemon/plist_test.go`: pass a third argument at both call sites.
   - [ ] `cmd/start.go`: in `enrollDaemon`, call `exec.LookPath("claude")` before `daemon.Plist`
     and pass the result through.
   - [ ] `cmd/start_test.go`: give every existing start test a `PATH` containing a stub `claude`.

3. **More tests (RED → GREEN):**
   - [ ] `TestStartCmd_FailsWhenClaudeIsNotOnThePath`
     - **Behavior:** `bp start` refuses rather than enrolling a daemon that cannot run `claude`,
       so the failure surfaces in the operator's terminal instead of once per tick in a log.
     - **Setup:** same `HOME`/`XDG_DATA_HOME` setup; `t.Setenv("PATH", t.TempDir())` — an empty
       directory. Run `runWithDaemon(t, nothingLoaded, startCmdUse)`.
     - **Assertions:** the returned error is non-nil and its message names `claude`. No plist
       exists at the expected path (`os.Stat` returns `fs.ErrNotExist`).
     - **Boundary:** `PATH` holds zero `claude` — the not-found branch of `exec.LookPath`, and
       the lookup happens before any write, so nothing is left half-enrolled.
   - [ ] Confirm fails: `bp start` currently succeeds and writes a plist regardless of `PATH`.
   - [ ] Return `fmt.Errorf("locating the claude binary: %w", err)` from `enrollDaemon` before it
     reads or writes the plist.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(daemon): pin the resolved claude path in the plist`