---
id: BIT-28
title: daemon lifecycle (serve / start / stop / status)
status: done
---
## Why

The automation loop `bp` is building toward needs a long-running process to drive it. Today there's no way to start one and no answer to "is it running?" except checking the process table by hand. Worse, an unsupervised process is the wrong foundation for the thing it will eventually do: a daemon that dies mid-track, or silently fails to come back after a reboot, leaves queued bars sitting undispatched with nothing to say so. This track closes the on-ramp gap — a daemon an operator can run and watch, and a supervised background one they can start, stop, and ask about.

## Summary

Add four top-level commands. `bp serve` **is** the daemon: a plain foreground process that runs the loop attached to the terminal. `bp start`, `bp stop`, and `bp status` are thin `launchctl` wrappers that put that same body under **launchd** as a per-user LaunchAgent — `bp start` writes the agent's plist on first run (pointing at `bp serve`) and bootstraps it, `bp stop` tears it down durably, `bp status` reports whether it is running. State — logs, and later the registry DB — lives in `~/.local/share/bit-pro/`. No loop logic in this track: the body is a stub. This track proves the lifecycle works.

## Visual aid

```
~/Library/LaunchAgents/<label>.plist        ← written on first `bp start`
   │  ProgramArguments  → /abs/path/to/bp serve   (os.Executable())
   │  StandardOutPath   → ~/.local/share/bit-pro/daemon.log
   └─ launchd auto-loads this directory at login

bp serve   ─▶ the daemon itself: runs the loop in the foreground and exits cleanly on
             Ctrl-C / SIGTERM. Silent at the default `info` level; `bp serve -v` drops to
             `debug` and shows each tick. What the plist points at, and how an operator
             watches the loop live instead of tailing a log.

               $ bp serve                 → (silence until something happens)
               $ bp serve -v              → time=… level=DEBUG msg=tick   (every 10s)

bp start   ─▶ plist missing?      → write it
             launchctl enable gui/$UID/<label>      (mandatory: bootstrap of a disabled
                                                     label fails, exit 5)
             not loaded?          → launchctl bootstrap gui/$UID <plist>
             loaded, not running? → launchctl kickstart gui/$UID/<label>
             running?             → "already running (pid N)"
             otherwise            → "started (pid N)"

bp stop    ─▶ launchctl bootout  gui/$UID/<label>      (unloads + kills, this session)
             launchctl disable gui/$UID/<label>      (persists: stays down after reboot)
             → "stopped"

bp status  ─▶ launchctl print-disabled gui/$UID  → parse the *value*, not the key:
               "<label>" => disabled  → "stopped"
               "<label>" => enabled   → fall through
               label absent           → fall through
             launchctl list <label>
               dict with "PID" = N  → "running (pid N)"
               dict, no PID         → "not running"
               exit 113             → "not running"

reboot     ─▶ launchd walks ~/Library/LaunchAgents/, honours the disabled store:
               stopped via bp stop → stays down
               otherwise           → loads + RunAtLoad starts it
```

## Decisions

- **launchd hosts the daemon; `bp` does not fork itself.** Replaces the earlier `cmd.Start()` +
  `syscall.SysProcAttr{Setsid: true}` decision. A Setsid orphan survives the terminal but nothing
  supervises it — no restart if it dies mid-track, no start at login, and its TCC/permission
  identity is inherited from whichever terminal spawned it. launchd owns the pid, liveness, restart
  policy, and log redirection, which makes this route *less* Go code: no pid file, no liveness
  probe, no orphan cleanup.
- **A per-user LaunchAgent in the `gui/$UID` domain, not a system LaunchDaemon.** The loop spawns
  Claude sessions as the operator and needs their environment and keychain. A LaunchDaemon runs as
  root before login, which is the wrong identity, and unsigned binaries under
  `/Library/LaunchDaemons` draw Gatekeeper/TCC prompts.
- **The agent's label is `com.github.b4dmonkey.bit-pro`.** Reverse-DNS by convention, derived from
  the module path. It is the plist's filename (`~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist`)
  and the service name in every `launchctl` call.
