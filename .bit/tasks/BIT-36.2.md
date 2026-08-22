---
id: BIT-36.2
title: Contradiction forces plist to emit serve daemon
status: todo
phase: 1
phase_label: serve daemon
---
## **Verse 1**

Bar 1's daemon runs but the plist template still emits `<string>serve</string>` alone — launchd would point at the parent command which prints help and exits 0. A new test on the `Plist` return value contradicts the template and forces the addition of `<string>daemon</string>`.

This is the last bar of Verse 1, so it carries the integration check.

## Scope
- `daemon/plist.go` — add `\t\t<string>daemon</string>` line to `plistTemplate` after `<string>serve</string>`
- `daemon/plist_test.go` — add `TestPlist_ProgramArgumentsIncludesDaemon`
- `cmd/start_test.go` — update `assertEnrollsTheDaemon` to check for `daemon` in ProgramArguments

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestPlist_ProgramArgumentsIncludesDaemon` in `daemon/plist_test.go`
     - **Behavior:** the rendered plist's `ProgramArguments` array contains both `serve` and `daemon` as separate `<string>` entries
     - **Setup:** `plist := string(Plist("/usr/local/bin/bp", "/tmp/daemon.log"))`
     - **Assertions:** plist contains `"<string>serve</string>"` and `"<string>daemon</string>"` (in that order); use `strings.Index` to verify ordering if needed
     - **Boundary:** `ProgramArguments` count == 2 — proves the two-element slice, not a single combined string
   - [ ] Confirm fails: template produces only `<string>serve</string>`, not `<string>daemon</string>`

2. **Implement (GREEN):**
   - [ ] In `daemon/plist.go`, in `plistTemplate`, add `\t\t<string>daemon</string>` on the line immediately after `\t\t<string>serve</string>`

3. **More tests (RED → GREEN):**
   - [ ] Update `assertEnrollsTheDaemon` in `cmd/start_test.go`: change the `"<string>serve</string>"` entry in the `want` slice to check for both `"<string>serve</string>"` and `"<string>daemon</string>"` (add `daemon` as a separate entry in the wants list)

## Claude verifies
- [ ] `just test ./cmd/... ./daemon/...` passes
- [ ] `just lint` passes

## User verifies
- [ ] Verse 1 integration: `just install && bp start` — output says `started (pid …)` or `already running (pid …)`; `cat ~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist` — contains `<string>serve</string>` followed by `<string>daemon</string>` in `ProgramArguments`; `bp serve daemon --help` prints daemon-loop help; `bp serve --help` lists `daemon` as a subcommand

## Commit (user)
`feat(serve): plist emits serve daemon in ProgramArguments`