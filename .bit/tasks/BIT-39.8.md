---
id: BIT-39.8
title: A live session holds the project
status: todo
phase: 2
phase_label: Bar runs unattended
---
## **Verse 2**

One bar in flight per project: a project with any live Claude session is skipped before its queue
is even read. Contradicts every bar so far — same single queued row, but a live session in the
project means no `--bg` call and no delete. Without this the next tick stampedes the rest of the
queue.

## Scope
- `claude/dispatch.go` — `func (a Agent) Under(root string) bool`, the at-or-beneath path test.
- `claude/dispatch_test.go` — the path-matching table.
- `daemon/loop.go` — one `Agents` snapshot at the top of `Tick`, consulted per project before the
  queue is read; the post-spawn confirm keeps its own fresh call, since a session spawned this tick
  cannot be in a snapshot taken before it.
- `daemon/loop_test.go` — the RED test.

## References
- `automation-notes.md` — "Measured 2026-08-25": `claude agents --json` is machine-wide, so the
  filter is the loop's own job; `--cwd` is not used, because whether it reaches into
  `.claude/worktrees/` is unmeasured and the at-or-beneath rule settles it without depending on
  that.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAgent_Under` in `claude/dispatch_test.go`, table-driven
     - **Behavior:** a session belongs to a project when its directory is at or beneath the
       registered project path — which is the one rule covering all three shapes a real row takes:
       the checkout itself, a dispatched session in `<repo>/.claude/worktrees/<name>`, and an
       interactive session started in a subdirectory.
     - **Setup:** `root = "/p/foo"`, cases `{cwd, want}`:
       `{"/p/foo", true}`; `{"/p/foo/.claude/worktrees/BIT-1-a-track", true}`;
       `{"/p/foo/cmd", true}`; `{"/p/foobar", false}`; `{"/p", false}`; `{"/q/foo", false}`;
       `{"/p/foo/", true}`.
     - **Assertions:** `Agent{Cwd: cwd}.Under(root) == want`.
     - **Boundary:** `"/p/foobar"` is the adversarial case a plain `strings.HasPrefix` gets wrong —
       one character past the root with no separator, the boundary between "beneath" and "merely
       shares a prefix". `"/p"` is the other side: an ancestor, not a descendant.
   - [ ] `TestTick_HoldsAProjectThatHasALiveSession` in `daemon/loop_test.go`
     - **Behavior:** the loop dispatches at most one bar per project at a time, so a track's bars
       run in order instead of all at once.
     - **Setup:** the `TestTick_DispatchesTheQueuedBar` fixture. The fake runner answers
       `agents --json` with a one-row array whose `cwd` is `filepath.Join(dir, "cmd")` — a live
       session in a subdirectory of the project, with a `name` that matches no dispatch name, so
       only the cwd rule can make this pass.
     - **Assertions:** no recorded call has `args[0] == "--bg"`, and `ListQueueByProject` still
       returns 1 row — held, not dropped.
     - **Boundary:** live sessions under this project == 1 — the lower non-empty bound; every
       earlier bar ran this same fixture at 0.
   - [ ] Confirm fails: a `--bg` call is recorded — the loop dispatches on top of a live session.

2. **Implement (GREEN):**
   - [ ] In `claude/dispatch.go`: `func (a Agent) Under(root string) bool` — `filepath.Clean` both,
     return `cwd == root || strings.HasPrefix(cwd, root+string(filepath.Separator))`. The explicit
     separator is what excludes `/p/foobar`.
   - [ ] In `daemon/loop.go`, at the top of `Tick`, take one `claude.Agents` snapshot; on error log
     and skip dispatch entirely for this tick (the counts write still runs). Per project, before
     `ListQueueByProject`: if any snapshot row satisfies `Under(p.Path)`, `log.Debug` and continue.
     Presence in the snapshot is the whole liveness test — no `state` field is read.
   - [ ] Check the earlier Tick tests still pass: they now see two `agents --json` calls per tick,
     and the payloads they return were chosen so the guard does not fire.

## Claude verifies
- [ ] `just test` — every Verse 2 test together, since this bar changes what the shared fixture's
  runner is asked
- [ ] `just lint`

## User verifies
- [ ] Whole slice: `just install`. Pick a real approved, not-done bar on this repo and enqueue it
  from `bp tui` (select it, press the enqueue key), confirm `bp list` shows the project, then run
  `bp serve daemon -v` in a terminal with **no other Claude session open in bit-pro** and walk
  away. Within a couple of ticks `claude agents --json` shows a background row named
  `BIT-39-dispatch-the-daemon-works-queued-bars-unattended`, and the queue row is gone. Come back:
  the bar's own work is committed on `worktree-BIT-39-dispatch-the-daemon-works-queued-bars-unattended`.
  A bar ran with nobody watching, which is the whole point of the verse.

## Commit (user)
`feat(daemon): hold a project that already has a live session`