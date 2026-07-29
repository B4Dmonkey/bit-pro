---
id: BIT-3.6
title: A phase survives a round trip
status: done
phase: 3
phase_label: A step shows its phase
---
## Step 6 (Phase 3 — A step shows its phase) — A phase survives a round trip

**Status:** ✅ Done — verified 2026-07-17

A bar records which slice of the track it serves, and `bit task read` shows it. Nothing
below this is hardcodable — the field either round-trips through YAML or it doesn't.

**Scope:**
- `task/task.go` — `Task`: add `Phase int` and `PhaseLabel string`
- `cmd/task_create.go` — `--phase` and `--phase-label` flags
- `cmd/task_read.go` — render the phase when present

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskReadCmd_ShowsPhase`
     - **Behavior:** you can see which phase a step serves without opening its scope —
       the indicator the TUI needs to review one bar at a time.
     - **Setup:** `initProject(t, "BIT")`; create BIT-1; then `task create "List cmd"
       -d "..." --parent BIT-1 --phase 2 --phase-label "List & read"`.
     - **Assertions:** `task read BIT-1.1` output contains `phase 2 — List & read`
       (exact rendering settled in implementation; assert against the whole first line).
     - **Boundary:** `phase` == 2, a non-zero value — 0 is the absent case and must not
       be confusable with it.
   - [x] Confirm fails: unknown flag `--phase`

2. **Implement (GREEN):**
   - [x] `Task`: `Phase int \`yaml:"phase,omitempty"\`` and
         `PhaseLabel string \`yaml:"phase_label,omitempty"\``. `omitempty` is load-bearing:
         a track has no phase and must not carry an empty one. `Task` stays comparable, so
         `TestStoreSaveLoad_RoundTrips`'s `*got != want` keeps working.
   - [x] Add `--phase` (int) and `--phase-label` (string) to `newTaskCreateCmd`
   - [x] `task_read.go`: when `t.Phase != 0`, print it on the header line

3. **More tests (RED → GREEN):**
   - [x] `TestTaskReadCmd_OmitsPhaseWhenAbsent`
     - **Behavior:** a track shows no phase noise — the approve/disapprove view stays clean.
     - **Setup:** create a plain task, no `--phase`.
     - **Assertions:** output equals the current `"BIT-1\ttodo\tTitle\n\nBody"` format
       exactly; the file on disk has no `phase:` key.
     - **Boundary:** `phase` == 0 — the absent/zero bound, and the reason `omitempty` is
       there rather than a pointer.

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**User verifies:**
- [x] The header rendering reads well — this is the line the TUI will inherit

**Commit (user):** `feat(task): record and show which phase a step serves`