- **The daemon body is its own command, `bp serve` — not a flag on `bp start`.** Replaces the earlier
  `bp start --fg` decision. Since launchd hosts the daemon, `bp start` and the daemon are two
  unrelated programs: one shells out to `launchctl`, the other runs the loop. A flag that swaps a
  command's whole job is the wrong shape, and it makes the plist point `bp start` at `bp start`,
  which reads as recursion to anyone auditing it. Two verbs, one job each.
- **The name is `serve`, not `dispatch` or `srv`.** `dispatch` already means the shelved chaining
  design's one-shot "run the next bar and exit" (BIT-25), so reusing it for a long-running loop
  would mislead anyone who read that history. `srv` is the right idea abbreviated; Cobra convention
  spells the verb out.
- **`bp serve` is documented surface, not a hidden command.** Running the loop in the terminal is a
  real operator affordance — the way to watch it live rather than tailing a log — so it appears in
  `bp --help` like any other command.
- **`bp serve` is the only foreground entry point in this track.** No separate one-shot "run a single
  bar and exit" command. Under the daemon route the loop *is* the dispatcher; a second entry point
  would be a second definition of how work gets picked up.
- **The four commands are top-level: `bp serve`, `bp start`, `bp stop`, `bp status`.** No `bg` parent
  group. There is one daemon, so a group buys nothing, and it matches the rest of the daemon-facing
  surface (`bp add`, `bp list`) rather than splitting it across two shapes. Nothing already
  registered on `rootCmd` collides.
- **`bp start` writes the plist on first run — `bp init` does not.** `bp init` is per-project
  (`claude.WriteSettings`, `cmd/init.go:48`); the agent is machine-global, so it does not belong to
  project enrollment. `ProgramArguments[0]` is resolved with `os.Executable()`, so the plist points
  at whichever binary generated it.
- **`bp start` reconciles state rather than assuming it.** Plist missing, job not loaded, job loaded
  but not running, and job already running are four distinct cases and `bp start` handles all four.
  This is what makes it safe to run twice.
- **`bp stop` is durable, and the order is `bootout` then `disable` — measured, not reasoned.**
  Stopping is an explicit operator intent that outlives the login session, so `bp stop` is `bootout`
  *and* `disable`: `bootout` alone is session-scoped, and launchd re-loads `~/Library/LaunchAgents/`
  at login, which would silently resurrect a daemon the operator had stopped. Conversely a daemon
  that was never stopped comes back on its own after a reboot, because that is the point of using
  launchd. Both orderings were run against a throwaway label and reach the same end state, so the
  order is chosen on the interrupted case: `disable` does **not** kill a running job (pid survived it),
  so disabling first leaves the daemon alive while marked disabled — and since `bp status` checks
  `print-disabled` first, it would report `stopped` about a live daemon. Booting out first fails
  honestly instead: the daemon is genuinely down, it just comes back at the next login.
- **`bp stop` disables the agent; it never deletes the plist.** Deleting would discard the
  configuration and is not what "stop" means. The disabled store is the mechanism.
- **`bp status` reports three states, not two: `running`, `not running`, `stopped`.** The durable
  stop makes "deliberately stopped, and will not come back at login" distinct from "not currently
  running" — collapsing them would hide the one fact an operator needs when queued bars aren't
  moving. `launchctl print-disabled gui/$UID` is the source for `stopped`.
- **State dir is `~/.local/share/bit-pro/`**, created on first use. Follows XDG convention, survives
  reboots, findable without knowing which project you're in. Holds `StandardOutPath` /
  `StandardErrorPath` and later BIT-29's `state.db`. It holds **no pid file** — launchd owns the pid.
- **`KeepAlive` restarts on crash only** (`{SuccessfulExit: false}`), not unconditionally. A daemon
  that dies mid-track should come back; one that exits cleanly meant to. `bootout` unloads regardless
  of `KeepAlive`, so this does not fight `bp stop`.
- **macOS only for now.** launchd is macOS-only and the Linux equivalent is a systemd user unit. The
  repo has no `GOOS`-specific code or build tags today, so this track introduces the first platform
  boundary — but the self-fork route is unsupervised on Linux too, so portability is not a reason to
  prefer it.
- **No loop logic in this track.** The daemon body is a stub. The loop is a later track — this one
  only proves the lifecycle plumbing works.

