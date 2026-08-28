---
id: BIT-39.11
title: The daemon's hold message says what it means
status: todo
approved: true
---
## Collected, not planned

A cleanup item noticed while verifying BIT-39, parked here so it isn't forgotten. **This bar is
not planned and not approved** — the Method below is a sketch, not settled detail. It needs a
`/bit:plan` pass before anyone runs it.

It carries no verse: it serves none of the four value slices, it landed after they were written,
and it is polish rather than delivery.

## The problem

`daemon/loop.go:80` logs, once per project per tick:

```
level=DEBUG msg="holding a project that has a live session" project=BIT
```

Observed 2026-08-26 while running Verse 1's cadence check. The line is accurate and unreadable —
it names a condition without naming its consequence or its cause. Three things an operator needs
from it and doesn't get:

- **What is being withheld.** "Holding" doesn't say that the project's *queue* is untouched and
  that nothing will dispatch. A reader can't tell normal idling from a stall.
- **Which session.** The guard matched some live row's `cwd` at or beneath the project path, and
  in practice that row is very often the operator's own interactive session — which is the whole
  reason the line appears at all. It logs neither the matched `cwd` nor the session name, so
  there's no way to tell *which* session to close.
- **That closing it is the fix.** The behaviour is by design — a BIT-39 Decision makes liveness
  mean presence in `claude agents --json`, so a human working a project holds that project's
  queue. The message reads like a fault instead.

The failure mode is concrete: the line repeats every 5s, the operator reads it as a broken
daemon, and goes looking in the wrong place.

## Scope
- `daemon/loop.go` — the guard at `:79-83`.
- `daemon/loop_test.go` — whatever pins the new fields.

## Method (sketch — not settled)
- [ ] Probably more than wording. Naming the blocking session means keeping the row the guard
  matched rather than discarding it: `slices.ContainsFunc` answers yes/no and throws the row
  away, so it likely becomes a find that returns the `claude.Agent`.
- [ ] Then the record can carry the session's name and `cwd` alongside `project`, and the message
  can state the consequence ("not dispatching") rather than the condition ("holding").
- [ ] Open question for planning: whether `DEBUG` is the right level. An operator running
  `bp serve daemon -v` to find out why nothing dispatches is the exact audience for this line,
  and under launchd it may not be recorded at all.

## Claude verifies
- [ ] `just test`, `just lint`.

## User verifies
- [ ] Run `bp serve daemon -v` with a Claude session open in the repo and read the line cold —
  it should be obvious which session to close and that closing it is what starts dispatch.

## Commit (user)
`feat(daemon): name the session that holds a project`