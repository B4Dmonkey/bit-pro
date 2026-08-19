---
id: BIT-28.8
title: bp start writes the plist on first run
status: done
phase: 2
phase_label: lifecycle
---
## **Verse 2**

`bp start` writes the LaunchAgent plist on first run and leaves an existing one alone. No
`launchctl` calls yet and no output — this bar only proves the pointer file that ties the binary,
the log, and launchd together is generated correctly.

## Scope
- `store/store.go` — new package: `Dir() (string, error)`, returning `~/.local/share/bit-pro` and
  creating it. It is deliberately not part of `launchd`: BIT-29's `state.db` lives in the same
  directory and must not have to import the launchd package to find it.
- `launchd/plist.go` — new: `PlistPath() (string, error)`, `Plist(exe, logPath string) []byte`,
  `WritePlist(path string, data []byte) error`
- `cmd/start.go` — new; `newStartCmd(lc launchd.Runner)`
- `cmd/root.go` — register it
- `store/store_test.go`, `launchd/plist_test.go`, `cmd/start_test.go` — the new tests

`ProgramArguments[0]` comes from `os.Executable()`, so the plist points at whichever binary
generated it. Under `go test` that is the test binary — the assertion below reads `os.Executable()`
in the test rather than hardcoding a path, which is honest about what is being proven: that the
resolved path is embedded, not which path it is.

## References
- `automation-notes.md` (repo root, untracked) — "Daemons on macOS", "Plist gotchas": agents inherit
  almost no environment, so every path in the file is absolute. `man launchd.plist` is the
  authoritative key list.

## TDD cycle

1. **Write test (RED):** `cmd/start_test.go`
   - [ ] `TestStartCmd_WritesThePlistOnlyWhenMissing` (table-driven)
     - **Behavior:** enrolling the daemon with launchd is a first-run side effect of `bp start`, not
       a separate install step and not part of the per-project `bp init` — and re-running `bp start`
       never clobbers a plist an operator may have edited.
     - **Setup:** `t.Setenv("HOME", t.TempDir())` and `t.Setenv("XDG_DATA_HOME", "")`; a fake
       `launchd.Runner` returning `("", 113, nil)` for everything. Two rows: (a) no plist on disk;
       (b) `$HOME/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist` pre-written with the
       sentinel bytes `not a plist`. Run `runWithLaunchd(t, lc, startCmdUse)`.
     - **Assertions:** row (a) — the file exists at
       `$HOME/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist` and its contents contain, as
       substrings: the label; the value of `os.Executable()` read in the test; `<string>serve</string>`;
       `$HOME/.local/share/bit-pro/daemon.log` appearing under both `StandardOutPath` and
       `StandardErrorPath`; `<key>RunAtLoad</key>` followed by `<true/>`; and a `KeepAlive` dict
       whose `SuccessfulExit` is `<false/>`. Also `$HOME/.local/share/bit-pro` exists as a directory.
       Row (b) — the file still reads exactly `not a plist`.
     - **Boundary:** the plist's presence in both states — the missing end is first-run enrollment,
       the present end is the idempotence that makes `bp start` safe to run twice.
   - [ ] `TestPlist_KeepAliveRestartsOnCrashOnly` (in `launchd/plist_test.go`)
     - **Behavior:** a daemon that dies mid-track comes back; one that exited cleanly meant to.
       `KeepAlive` is a dict, not a bare `<true/>`, and the difference is the whole policy.
     - **Setup:** `Plist("/usr/local/bin/bp", "/tmp/daemon.log")`.
     - **Assertions:** the output contains a `<key>KeepAlive</key>` whose value is a `<dict>`
       containing `<key>SuccessfulExit</key>` and `<false/>`; it does not contain
       `<key>KeepAlive</key>` immediately followed by `<true/>`.
     - **Boundary:** `SuccessfulExit` in its false state — the only value that distinguishes
       restart-on-crash from restart-always.
   - [ ] Confirm fails: `unknown command "start" for "bp"`, and `undefined: Plist`.

2. **Implement (GREEN):**
   - [ ] `store/store.go`: `Dir() (string, error)` — `os.UserHomeDir()`, join
         `.local/share/bit-pro`, `os.MkdirAll(dir, 0o755)`, return it. The `XDG_DATA_HOME` branch is
         BIT-28.9's job; hardcoding the home-relative path is correct for this bar.
   - [ ] `launchd/plist.go`: `PlistPath() (string, error)` — `os.UserHomeDir()` joined with
         `Library/LaunchAgents/` + `Label` + `.plist`; `os.MkdirAll` its parent directory, since a
         fresh account may not have `~/Library/LaunchAgents` yet.
   - [ ] `launchd/plist.go`: `Plist(exe, logPath string) []byte` — render the XML from a
         `text/template` with `Label`, `ProgramArguments` of `[exe, "serve"]`, `RunAtLoad` true,
         `KeepAlive` as a dict with `SuccessfulExit` false, and `StandardOutPath` /
         `StandardErrorPath` both set to `logPath`.
   - [ ] `launchd/plist.go`: `WritePlist(path string, data []byte) error` — write with `0o644`.
   - [ ] `cmd/start.go`: `const startCmdUse = "start"`; `newStartCmd(lc launchd.Runner)` with
         `Args: cobra.NoArgs`; `RunE` resolves `os.Executable()`, `store.Dir()`, and `PlistPath()`,
         and calls `WritePlist` only when `os.Stat` on the plist path returns `fs.ErrNotExist`.
         Print nothing — output is BIT-28.10's contradiction.
   - [ ] `cmd/root.go`: `rootCmd.AddCommand(newStartCmd(lc))`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(start): write the LaunchAgent plist on first run`