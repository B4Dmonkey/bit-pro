---
id: BIT-39.10
title: The cycle runs under launchd
status: doing
phase: 4
phase_label: Works with the terminal closed
---
## **Verse 4**

The same cycle under the launchd-hosted daemon, so `bp start` and closing the terminal is enough.
**No code is planned here.** Two things the verse worried about are already settled: `claude
agents --json` needs no TTY (its `--help` says so outright), and `newHandler`
(`cmd/serve.go:53`) already picks the JSON handler when stdout is not a character device, which is
what launchd gives it. And the environment question was decided away — `claude` is assumed to be on
the `PATH`, so the plist names no absolute path and gets no `EnvironmentVariables`. Nothing in the
loop reads a relative path either, so `WorkingDirectory` stays unset.

This bar is therefore the observation that closes the verse. If it fails, `PATH` is the first
suspect — a LaunchAgent's is not the operator's login `PATH`, and `claude` lives in
`~/.local/bin` on this machine — and the fix is a new Decision in the track, not an edit here.

## Scope
- Nothing expected. Any file this turns out to need is a signal the verse was mis-scoped — stop and
  say so rather than patching it in.

## References
- `automation-notes.md` — "Daemon hosting" (settled 2026-08-19) for the plist mechanics, and the
  2026-08-21 entry establishing `claude agents --json` as the only poll surface without a TTY.

## Method
- [ ] `just install`, then `bp stop` and `bp start` so the plist is rewritten from the current
  binary. `bp status` reports running.
- [ ] Enqueue one real approved, not-done bar from `bp tui`.
- [ ] Close every Claude session in this repo, including the interactive one — the Verse 2 guard
  counts an operator's own session as live and will hold the project otherwise. Close the terminal.
- [ ] Come back and read `~/Library/Logs/` (or wherever the plist's `StandardOutPath` points —
  check `daemon.Plist`) for the daemon's JSON records.

## Claude verifies
- [ ] `just test` and `just lint` still pass — a no-op here, recorded so the bar has a green gate
  like every other.

## User verifies
- [ ] Whole slice, in `tools/example`, with the terminal closed. `./reset.sh last`, approve `EX-2`'s
  three bars, answer `y` at the play prompt, then `bp start` and **close the terminal**.

  With no interactive session in `tools/example`, the bars still drain: `claude agents --json` shows
  each session come up, the queue rows clear, and `./check.sh EX-2` ends with three commits on
  `worktree-EX-2-shout-dispatch-drain-workload`. Then read `~/.local/share/bit-pro/daemon.log` — it
  holds the tick records, and specifically *not* `claude: executable file not found in $PATH`, which
  is the exact failure the plist's no-`EnvironmentVariables` assumption would produce.

  Finish with `bp stop`. A daemon left running will dispatch whatever is queued next, which is not
  what you want while planning the following track.

## Report back
- [ ] If it fails on `PATH`, take that to bit_scope: how the daemon locates `claude` becomes a
  Decision, and this verse gets a real bar.

## Commit (user)
`docs(daemon): record the launchd dispatch check`