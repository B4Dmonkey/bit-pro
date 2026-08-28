---
id: BIT-39.16
title: Agents and Spawn run the binary they are given
status: done
approved: true
phase: 4
phase_label: Works with the terminal closed
---
## **Verse 4**

`claude.Agents` and `claude.Spawn` hardcode the string `"claude"` as the command they run, so
there is nowhere for the pinned path from the previous bar to go. They take the binary instead,
and the daemon threads it down from its caller. `cmd/serve.go` passes the literal `"claude"` here,
which keeps behaviour identical — the next bar is what replaces that literal with `BP_CLAUDE`.

This bar is forced by a test asserting a non-`"claude"` binary reaches `exec`, which the current
hardcode cannot satisfy.

## Scope
- `claude/dispatch.go` — `Agents(ctx, run, bin)` and `Spawn(ctx, run, bin, dir, name, bar)`.
- `daemon/loop.go` — `Loop`, `Tick` and `dispatch` carry the binary through to both calls.
- `cmd/serve.go` — passes the literal `"claude"` to `daemon.Loop`.
- `claude/dispatch_test.go`, `daemon/loop_test.go` — the existing call sites gain the argument.

The whole chain moves in one bar deliberately: `Agents` and `Spawn` are called directly from
`daemon`, not through an interface, so widening them alone would leave the module uncompilable.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAgents_RunsTheBinaryItIsGiven`
     - **Behavior:** the listing shells out to the binary it is handed, not to a name it assumes
       is on the `PATH`.
     - **Setup:** the existing `dispatch_test.go` recording runner (captures `dir`, `name`,
       `args` into a `call` and returns `"[]", 0, nil`). Call
       `Agents(t.Context(), run, "/opt/homebrew/bin/claude")`.
     - **Assertions:** `got.name == "/opt/homebrew/bin/claude"`; `got.args` still equals
       `[]string{"agents", "--json"}`; `got.dir == ""`.
     - **Boundary:** the binary is an absolute path rather than the bare name — the only shape
       launchd can actually execute, and the value that contradicts the hardcoded `"claude"`.
   - [ ] Confirm fails: `Agents` takes two arguments, so the test does not compile; once widened
     but still hardcoded, `got.name` is `"claude"`.

2. **Implement (GREEN):**
   - [ ] `claude/dispatch.go`: add a `bin string` parameter to `Agents` and to `Spawn`, and use it
     as the first argument to `run` in both.
   - [ ] `daemon/loop.go`: add `bin string` to `Loop`, `Tick` and `dispatch`, passing it to both
     `claude.Agents` and `claude.Spawn`.
   - [ ] `cmd/serve.go`: pass `"claude"` to `daemon.Loop`.
   - [ ] `daemon/loop_test.go`: pass `"claude"` at every `Tick(...)` and `Loop(...)` call site.

3. **More tests (RED → GREEN):**
   - [ ] `TestTick_SpawnsWithTheBinaryItIsGiven`
     - **Behavior:** the spawn path carries the binary too, not just the listing — otherwise the
       daemon could see its sessions and still fail to start one.
     - **Setup:** `queries, project := queuedBar(t)`; `calls, run := recordingRunner()`;
       `Tick(t.Context(), queries, log, run, "/opt/homebrew/bin/claude")`.
     - **Assertions:** the recorded call whose `args[0] == "--bg"` has
       `name == "/opt/homebrew/bin/claude"`, and its `args` are unchanged from the values
       `TestTick_DispatchesTheQueuedBar` already pins.
     - **Boundary:** the `--bg` call specifically — the second of the two places the binary is
       used, and the one that is not exercised by the `Agents` test above.
   - [ ] Confirm fails: `Spawn` still names `"claude"` regardless of what `Tick` was given.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic; `bp serve daemon -v` behaves exactly as before, which is the point.

## Commit (user)
`refactor(claude): run the binary given instead of a hardcoded name`