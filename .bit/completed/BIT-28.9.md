---
id: BIT-28.9
title: XDG_DATA_HOME moves the state directory
status: done
phase: 2
phase_label: lifecycle
---
## **Verse 2**

`$XDG_DATA_HOME`, when set, moves the state directory — and with it the daemon log the plist points
at. Contradicts BIT-28.8, which hardcoded `~/.local/share`.

## Scope
- `store/store.go` — branch on `$XDG_DATA_HOME`
- `store/store_test.go`, `cmd/start_test.go` — the new tests

The scope calls this directory "XDG convention", and honouring the variable is what that convention
actually is. `XDG_DATA_HOME` is **unset on this machine** (checked 2026-08-19), so the live path
stays `~/.local/share/bit-pro/` exactly as the scope's Decisions name it — this bar adds the branch
without moving anything today. Both states get a test precisely because only one of them is
observable here.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDir_FollowsXDGDataHome` (table-driven, in `store/store_test.go`)
     - **Behavior:** the daemon's state lives where the environment says user data lives, so an
       operator who has moved their XDG root does not end up with a stray `~/.local/share`.
     - **Setup:** two rows. (a) `t.Setenv("XDG_DATA_HOME", "")` and `t.Setenv("HOME", t.TempDir())`.
       (b) `t.Setenv("XDG_DATA_HOME", t.TempDir())`.
     - **Assertions:** (a) returns `$HOME/.local/share/bit-pro`; (b) returns
       `$XDG_DATA_HOME/bit-pro`. In both rows the returned path exists and `os.Stat` reports a
       directory.
     - **Boundary:** `XDG_DATA_HOME` in both of its states — unset (the fallback, and the only one
       live on this machine) and set (the branch this bar adds).
   - [ ] `TestStartCmd_LogPathFollowsXDGDataHome` (in `cmd/start_test.go`)
     - **Behavior:** the branch reaches the artifact that matters — the plist's log path, which is
       the only place the state directory is written down for launchd to read.
     - **Setup:** `t.Setenv("HOME", t.TempDir())`, `t.Setenv("XDG_DATA_HOME", t.TempDir())`, the
       same all-113 fake runner as BIT-28.8; run `runWithDaemon(t, lc, startCmdUse)`.
     - **Assertions:** the written plist contains `$XDG_DATA_HOME/bit-pro/daemon.log` and does not
       contain `.local/share`.
     - **Boundary:** `XDG_DATA_HOME` set — the state BIT-28.8 could not reach.
   - [ ] Confirm fails: both fail with the path under `$HOME/.local/share/bit-pro`, want it under
         `$XDG_DATA_HOME/bit-pro`.

2. **Implement (GREEN):**
   - [ ] `store/store.go`: in `Dir()`, read `os.Getenv("XDG_DATA_HOME")` first; when it is non-empty,
         use it as the base instead of `~/.local/share`. Keep the existing `os.MkdirAll` and the
         `bit-pro` leaf for both branches, so the two paths differ only in their base.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(store): honour XDG_DATA_HOME for the state directory`