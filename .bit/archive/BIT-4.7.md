---
id: BIT-4.7
title: '`→`/`←` move focus between the panes'
status: done
phase: 4
phase_label: focus & affordances
---
## Step 7 (Phase 4 — focus & affordances) — `→`/`←` move focus between the panes
**Status:** ✅ Done — verified 2026-07-18
The focus state the rest of Phase 4 hangs on. Today `←`/`→` fall through to the list and
page it; this step intercepts them to move focus between the two panes instead, defaulting
to the list. Forced by two contradicting transitions — one handler can't both leave focus
alone and move it.

**Scope:**
- `tui/model.go` — a `detailFocused bool` field (default `false` = list focused) set in
  `New`; `Update` intercepts `KeyLeft`/`KeyRight` to set it (clamped at each end) and returns
  before the message reaches the list, so the list no longer pages on those keys.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestUpdate_Focus` (table)
     - **Behavior:** `→` moves focus to the detail pane, `←` back to the list; focus is the
       state that decides which pane the arrows drive.
     - **Setup:** `New` with one task, `WindowSizeMsg{80,24}`; drive `KeyRight`, then `KeyLeft`.
     - **Assertions:** default `!m.detailFocused`; after `KeyRight`, `m.detailFocused`; after a
       further `KeyLeft`, `!m.detailFocused`.
     - **Boundary:** both ends clamp — `KeyLeft` while already on the list stays list-focused;
       `KeyRight` while already on the detail stays detail-focused.
   - [x] `TestUpdate_RightDoesNotPageList`
     - **Behavior:** `→` no longer pages the list — it only moves focus.
     - **Setup:** `New` with 3 tasks, `WindowSizeMsg{80,24}`, `KeyRight`.
     - **Assertions:** `Index() == 0` (unchanged) and `Paginator.Page == 0`.
     - **Boundary:** the key that previously paged now drives focus only.
   - [x] Confirm fails: `model` has no `detailFocused` field (compile error).

2. **Implement (GREEN):**
   - [x] Add `detailFocused`; handle `KeyLeft`/`KeyRight` in `Update` (clamped), returning
     before the list sees them.

**Claude verifies:**
- [x] `just test` green, `just lint` clean

**User verifies:**
- [x] `→`/`←` move focus between the two panes and no longer page the list

**Commit (user):** `feat(tui): move focus between the list and detail panes`

Note: `ctrl+f`/`ctrl+b` for list paging (your "we could consider") is deliberately left out
here — YAGNI until you want it; easy to add as its own step.