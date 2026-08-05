---
id: BIT-20.10
title: left/right/h/l page the list when expanded, focus-switch when collapsed
status: todo
phase: 3
phase_label: Expanded detail pane
---
## **Verse 3 (final)**

Closes Verse 3: while `detailExpanded`, left/right/h/l page to the previous/next item in the
list instead of switching pane focus. Collapsed (the normal split), left/right keep their
existing meaning — this bar also adds `h`/`l` as aliases there, since the scope's h/l bindings
were only ever wired into the modal's scroll handling, never into list view's focus switch.
Paging clamps at the first/last item — no wraparound — matching the list's own existing
selection-clamp behavior (`up`/`down` already stop at the ends via the embedded `list.Model`).

## Scope
- `tui/model.go` — `Update`'s list-mode `KeyPressMsg` handling: replace the
  `tea.KeyRight`/`tea.KeyLeft`-only `msg.Code` switch with a `msg.String()` match on
  `"left", "right", "h", "l"`, branching on `m.detailExpanded`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_PagesListWhenExpanded`
     - **Behavior:** with the detail pane expanded, `right`/`l` moves to the next item.
     - **Setup:** `New([]*task.Task{{ID: "BIT-2"}, {ID: "BIT-1"}})`, `Tab` to list mode,
       `Enter` to expand.
     - **Assertions:** after `right`, `m.Index() == 1`.
     - **Boundary:** the simplest forward step — one item, from the start of the list.
   - [ ] Confirm fails: `right` still sets `detailFocused = true` (today's unconditional
     behavior), `Index()` stays `0`.

2. **Implement (GREEN):**
   - [ ] Minimal branch: `if m.detailExpanded { if right/l { m.Select(m.Index()+1) } else { m.Select(m.Index()-1) }; m.refreshDetail(); return m, nil }` else fall through to the
     existing focus-switch behavior (now also matching `h`/`l`, not just arrow keys).

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_PagesListLeftWhenExpanded`
     - **Behavior:** `left`/`h` moves backward — proves the direction isn't hardcoded to
       "always forward".
     - **Setup:** three tasks, expanded, start at index 2 (`m.Select(2)` before the key press).
     - **Assertions:** after `left`, `m.Index() == 1`.
     - **Boundary:** contradicts a forward-only hardcode — forces a real signed delta.
   - [ ] `TestUpdate_PagingClampsAtListEnds`
     - **Behavior:** paging past either end of the list holds at that end.
     - **Setup:** two tasks, expanded; at index 0, press `left`; separately at index 1
       (the last), press `right`.
     - **Assertions:** `Index()` stays `0` and stays `1` respectively.
     - **Boundary:** both ends — `idx == 0` and `idx == len-1`.
   - [ ] `TestUpdate_FocusStillSwitchesWhenCollapsed` (extends the existing
     `TestUpdate_Focus` table in `tui/model_test.go` rather than duplicating it — add `h`/`l`
     cases alongside the existing arrow-key ones, e.g. `{"h focuses list like left", []rune{tea.KeyRight, 'h'}, false}`, `{"l focuses detail like right", []rune{'l'}, true}`)
     - **Behavior:** collapsed (the default), left/right/h/l still only switch pane focus —
       today's behavior is unchanged, and h/l now do what left/right already did.
     - **Setup:** `m.detailExpanded == false` (the default after the `Tab`-to-list-mode setup
       these tests already do).
     - **Assertions:** `detailFocused` matches the existing table's expectations, with h/l
       producing the same result as their arrow-key equivalents.
     - **Boundary:** the collapsed state — the other half of the `detailExpanded` gate.
   - [ ] Implement the real dispatch: match `msg.String()` against `"left", "right", "h", "l"`;
     if `m.detailExpanded`, compute `delta := 1` for `right`/`l`, `-1` for `left`/`h`, then
     `m.Select(min(max(m.Index()+delta, 0), len(m.Items())-1))` and `m.refreshDetail()`;
     otherwise set `detailFocused = (msg.String() == "right" || msg.String() == "l")`.

## Claude verifies
- [ ] `go test ./tui/...` passes, full package — including the extended `TestUpdate_Focus`
  and every other Verse 1–3 test.
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] Whole slice: in `bp tui`'s list view, press `Enter` on a task with a long body — the
  detail pane expands to a comfortable reading width; page to the next/previous task with
  `l`/`h` (or the arrows) without collapsing; press `Enter` again — it collapses back to the
  normal list/detail split, and left/right/h/l go back to switching focus between the two
  panes.

## Commit (user)
`feat(tui): page the expanded detail pane with left/right/h/l`