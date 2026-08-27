---
id: BIT-39.12
title: A spawn that produced no session is not a dispatch
status: todo
---
## Collected, not planned

Found 2026-08-26 while running the track's own verification in `tools/example`. **This bar is not
planned and not approved** — the Method below is a sketch, not settled detail. It needs a
`/bit:plan` pass before anyone runs it.

It carries no verse. Verses 1–3 shipped and their bars pass; this is a hole underneath them that
their tests could not see, because every test fakes the runner and a fake runner never lies about
its exit code.

## The problem

Two defects that are really one, which is why they share a bar.

**`claude` exits 0 when the spawn fails.** Measured by hand:

```
$ claude --agent bit:bot-dev -w EX-2-... -n EX-2-... 'Reply with exactly OK and stop.'
--agent 'bit:bot-dev' not found. Available agents: claude, Explore, general-purpose, Plan, statusline-setup
$ echo $?
0
```

`claude.Spawn` (`claude/dispatch.go:99`) only fails on a non-zero code, so it returns nil and the
loop logs `dispatched` for a session that printed an error and died. The output is captured in
`out` and thrown away. Nothing in the log distinguishes this from a good dispatch.

**Confirmation cannot tell "started" from "survived".** The doomed process still registers in
`claude agents --json` on its way down, so whether the confirm poll catches it is a sub-second
race. Both outcomes were observed in one 30-second run, from the same underlying failure:

```
07:20:05 INFO  dispatched project=EX bar=EX-2.1 worktree=EX-2-shout-dispatch-drain-workload
07:20:09 DEBUG holding a project that has a live session project=EX
07:20:14 DEBUG holding a project that has a live session project=EX
07:20:20 INFO  dispatched project=EX bar=EX-2.2 ...
07:20:20 WARN  dispatched session not visible yet project=EX bar=EX-2.2 ...
```

EX-2.1 won the race: confirmed, row deleted, bar left `todo` with no work done and **no record
anywhere that it was ever attempted**. EX-2.2 lost it: row retained, redispatched every tick
forever. The queue afterwards, with EX-2.1's row simply gone:

```
2  2  EX-2.2  bar
3  2  EX-2.3  bar
```

Losing the row is the worse half. "Dequeue on confirmed spawn" (a Decision on this track) is
sound only if confirmation means the session is doing the work. It currently means the process
existed for an instant.

## Scope
- `claude/dispatch.go` — `Spawn`, which discards `out`.
- `daemon/loop.go` — the confirm-then-delete sequence at `:138-153`.
- `claude/dispatch_test.go`, `daemon/loop_test.go`.

## Method (sketch — not settled)
- [ ] `Spawn` should not treat exit 0 as success on its own. It already holds the combined
      output; a spawn that printed a diagnostic and produced no session is a failed spawn.
- [ ] Open question for planning: what is the honest success signal? Candidates are parsing the
      `backgrounded · <id> · <name>` line (a Decision on this track deliberately rejected this),
      matching known error text (brittle), or keeping the confirm poll but requiring the session
      to still be present on the *next* tick before dequeuing. The last one costs one tick of
      latency and needs no new surface, which makes it the one to price first.
- [ ] Whichever is chosen, a failed spawn must log at ERROR or WARN with the captured output. The
      operator currently gets `INFO dispatched` and nothing else.
- [ ] Open question: should a bar that was dispatched and did not survive be re-queued, or left
      for the operator? The "a session that ends without landing its bar is not the loop's
      problem" Decision says the latter — but that Decision assumed the session *ran*.

## Claude verifies
- [ ] `just test`, `just lint`.
- [ ] A table test where the fake runner returns exit 0 with error text on stdout and no matching
      agent row — the loop must not report success and must not delete the row.

## User verifies
- [ ] With the bit plugin still missing from `tools/example`, start the daemon and confirm the
      failure is legible in the log and the queue row survives.

## Commit (user)
`fix(daemon): a spawn that produced no session is not a dispatch`