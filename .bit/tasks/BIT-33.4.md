---
id: BIT-33.4
title: Contradiction forces cyan color + reload pulls queue
status: todo
phase: 4
phase_label: Queued items render cyan
---
## **Verse 4**

Delegate test: queued task with existing approved=true style resolves to yellow now → add `queuedIDs` field + cyan path in `resolveStyle`; a contradicting test proves selected state beats cyan. Reload test: tick fires → `listQueue` stub called → `m.queuedIDs` populated.

## Scope
- `tui/model.go` — add `queuedIDs map[string]bool` field; `WithListQueue(fn func() ([]string, error)) Option`; in `reloadCmd`: also call `listQueue()` and return both task list and queue IDs in a combined msg (or extend `reloadedMsg`); in `handleReloaded`: set `m.queuedIDs` from the queue result
- `tui/delegate.go` — `resolveStyle` receives `queuedIDs map[string]bool` (or reads from model); add cyan path: not selected + in queue → `lipgloss.Color("6")`; priority: selected (green) > cyan (queued) > yellow (unapproved) > plain
- `cmd/tui.go` — wire `WithListQueue(func() ([]string, error) { rows, err := orm.New(sqlDB).ListQueueByProject(ctx, projectID); if err != nil { return nil, err }; ids := make([]string, len(rows)); for i, r := range rows { ids[i] = r.SubjectID }; return ids, nil })`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestResolveStyle_Queued_IsCyan`
     - **Behavior:** a queued, approved, non-selected task renders with the cyan color
     - **Setup:** delegate (or resolveStyle call) with `queuedIDs = map[string]bool{"BIT-33": true}`; task `{ID: "BIT-33", Approved: true}`; not selected
     - **Assertions:** returned style contains `lipgloss.Color("6")` (terminal cyan, ANSI 36m)
     - **Boundary:** `Approved: true` — proves cyan overrides the plain approved style (not just the unapproved yellow)
   - [ ] Confirm fails: `resolveStyle` has no cyan path; approved non-selected task returns plain style

2. **Implement (GREEN):**
   - [ ] Add `queuedIDs map[string]bool` to model; thread it into `resolveStyle` (pass as param or via delegate struct field)
   - [ ] Add cyan branch to `resolveStyle`: `if queuedIDs[t.ID] && !selected { return cyanStyle }`

3. **More tests (RED → GREEN):**
   - [ ] `TestResolveStyle_Selected_BeatsQueuedCyan`
     - **Behavior:** a queued + selected task renders with the selected (green) style, not cyan
     - **Setup:** same `queuedIDs`, task selected
     - **Assertions:** style is `selectedStyle` / `selectedBoardStyle`, not cyan
     - **Boundary:** `selected=true` — contradicts any "queued always wins" implementation

   - [ ] `TestUpdate_Tick_PopulatesQueuedIDs`
     - **Behavior:** after a tick cycle completes, `m.queuedIDs` reflects what `listQueue` returned
     - **Setup:** model with `WithListQueue` stub returning `["BIT-33"]`; fire `tickMsg{}`
     - **Assertions:** after `Update` processes `reloadedMsg`: `m.queuedIDs["BIT-33"] == true`; `m.queuedIDs["BIT-99"] == false`
     - **Boundary:** one item in queue — proves the map is populated from the stub result (not left empty)

## Claude verifies
- [ ] `go test ./tui/...` passes
- [ ] `go build ./...` passes
- [ ] `just install`

## User verifies
- [ ] In `bp tui` with a queued item (from Bar 2 or Bar 3): the queued task renders cyan in the list; navigate to a different task and back — color persists; if the queue row is deleted from sqlite3 directly the color clears on the next reload tick (≤1s)
- [ ] Whole slice: enqueue a track via `y` at the play prompt and via `e` shortcut; confirm cyan appears for both and clears after manual queue deletion — the declutter goal lands

## Commit (user)
`feat(tui): queued tasks render cyan; reload loop pulls queue state`