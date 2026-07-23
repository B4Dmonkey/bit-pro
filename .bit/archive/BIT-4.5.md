---
id: BIT-4.5
title: Split the width between list and detail panes
status: done
phase: 2
phase_label: read a task
---
## Step 5 (Phase 2 — read a task) — Split the width between list and detail panes
**Status:** ✅ Done — verified 2026-07-18
> **Re-planned 2026-07-18** (was "Update toggles detail mode"). The scope pivoted from a
> full-screen modal to a persistent side-by-side preview pane, so the modal this step
> originally built is now unwanted. This step **removes** it and lays the width groundwork
> for the split instead. The old modal commit (`feat(tui): open and close the detail view`)
> stays in history; this step reverses that code.

The two-pane layout means the list can no longer own the full width — a detail pane sits
beside it. A pure `splitWidth(total) (listW, detailW)` is the testable core; `Update` sizes
the embedded list to `listW` on resize and stashes `detailW` for the renderer (Step 6).
Forced by a narrow-vs-normal contradiction: a split that just halves the width goes negative
at tiny terminal sizes, so the test at `total == 0` can't pass without a guard.

**Scope:**
- `tui/model.go` — add `splitWidth(total int) (int, int)` and a `detailWidth int` field;
  the `tea.WindowSizeMsg` case sizes the list to `listW` and stores `detailW`. **Delete** the
  `detail bool` field and the `enter`/`esc` detail branches in `Update` — the preview is
  always visible, so there is no mode to toggle and esc returns to pure list delegation.
- `tui/model_test.go` — add `TestSplitWidth`; **delete** `TestUpdate_EnterOpensDetail` and
  `TestUpdate_EscClosesDetail` (the modal they covered is gone). `TestUpdate_EscQuitsFromList`
  stays green — esc is once again plain delegation to the list's inherited quit.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestSplitWidth` (table)
     - **Behavior:** the total terminal width divides into a list column and a detail column
       that never overlap and never go negative, so both panes always fit — the list can't be
       handed the whole width or the detail pane vanishes.
     - **Setup:** totals of `0`, `1`, and a realistic `120`.
     - **Assertions:** for each, `listW >= 0`, `detailW >= 0`, and `listW + detailW <= total`
       (the remainder absorbs a one-column gap). At `120`, both `listW > 0` and `detailW > 0`,
       and `detailW > listW` — the 40/60 split makes detail the wider pane.
     - **Boundary:** total `0` (lower bound — both panes zero, no negative / no panic) vs a
       normal `120` (both panes positive, detail wider).
   - [x] Confirm fails: `splitWidth` undefined.

2. **Implement (GREEN):**
   - [x] `splitWidth` gives the list ~40% and the detail ~60% of `total` (list column is
     `total * 40 / 100`, detail is the remainder less a one-column gap), guarding tiny widths
     so neither goes negative.
   - [x] `Update` window-size case: `listW, detailW := splitWidth(msg.Width)`;
     `m.SetSize(listW, msg.Height)`; `m.detailWidth = detailW`.
   - [x] Remove the `detail bool` field and the enter/esc detail branches; delete the two
     modal tests.

**Claude verifies:**
- [x] `just test` green (modal tests gone; `TestUpdate_EscQuitsFromList` still passes)
- [x] `just lint` clean

**Commit (user):** `refactor(tui): replace modal detail with a pane width split`