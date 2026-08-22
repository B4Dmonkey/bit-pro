---
id: BIT-33.5
title: clear-queue.sh empties the queue table
status: done
approved: true
phase: 5
phase_label: Operator can clear the queue
---
## **Verse 5**

Adds the throwaway operator script that empties the `queue` table, so the operator can reset
between tracks while nothing in this scope dequeues.

## Scope
- `clear-queue.sh` — new, repo root, beside `install.sh`. Deliberately minimal; **deleted, not
  migrated**, when the dispatch track lands a real dequeue surface.

## No test — deliberate
The operator running it is the check. Recorded here so this bar does not get "helpfully" given a
test file: `update/normalize.sh` has a companion `update/normalize_test.sh`, this script does not.
It is six lines with no branching, and its entire surface is the one command under **User verifies**.

## Implement
- [ ] `#!/usr/bin/env bash` and `set -euo pipefail`.
- [ ] Resolve the **runtime** database the way `store.Dir()` does (`store/store.go:63` — it reads
      `XDG_DATA_HOME` before falling back to `~/.local/share`):
      `db="${XDG_DATA_HOME:-$HOME/.local/share}/bit-pro/bit.db"`.
      Not `db/bit.db` — that is the dbmate/sqlc throwaway the Justfile exports as `DATABASE_URL`.
- [ ] `n="$(sqlite3 "$db" 'DELETE FROM queue; SELECT changes();')"` then
      `echo "cleared $n queue rows"`. Verified against sqlite3 3.51.0: prints the count, and `0`
      on a second run rather than nothing.
- [ ] No arguments, no `--project` filter, no missing-table handling. A missing database or a
      missing `queue` table surfaces sqlite3's own `Error: in prepare, no such table: queue` and
      `set -e` propagates exit 1 — verified, including that `set -e` does fire through the
      command substitution in the assignment.
- [ ] `chmod +x clear-queue.sh`.

Known wart, accepted rather than handled: sqlite3 creates the database file before failing, so a
wrong `XDG_DATA_HOME` leaves a stray empty `bit.db` behind. Verified. Left alone to keep the
script throwaway.

## Claude verifies
- [ ] `shellcheck clear-queue.sh` reports no errors — the same check the BIT-21 bars ran on
      `normalize.sh`; shellcheck is installed at `/opt/homebrew/bin/shellcheck`.
- [ ] `test -x clear-queue.sh`.

## User verifies
- [ ] In `bp tui`, press `e` on a track and on a bar (Verse 3) to queue two items. Run
      `./clear-queue.sh` — it prints `cleared 2 queue rows`.
- [ ] Run it again immediately — it prints `cleared 0 queue rows`.
- [ ] Whole slice: with the TUI still open, the cyan on those two rows clears on the next reload
      cycle. The queue is now resettable between tracks without hand-writing SQL — the capability
      Verse 5 exists to deliver.

## Commit (user)
`chore(queue): temporary clear-queue.sh for resetting between tracks`