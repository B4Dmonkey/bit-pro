---
id: BIT-39.5
title: A confirmed session dequeues its row
status: done
approved: true
phase: 2
phase_label: Bar runs unattended
---
## **Verse 2**

A spawn is only finished once the session is visible: `Tick` polls `claude agents --json` and
deletes the queue row when a row under the dispatched name is there. Forced by a test that
asserts the row is gone — the previous bars spawn and leave the queue untouched, so the same bar
would be re-dispatched on every tick.

## Scope
- `db/queries/queue.sql` — add `DeleteQueueRow`; `just db-gen-queries` regenerates
  `db/orm/queue.sql.go`.
- `claude/dispatch.go` — `Agent` (`name`, `cwd`) and `Agents(ctx, run)`, which runs
  `claude agents --json` and unmarshals the array.
- `claude/dispatch_test.go` — a parse test for `Agents` against a real captured payload.
- `daemon/loop.go` — after `Spawn`, call `Agents`, look for the dispatched name, delete the row.
- `daemon/loop_test.go` — the RED test.

## References
- `automation-notes.md` — "Measured 2026-08-25" and the 2026-08-21 entries: `claude agents --json`
  is machine-wide, needs no TTY, and is the only poll surface a launchd-hosted daemon has.

## Needs real data
The `Agents` parse test must run against a payload this machine actually emits, not a hand-written
one — the row shape differs between `kind: interactive` and `kind: background`.

- [ ] `claude agents --json > testdata/agents.json` from the repo root while at least one
  background and one interactive session are live, then trim it to two or three rows and strip the
  absolute paths down to the fixture's temp dirs. A payload captured 2026-08-25 on 2.1.245 had
  every row carrying `pid`, `cwd`, `kind`, `startedAt`, `sessionId`, `name`, `status`, with
  background rows additionally carrying `id` and `state`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAgents_ParsesTheRealPayload` in `claude/dispatch_test.go`
     - **Behavior:** `Agents` reads the `claude agents --json` array into `name`/`cwd` pairs,
       ignoring every other field, so the loop can ask "is anything running here" without tracking
       the CLI's full row shape.
     - **Setup:** a fake `DirRunner` returning the captured `testdata/agents.json` bytes, code 0.
     - **Assertions:** one `Agent` per row in the payload, in payload order, with `Name` and `Cwd`
       matching the file. The runner was called as
       `(ctx, "", "claude", "agents", "--json")`.
     - **Boundary:** the payload mixes rows **with** and **without** the `state` and `id` keys —
       the absent-field case, which must decode rather than error.
   - [ ] `TestTick_DequeuesAConfirmedDispatch` in `daemon/loop_test.go`
     - **Behavior:** once the spawned session is visible under its dispatched name, the row is
       gone — the bar is in flight and must not be dispatched twice.
     - **Setup:** the `TestTick_DispatchesTheQueuedBar` fixture. The fake runner answers the
       `agents --json` call with a one-row array whose `name` is `"ACME-1-a-track"` and whose `cwd`
       is a directory **outside** the project (so the not-yet-written in-flight guard cannot be what
       makes this pass), and returns `("", 0, nil)` for the `--bg` call.
     - **Assertions:** after `Tick`, `queries.ListQueueByProject(ctx, id)` returns zero rows.
     - **Boundary:** rows remaining == 0 — the lower bound of the queue after a successful
       dispatch, and the only value that distinguishes delete-on-confirm from no delete at all.
   - [ ] Confirm fails: `ListQueueByProject` still returns 1 row. (`undefined:
     queries.DeleteQueueRow` comes first — add the query, then this is the real RED.)

2. **Implement (GREEN):**
   - [ ] Add to `db/queries/queue.sql`:
     `-- name: DeleteQueueRow :exec` / `DELETE FROM queue WHERE id = ?;` and run
     `just db-gen-queries`.
   - [ ] In `claude/dispatch.go`: `type Agent struct { Name string \`json:"name"\`; Cwd string \`json:"cwd"\` }`
     and `func Agents(ctx context.Context, run DirRunner) ([]Agent, error)` — call
     `run(ctx, "", "claude", "agents", "--json")`, error on non-zero code, `json.Unmarshal` into
     `[]Agent`.
   - [ ] In `daemon/loop.go`, after a successful `Spawn`: call `claude.Agents`, and if any returned
     `Agent.Name == name`, `queries.DeleteQueueRow(ctx, row.ID)`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. The runner is faked and no real session is spawned.

## Commit (user)
`feat(daemon): dequeue a bar once its session is confirmed`