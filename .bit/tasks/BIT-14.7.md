---
id: BIT-14.7
title: Drop the list title heading
status: done
phase: 3
phase_label: Declutter
---
## **Verse 3**

Hides the bubbles default `List` title heading on the main list — pure noise sitting above the tasks. The board columns already suppress it via `SetShowTitle(false)`; the main list never did.

## Scope
- `tui/model.go` — `New()`: after building the list, call `l.SetShowTitle(false)`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestView_ListHidesTitleHeading`
     - **Behavior:** the main list no longer renders the `List` title heading.
     - **Setup:** `New(nil)`; WindowSize 60x16; `view := View().Content`.
     - **Assertions:** `strings.Contains(view, "List")` is false. (The test controls the tasks — no task title contains "List" — so the only source of that substring is the heading chip.)
     - **Boundary:** item count == 0 — the heading renders regardless of count; the empty state is where the noise is most glaring.
   - [ ] Confirm fails: the list currently renders the `List` chip (`…List\x1b[m…`).

2. **Implement (GREEN):**
   - [ ] Add `l.SetShowTitle(false)` in `New()`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic (the declutter verse's observe-it check lands on its last bar).

## Commit (user)
`feat(tui): drop the list title heading`
