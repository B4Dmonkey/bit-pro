---
id: BIT-39.11
title: The daemon's hold message says what it means
status: todo
approved: true
phase: 5
phase_label: Cleanup
---
## **Verse 5**

The hold message names the session holding the project and says what is not happening because of
it. Today `daemon/loop.go:84` logs, once per project per tick:

```
level=DEBUG msg="holding a project that has a live session" project=BIT
```

Accurate and unreadable. It names a condition without its consequence or its cause, it repeats every
5 seconds, and the session it matched is very often the operator's own interactive one — which is the
whole reason the line appears. An operator reads it as a broken daemon and goes looking in the wrong
place.

Two changes, both settled by Decisions on the track: the level moves to `INFO`, because an operator
running `bp serve daemon -v` to find out why nothing dispatches is the exact audience and a `DEBUG`
record may not be written at all under launchd; and the record carries the matched session's name
and `cwd`, which means the guard has to keep the row it matched instead of throwing it away.

The guard's hold/release *rule* does not change — hold while any live row's `cwd` is at or beneath
the project path, release once none is. Only what it reports changes.

## Scope
- `daemon/loop.go` — the guard at `:83-87`. `slices.ContainsFunc` answers yes/no and discards the
  row, so it becomes `slices.IndexFunc` (or a small find returning the `claude.Agent`).
- `daemon/loop_test.go` — the new assertion, and a log helper: the existing `loggedAbout` matches on
  a `bar` attribute, which a hold record does not have.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTick_NamesTheSessionHoldingAProject`
     - **Behavior:** an operator reading the log can tell which session to close and that closing
       it is what starts dispatch — so the record must carry the session's identity, not just the
       project's.
     - **Setup:** `queries, project := queuedBar(t)`. A runner returning one live row:
       `[{"name":"6a4a7973","cwd":"<project.Path>/cmd"}]` — the interactive-session shape, matched
       by the at-or-beneath rule rather than by name. Log to a `bytes.Buffer` through a JSON
       handler at `slog.LevelDebug`, so a record at either level would be captured and the level
       assertion is real.
     - **Assertions:** exactly one record whose `msg` is `not dispatching`; its `level` is `INFO`;
       its attributes are `project == "ACME"`, `session == "6a4a7973"`, and
       `cwd == filepath.Join(project.Path, "cmd")`.
     - **Boundary:** exactly one live row under the project — the minimum that triggers a hold —
       and the level asserted as `INFO`, which is the state the current code is *not* in. The `cwd`
       is a subdirectory rather than the project root, so the record proves the matched row is
       reported rather than the project path being echoed back.
   - [ ] Confirm fails: the record is at `DEBUG`, its `msg` is
     `holding a project that has a live session`, and it carries no `session` or `cwd` attribute.

2. **Implement (GREEN):**
   - [ ] `daemon/loop.go`: replace the `slices.ContainsFunc` guard with `slices.IndexFunc`, keep
     the matched `claude.Agent`, and log
     `log.Info("not dispatching", "project", p.Code, "session", a.Name, "cwd", a.Cwd)` before
     `continue`.
   - [ ] `daemon/loop_test.go`: a helper that finds a record by `msg` and returns it, since
     `loggedAbout` keys on `bar`.

3. **More tests (RED → GREEN):**
   - [ ] `TestTick_HoldsAProjectThatHasALiveSession` (existing) keeps passing unchanged — it asserts
     no spawn and one surviving queue row, which is the hold/release rule this bar must not touch.
     Run it deliberately rather than assuming.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] With an interactive Claude session open in this repo and at least one approved bar queued,
  run `bp serve daemon -v`. Read the repeating line cold: it names that session and the directory
  it is sitting in, and says it is not dispatching. Close that session and the next tick dispatches.

## Commit (user)
`feat(daemon): name the session that holds a project`