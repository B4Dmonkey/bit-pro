---
id: BIT-1.3
title: '`just build` produces a runnable binary'
status: done
phase: 1
phase_label: Bootstrap
---
## Step 3 (Phase 1 — Bootstrap) — `just build` produces a runnable binary
**Status:** ✅ Done — verified 2026-07-15
No new Go behavior to test — Steps 1–2 already proved `--help` and `--version` work at
the Cobra-command level. This step wraps that proven behavior in the `just` task runner
the scope calls for, and closes the loop by running the *actual built binary*, not just
the in-process command — that's the real acceptance bar the scope's Phase 1 sets
("`just build` produces a binary, and `bit --help` / `bit --version` work").

**Scope:**
- `Justfile` — new, `build`, `run`, `test` recipes

**Implementation:**
- [ ] `Justfile`:
  ```just
  build:
      go build -o bin/bit .

  run *ARGS:
      go run . {{ARGS}}

  test:
      go test ./...
  ```

**Claude verifies:**
- [x] `just test` passes (proves the `test` recipe correctly wraps `go test ./...`)
- [x] `just build` succeeds and produces `bin/bit`
- [x] `./bin/bit --help` output contains `"bit"` — the `"Usage:"` part of this assertion
  was wrong: it contradicts Step 1's own documented finding that Cobra only emits
  `Usage:` once the command is runnable or has a subcommand, neither of which is true
  until Step 4. Noted in the README under "Known limitations" as a future cleanup item
  instead of forcing it now.
- [x] `./bin/bit --version` outputs exactly `bit version 0.1.0-dev`

**User verifies:**
- [x] `just` as the task runner feels right for day-to-day use (build/run/test)

**Commit (user):** `feat(bootstrap): add Justfile task runner`