---
id: BIT-4.8
title: Arrows drive the focused pane
status: done
phase: 4
phase_label: focus & affordances
---
## Step 8 (Phase 4 — focus & affordances) — Arrows drive the focused pane
**Status:** ✅ Done — verified 2026-07-18
The payload of focus: `↑`/`↓`/`j`/`k` move the list selection when the list is focused and
scroll the detail when the detail is focused. Forced by a contradiction — the same `↓` must
move the list in one state and leave it still (scrolling the body) in the other. This
supersedes Step 6's `ctrl+u`/`ctrl+d` workaround: with routing, the viewport carries its own
default keymap (`↑↓ jk`) and never collides with the list.

**Scope:**
- `tui/model.go` — the default `Update` path forwards each message to *only* the focused
  component (list when `!detailFocused`, viewport when `detailFocused`) instead of both;
  restore `viewport.DefaultKeyMap()` in `New` (drop the `ctrl+u`/`ctrl+d`-only map).
  `refreshDetail` on selection change is now reached only when the list is focused.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestUpdate_FocusRoutesArrows` (table)
     - **Behavior:** arrows act on the focused pane only — the unfocused pane holds still.
     - **Setup:** two long-bodied tasks, `WindowSizeMsg{80,24}`. Case A: default (list),
       `KeyDown`. Case B: `KeyRight` then `KeyDown` (detail focused).
     - **Assertions:** A — `Index() == 1` and `viewport.YOffset == 0` (list moved, detail
       still). B — `Index() == 0` and `viewport.YOffset > 0` (detail scrolled, list still).
     - **Boundary:** the two focus states are the two branches; each proves the *other* pane
       is untouched.
   - [x] Confirm fails: with forward-to-both, case B moves the list too — `Index() == 0` fails.

2. **Implement (GREEN):**
   - [x] Route the message to the focused component only; restore the viewport default keymap.

**Claude verifies:**
- [x] `just test` green — `TestUpdate_CtrlDScrollsDetail` **and** `TestUpdate_NavigationResetsDetailScroll`
  are both revised to focus the detail before scrolling it (routing means `ctrl+d` only
  scrolls the focused pane, and the list binds bare `d`/`u` but not `ctrl+d`/`ctrl+u`, so the
  plan's earlier "stays green unchanged" note for the latter was wrong) — and `just lint` clean

**User verifies:**
- [x] with the detail focused, `↑`/`↓` scroll a long body; with the list focused, they move
  the selection — neither disturbs the other

**Commit (user):** `feat(tui): route arrow keys to the focused pane`