- **The stub loop ticks every 10 seconds.** Revised down from 30s during BIT-28.3: while the body is
  a stub the tick is the only liveness signal an operator watching `bp serve -v` gets, and 30s is long
  enough to read as hung. A bar takes minutes to run, so pickup latency is negligible either way once
  the real loop lands, and `daemon.log` stays skimmable (~8,600 tick lines/day at `debug`, none at the
  default level).
- **The daemon logs `started` and `stopped` at `info`.** Added during BIT-28.3. Without them the
  default level emits nothing ever, so `daemon.log` cannot answer whether the daemon came up — the
  exact silence the Why calls out. `stopped` is deferred, so it also covers the real loop's later error
  exits; it only appears on Ctrl-C once BIT-28.4 wires the signal to the root context.
- **A tick logs at `debug`; `bp serve` defaults to `info` and `-v` opts in.** Ticks are the operator's
  liveness signal while watching, and noise in a log kept for real events. Default `info` means plain
  `bp serve` prints nothing until something happens, and the plist passes no level argument — the
  daemon is quiet because that is the default, not because it was configured to be.
- **Logging is `log/slog`: the text handler when stdout is a terminal, JSON otherwise.** Text is
  readable while watching `bp serve -v` live, which is the affordance the command exists for; under
  launchd stdout is `daemon.log`, so the same binary writes JSON for anything that later reads it
  back. The tradeoff accepted here is that redirecting `bp serve` to a file yields JSON, not text.
- **`bp start` and `bp stop` report the terminal state only: `started (pid N)`, `already running
  (pid N)`, `stopped`.** No per-action narration of the `launchctl` calls or the first-run plist
  write. This matches `bp status`'s three-word vocabulary and keeps the operator surface about *what
  state the daemon is in* rather than *what was done to get it there*.
- **`bp start` must `enable` before `bootstrap` — it is mandatory, not hygiene.** Measured: with the
  label in the disabled store, `launchctl bootstrap` fails with `Bootstrap failed: 5: Input/output
  error` and exit 5. Without the `enable`, `bp start` after any `bp stop` would hard-fail rather than
  restart.
- **`bp status` reads `print-disabled` by value, never by the label's presence.** Measured:
  `launchctl enable` flips the entry to `"<label>" => enabled` rather than removing it, so the label
  stays in the store forever after the first `bp stop`. Matching on the key alone would report
  `stopped` for the rest of the machine's life. The three shapes are: label absent (never disabled),
  `=> disabled`, `=> enabled`.
## Verses

- [x] Verse 1 — Operator can run the daemon in their terminal and watch it: `bp serve` runs the stub
  loop attached to the terminal and exits cleanly on Ctrl-C / `SIGTERM`; `bp serve -v` shows each
  10s tick. Nothing launchd-related yet — this is the walking skeleton the plist will later point at,
  and it fails cheap if the body can't run at all.
  Touches: `cmd/serve.go` (new), `cmd/root.go` (register it) — where to look to verify.
- [x] Verse 2 — Operator can manage a supervised background daemon: `bp start` writes the LaunchAgent
  plist on first run and bootstraps it, `bp status` reports `running` / `not running` / `stopped`,
  `bp stop` brings it down durably. It survives closing the terminal, comes back after a crash, and
  comes back after a reboot *unless* `bp stop` was called. The reboot half of that is the operator's
  to verify — nothing automated can prove it.
  Touches: `cmd/start.go`, `cmd/stop.go`, `cmd/status.go` (new), plus the plist generation and state
  dir helpers — where to look to verify.

## References

- `automation-notes.md` (repo root, untracked) — the working notes for the whole automation phase.
  Its "Todo → 1. Daemon lifecycle" entry is this track; its **"Daemons on macOS"** section is the
  authority for the launchd mechanics, including the two staleness cases (a replaced binary needs
  `kickstart -k`; a changed plist needs `bootout` then `bootstrap`) and the measured
  `launchctl list <label>` output shape. Its Decisions and Measured-facts sections are the authority
  for what the daemon later has to drive. Informs both verses.
  **Note:** its line 187 says the `bootout`/`disable` ordering is unverified and should be settled
  with a throwaway label at plan time. That was done during this scope pass — the Decisions above
  supersede it, along with two further measured facts (bootstrap of a disabled label exits 5;
  `print-disabled` must be parsed by value). Don't re-run the probe.