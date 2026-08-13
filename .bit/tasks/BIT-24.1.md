---
id: BIT-24.1
title: A doing row renders the in-progress marker
status: done
phase: 1
phase_label: In-progress marker
---
## **Verse 1**

Teach the list row renderer a second mark: a row whose status is `doing` renders `→` where a
done row renders `✓`. Forced by a test that a `doing` row is visually distinct from a `todo`
one — today both render the two-space blank.

## Scope
- `tui/delegate.go` — `delegate.Render`, the `mark` block at lines 54-57
- `tui/delegate_test.go` — new test alongside `TestDelegate_DoneRowShowsMarker`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDelegate_DoingRowShowsInProgressMarker`
     - **Behavior:** a list row for work in progress is distinguishable from an untouched
       one without opening it — the mark column carries `→`.
     - **Setup:** `l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Track", Status: "doing"}}}, delegate{}, 40, 4)`;
       render index 0 into a `bytes.Buffer` with `delegate{}.Render(&buf, l, 0, l.Items()[0])`.
     - **Assertions:** `buf.String()` contains `"→"`; does not contain `"✓"`.
     - **Boundary:** `Status == "doing"` — the middle of the three values the CLI writes.
       The other two are already pinned: `todo` by `TestDelegate_UnfinishedRowHasNoMarker`,
       `done` by `TestDelegate_DoneRowShowsMarker`.
   - [ ] Confirm fails: `mark` stays `"  "` for any non-`done` status, so the rendered row
     contains no `→`.

2. **Implement (GREEN):**
   - [ ] In `delegate.Render`, extend the mark block to a second branch:
     `else if t.Status == "doing" { mark = "→ " }`. Keep the two-cell width so row
     alignment matches `"✓ "` and `"  "`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic; the verse's end-to-end check lands on the next bar.

## Commit (user)
`feat(tui): mark in-progress rows with an arrow`