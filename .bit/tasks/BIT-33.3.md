---
id: BIT-33.3
title: e key calls enqueue stub for selected task
status: todo
approved: true
phase: 3
phase_label: TUI shortcut enqueues
---
## **Verse 3**

TUI test: pressing `e` calls no enqueue stub → fails (no `e` handler) → add handler in both list and board modes. A contradicting test uses a bar ID to prove the handler derives `kind` dynamically rather than hardcoding `"track"`.

## Scope
- `tui/model.go` — add `case "e":` in the Update switch (list mode path); call `m.enqueue(selectedID, kind)` where `kind` = `"bar"` if `isBar(selectedID)` else `"track"`; guard `m.enqueue != nil`
- `tui/board.go` — add `case "e":` in `updateBoard`'s key dispatch (board mode); same call pattern

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_EKey_EnqueuesTrack`
     - **Behavior:** pressing `e` in list mode enqueues the selected track
     - **Setup:** model with tasks `[{ID: "BIT-33", ...}]` selected, `WithEnqueue` stub
     - **Assertions:** stub called with `("BIT-33", "track")`
     - **Boundary:** `"e"` key, track ID (no dot) — proves the `"track"` kind path
   - [ ] Confirm fails: `"e"` key falls through unhandled (no enqueue call)

2. **Implement (GREEN):**
   - [ ] Add `case "e":` to list mode key dispatch in `tui/model.go`
   - [ ] Add `case "e":` to board mode key dispatch in `tui/board.go`

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_EKey_EnqueuesBar`
     - **Behavior:** pressing `e` with a bar selected passes `"bar"` as the kind
     - **Setup:** same but selected task ID `"BIT-33.1"` (contains a dot)
     - **Assertions:** stub called with `("BIT-33.1", "bar")`
     - **Boundary:** bar ID (contains `.`) — contradicts any hardcoded `"track"` kind

## Claude verifies
- [ ] `go test ./tui/...` passes
- [ ] `go build ./...` passes
- [ ] `just install`

## User verifies
- [ ] In `bp tui`, navigate to any track and press `e`; in board mode navigate to any bar and press `e`; after each, run `sqlite3 ~/.local/share/bit-pro/bit.db "SELECT * FROM queue;"` and confirm a row was added with the correct `target_id` and `target_typ`

## Commit (user)
`feat(tui): e key enqueues selected task`