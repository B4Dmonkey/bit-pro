---
id: BIT-39.10
title: The cycle runs under launchd
status: done
approved: true
phase: 4
phase_label: Works with the terminal closed
---
## **Verse 4**

The observation that closes the verse: the cycle still runs when the terminal that started it is
gone. Everything automatable is done by the three bars before this one — `bp start` resolves
`claude` and pins it, and the daemon runs what was pinned. What is left cannot be tested, only
watched.

Two of the verse's original worries were already settled and need no checking here: `claude agents
--json` needs no TTY (its own `--help` says so), and `newHandler` (`cmd/serve.go:26`) already
switches to the JSON handler when stdout is not a character device, which is what launchd gives it.
The third worry — the environment — was the one that failed on first contact, and the bars above are
the fix.

**This bar observes ONE bar running, not a drain.** With the terminal closed nobody is deleting
sessions, and under the slot Decision a project's slot frees only when its session is gone. So the
first bar dispatches and the second correctly does not. A second bar going would be the defect, not
the success.

## Scope
- No production code. Any file this turns out to need means the three bars above missed something —
  stop and say so rather than patching it in here.

## References
- `automation-notes.md` — "Daemon hosting" (settled 2026-08-19) for the plist mechanics.

## Method
- [ ] `just install`, then `bp stop` and `bp start`, so the plist is rewritten from the current
  binary and carries `BP_CLAUDE`. `bp status` reports running.
- [ ] Confirm the log path before walking away: read `StandardOutPath` out of
  `~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist`. It is
  `~/.local/share/bit-pro/daemon.log` unless `XDG_DATA_HOME` is set.
- [ ] Close every Claude session whose cwd is at or beneath the project you are testing, including
  your own interactive one — the guard counts it as live and will hold the project otherwise.
- [ ] Close the terminal. Come back and read the log.

## Claude verifies
- [ ] `just test` and `just lint` still pass — a no-op here, recorded so this bar has the same green
  gate as every other.

## User verifies
- [ ] In `tools/example`, with the terminal closed. `./reset.sh last`, approve `EX-2`'s bars, answer
  `y` at the play prompt, then `bp start` and **close the terminal**.

  Come back and read `~/.local/share/bit-pro/daemon.log`. It holds tick records in JSON — not text —
  and specifically **not** `exec: "claude": executable file not found in $PATH`, which is the exact
  failure the bars above exist to remove. `claude agents --json` shows one session named
  `EX-2-<slug>`, and `EX-2.1`'s queue row is gone while `EX-2.2`'s and `EX-2.3`'s remain.
- [ ] Whole slice: `./check.sh EX-2` shows one commit on
  `worktree-EX-2-shout-dispatch-drain-workload`. One bar's worth of work landed with nothing but
  launchd hosting the loop — which is the capability Verse 4 exists to deliver.
- [ ] Finish with `bp stop`. A daemon left running will dispatch whatever is queued next, which is
  not what you want while planning the following track.

## Commit (user)
`docs(daemon): record the launchd dispatch check`