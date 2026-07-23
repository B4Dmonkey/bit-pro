---
id: BIT-4.9
title: Titled, focus-accented borders
status: done
phase: 4
phase_label: focus & affordances
---
## Step 9 (Phase 4 — focus & affordances) — Titled, focus-accented borders
**Status:** ✅ Done — verified 2026-07-18
Each pane's top border carries a title — the list its live task count (`Tasks (29)`), the
detail `Details` — and the focused pane's border is accented so the active pane is obvious.
The pinned lipgloss has no border-title API (verified against the module source), so the
title is composited into the top border row by hand; the same helper applies the accent. The
title *text* is assertable; the exact look (alignment, colour) is manual.

**Scope:**
- `tui/model.go` — a `titledBorder(content, title string, width, height int, active bool) string`
  helper that renders the bordered box and overlays `title` into its top border row, using an
  accent border colour when `active`. `View` builds both panes through it: list title
  `fmt.Sprintf("Tasks (%d)", len(m.Items()))`, detail title `"Details"`, each `active` per
  `detailFocused`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestView_PaneTitles` (table)
     - **Behavior:** the borders announce what each pane is and how many tasks there are, so
       the frame is legible without a legend.
     - **Setup:** `New` with 29 tasks (and a second case with 1), `WindowSizeMsg{80,24}`.
     - **Assertions:** `strings.Contains(View(), "Tasks (29)")` and `strings.Contains(View(), "Details")`;
       the 1-task case contains `"Tasks (1)"`.
     - **Boundary:** count reflects `len(Items())` at N=1 (singular) and N=29.
   - [x] Confirm fails: `View` output has no title text.

2. **Implement (GREEN):**
   - [x] `titledBorder` compositing the title into a hand-built top border row (`BorderTop(false)`
     box + manual title row stacked with `JoinVertical`); accent `BorderForeground`/`Foreground`
     (`lipgloss.Color("99")`) when `active`; `View` routes both panes through it. The viewport's
     own border was dropped and it is now sized to the border-inset dims (`detailW-2`,
     `msg.Height-2`), which is why `TestUpdate_WindowSizeSizesViewport` was revised from
     `== 24` to `== 22` (the border moved off the viewport onto `titledBorder`).

**Claude verifies:**
- [x] `just test` green (`TestView_FitsWindowHeight` still ≤ 24), `just lint` clean

**User verifies:**
- [x] both panes show titled borders (`Tasks (N)`, `Details`); the focused pane is visibly
  accented and the accent moves with `→`/`←`
- [x] the accent colour reads well — accepted for now; the exact "flavor" is deferred to a
  future looks-refinement scope (`99` purple is a placeholder)

**Commit (user):** `feat(tui): title the pane borders and accent the focused one`