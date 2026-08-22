---
id: BIT-33
title: Queue
status: doing
---
## Why
The serve loop (`bp serve`) dispatches bars unattended, but it has no input: nothing tells
it what to work on or in what order. The queue is that input — a persistent, ordered list
of tracks and bars an operator has cleared for automation. Without it, approving work is
the end of the operator's involvement rather than the beginning of the machine's.

## Summary
Add a `queue` table to `bit.db` with a FK to `projects.id`. Enqueue from two surfaces: the
play-prompt popup ("yes") and a TUI shortcut for tracks or bars the operator wants to queue
without the popup. Queued items render cyan; the color clears when the item leaves the queue. Nothing in this
scope consumes the queue, so a throwaway top-level `clear-queue.sh` empties the table between
runs.

## Visual aid
```
Operator approves last bar → play prompt opens
  y  →  queue.add(project_id, subject="BIT-7", kind=track)  →  cyan in TUI

TUI on any approved track or bar  →  press e  →  same queue.add  →  same cyan

bp serve pops the head  →  row removed  →  cyan clears
```

## Decisions

- **Queue rows are dual-typed: track or bar.** A track row means "work its bars in order";
  a bar row means "work exactly this bar." The table stores `subject_id TEXT` (the BIT-N or
  BIT-N.M string) and `subject_kind TEXT` (track | bar).
- **FK to `projects.id`.** The queue is global state in `bit.db`; the project registry is
  the stable FK anchor. BIT-29 ensured `id` is the primary key for exactly this purpose.
- **FIFO within a project, by insertion order.** The queue table carries an auto-increment
  `id`; the serve loop always pops the row with the smallest `id` for a given `project_id`.
  Track and bar rows are sequenced together — there is no separate track queue.
- **Enqueue shortcut is `e`.** `q` conflicts with quit.
- **Unregistered project: silently do nothing.** If the TUI tries to enqueue but the current
  project has no `bp add` row, the enqueue is a no-op. No error surface in the TUI for now;
  noted as an open gap in automation-notes.md.
- **Queued rows render cyan.** The delegate's color logic gains a cyan path when a task
  appears in the queue. Cyan outranks yellow (unapproved) and yields to green (selected).
- **Cyan clears when the row leaves the queue.** The reload loop is extended to also pull
  queue state each cycle; the color is derived from DB state, not client-side memory.
- **Both enqueue paths produce the same row shape.** The popup and the shortcut call the
  same queue write function; the row is identical regardless of how it was requested.
- **No CLI queue commands in this step.** `bp queue add/list/rm` are out of scope.
  The queue is written by the TUI and popped by the serve loop; the operator escape hatch
  is `clear-queue.sh` (Verse 5). Step 6 (dispatch) is where a dequeue command earns its keep.
- **The escape hatch is a throwaway shell script, not a `bp` subcommand.** Nothing here dequeues,
  so rows accumulate with no way to remove them. `clear-queue.sh` at the repo root wipes the
  table so the operator can reset between tracks. Deliberately not Go: a `bp queue clear` would
  have to be designed, tested, and then unbuilt once dispatch owns dequeuing. It sits beside
  `install.sh` and is **deleted, not migrated**, when the dispatch track lands a real dequeue
  surface.
- **The script targets the runtime database, not the dev one.** Two `bit.db` files exist: the
  Justfile exports `DATABASE_URL` at `db/bit.db` for dbmate and sqlc, while the daemon and TUI
  read `~/.local/share/bit-pro/bit.db`. The script uses the runtime path.
- **It clears every row and takes no arguments.** One project is registered, so a `--project`
  filter would be speculative. Delete all, report the count removed.
- **Daemon state is not the TUI's concern for this MVP.** Answering "yes" while the daemon
  is stopped writes a queue row that nothing dispatches until the daemon starts. The TUI
  does not check or surface daemon state. The operator is expected to know whether `bp serve`
  is running.

## Verses

- [x] Verse 1 — Queue table exists: `queue` table migration and sqlc-generated queries land.
  No CLI surface; verifiable by inspecting `bit.db` directly.
  Touches: `db/migrations/` (new migration), `db/queries/queue.sql`, `db/orm/` (sqlc
  regenerated).

- [ ] Verse 2 — Popup "yes" enqueues: answering "yes" at the play prompt creates a `queue`
  row for the track. Verifiable by checking `bit.db`.
  Touches: `tui/board.go` (`handlePlayPrompt`), wired to the queue store from Verse 1.

- [ ] Verse 3 — TUI shortcut enqueues: pressing `e` on any track or bar creates a queue
  row without the popup. Works in both list and board modes.
  Touches: `tui/model.go`, `tui/board.go` key dispatch.

- [ ] Verse 4 — Queued items render cyan: the delegate colors queued tasks cyan; the reload
  loop is extended to pull queue state each cycle so the color clears when a row is removed
  from the queue (hand-edited `bit.db` at this point; `clear-queue.sh` from Verse 5, or dispatch,
  later).
  Touches: `tui/delegate.go`, `tui/model.go` reload loop.

- [ ] Verse 5 — Operator can clear the queue: `clear-queue.sh` at the repo root empties the
  `queue` table and reports how many rows it removed, so the operator can reset between tracks.
  Temporary — deleted once dispatch owns dequeuing.
  Touches: `clear-queue.sh` (new, repo root).