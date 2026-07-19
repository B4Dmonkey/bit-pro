---
id: BIT-4.6
title: Render the list and detail panes side by side
status: done
phase: 2
phase_label: read a task
---
## Step 6 (Phase 2 — read a task) — Render the list and detail panes side by side
**Status:** ✅ Done — verified 2026-07-18 (committed `f4d4147`)
> **Re-planned 2026-07-18** (was: raw `JoinHorizontal` of the list and an unbounded
> Glamour render, "a viewport for overflow stays optional — YAGNI"). Manual testing
> showed that bet was wrong: a long body (`BIT-2`/`BIT-3` — the exact backtick-heavy
> imported content this phase exists to validate) rendered as one tall block, and
> `JoinHorizontal` padded the whole view to that height, overflowing the alt-screen and
> pushing the list off the top. The reference layout is also two *bounded* panes, which
> the raw join never produced. Both are fixed together: each pane is a bordered,
> height-capped box, and the detail is a `bubbles/viewport` that scrolls a long body
> inside its box instead of growing without bound.

The payload: the list and the selected task's rendered body, side by side, each in a
bounded box. `View` joins a bordered list pane and a bordered detail viewport horizontally,
both sized to the terminal height so neither overflows.

**Scope:**
- `go.mod` — no change; `viewport` ships in the already-required `charmbracelet/bubbles`
  module. `glamour` and `lipgloss` are already direct requires.
- `tui/model.go` — add a `viewport.Model` (bordered `Style`; scroll keymap on
  `ctrl+u`/`ctrl+d`, keys the list's default keymap ignores) plus `listWidth`/`height`
  fields. `WindowSizeMsg` sizes the list and viewport to the border-inset dimensions and
  (re)builds the Glamour renderer; `refreshDetail` renders `selected().Body` into the
  viewport and resets it to the top; the default `Update` path forwards messages to both
  the list and the viewport, re-rendering the detail only when the selection moved. `View`
  wraps the list in a bordered box and joins it to the viewport's bordered view.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestView_FitsWindowHeight` — 500-line body, `WindowSizeMsg{80,24}`, assert
     `lipgloss.Height(View()) <= 24`. The direct regression for the overflow bug.
     *Boundary:* body far taller than the window.
   - [x] `TestUpdate_WindowSizeSizesViewport` — the detail pane is height-bounded
     (`viewport.Height == 24`), not just built.
   - [x] `TestUpdate_CtrlDScrollsDetail` — `ctrl+d` scrolls the detail (`YOffset > 0`),
     proving the non-colliding scroll key is wired. *Boundary:* top (0) → scrolled.
   - [x] `TestUpdate_NavigationResetsDetailScroll` — changing selection resets the detail
     to the top (`YOffset == 0`). *Boundary:* selection change → `GotoTop`.
   - [x] Confirm fails: `model` has no `viewport` field (compile error).

2. **Implement (GREEN):**
   - [x] Bordered `viewport` + list box, border-inset sizing on `WindowSizeMsg`,
     `refreshDetail` with raw-body fallback, dual message forwarding. Deliberately out of
     scope (matches the reference *frame*, not every widget): the Filters bar, per-field
     detail formatting, box titles (no lipgloss v1.1 API), mouse-wheel scroll.

**Claude verifies:**
- [x] `just build`, `just test` green, `just lint` clean

**User verifies:**
- [x] `bit tui` shows two **bounded** (bordered) panes side by side — the list left, the
  selected task's rendered body right — neither overflowing the screen
- [x] moving the cursor (arrows / j-k) updates the right pane live and starts it at the top
- [x] a long, backtick-heavy task (`BIT-3.9`, or a whole track like `BIT-2`/`BIT-3`) stays
  inside its box and scrolls with `ctrl+u`/`ctrl+d` — no longer takes over the screen
- [x] markdown reads well in the narrow pane — headings, code fences, lists intact
- [x] `q`/`esc`/`ctrl+c` still quit cleanly

**Commit (user):** `feat(tui): render the list and detail panes side by side`