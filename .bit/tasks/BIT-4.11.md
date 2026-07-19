---
id: BIT-4.11
title: Derive bar / verse from the ID and phase
status: done
phase: 3
phase_label: see the shape
---
## Step 11 (Phase 3 — see the shape) — Derive bar / verse from the ID and phase
**Status:** ✅ Done — verified 2026-07-18
The pure helpers behind the visual grouping: is an ID a bar (dotted), and what verse string
does it show. Forced by a track-vs-bar contradiction. `idParts`/`compareIDs` are unexported
to `task`, so `tui` derives this itself — a `strings.Contains(id, ".")` check plus the
already-exported `Phase`/`PhaseLabel`; no new `task` API.

**Scope:**
- `tui/model.go` — `isBar(id string) bool`; `verse(t *task.Task) string`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestIsBar` (table)
     - **Behavior:** a dotted ID is a bar (a step under a track); a plain ID is a track — so
       the delegate can indent and label correctly.
     - **Setup:** IDs `"BIT-2"` and `"BIT-2.5"`.
     - **Assertions:** `isBar("BIT-2") == false`, `isBar("BIT-2.5") == true`.
     - **Boundary:** presence/absence of the dot — the two sides of the classification.
   - [x] Confirm fails: `isBar` undefined.
   - [x] `TestVerse` (table)
     - **Behavior:** a bar shows the verse (phase) it serves; a track and an unphased bar
       show nothing — the indicator only appears where it means something.
     - **Setup:** a bar with `Phase: 2, PhaseLabel: "List & read"`; a bar with `Phase: 0`; a
       phased track (`Phase: 2`, to prove a track shows no verse even when phased).
     - **Assertions:** `"phase 2 — List & read"`; `""`; `""` respectively.
     - **Boundary:** `Phase == 0` (lower bound, no verse) vs `Phase > 0` (verse present);
       the phased-track case exercises the `!isBar` branch, which `Phase == 0` alone can't.

2. **Implement (GREEN):**
   - [x] `isBar` = dotted-ID check; `verse` formats `Phase`/`PhaseLabel`, empty when
     `Phase == 0`.

**Claude verifies:**
- [x] `just test` green
- [x] `just lint` clean

**Commit (user):** `feat(tui): derive bar and verse from task id and phase`