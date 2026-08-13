---
id: BIT-24.3
title: The list pane header reads done over total
status: done
phase: 2
phase_label: Progress counter
---
## **Verse 2**

The list pane's title becomes a progress fraction — `Tasks (3/7)` instead of `Tasks (7)` —
counting rows at `done` over all rows. One bar: a table-driven test whose rows contradict
each other (0 done, some done, all done) so no constant satisfies them.

## Scope
- `tui/model.go` — `content()`, the `listTitle` line (`fmt.Sprintf("Tasks (%d)", len(m.Items()))`,
  line 376); add a small helper that counts items at `done`
- `tui/model_test.go` — `TestView_PaneTitles` (line 509) asserts the old `Tasks (29)` /
  `Tasks (1)` format and must be rewritten here, not left alongside a new test; its table
  gains a done count per case

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestView_PaneTitles` (table-driven; change the case struct from `count int` to
     `total, done int`, build `tasks[i] = &task.Task{ID: "BIT-1"}` and set `Status: "done"`
     on the first `done` of them). Keep the existing driver: `New(tasks)` →
     `Update(tea.WindowSizeMsg{Width: 80, Height: 24})` → `Update(tea.KeyPressMsg{Code: tea.KeyTab})`
     (the model opens in board mode; `tab` switches to the list), then assert on
     `mdl.(model).View().Content`.
     - **Behavior:** the list pane header reports how far along the work is, not a raw
       inventory count — done rows over all rows.
     - **Setup:** four cases —
       `{"none done", total: 7, done: 0}`, `{"some done", total: 7, done: 3}`,
       `{"all done", total: 3, done: 3}`, `{"empty", total: 0, done: 0}`.
     - **Assertions:** the view contains `Tasks (0/7)`, `Tasks (3/7)`, `Tasks (3/3)`,
       `Tasks (0/0)` respectively; each case still contains `Details`.
     - **Boundary:** the done count across its whole range against a fixed total — 0 (lower
       bound), 3 of 7 (interior, contradicts any hardcoded numerator), 3 of 3 (upper bound);
       plus total == 0, the empty-list lower bound where the denominator would divide-by-zero
       in any percentage-style formatting.
   - [ ] Confirm fails: the header still renders `Tasks (7)` / `Tasks (3)` / `Tasks (0)` —
     no `/` in the title.

2. **Implement (GREEN):**
   - [ ] Add `func doneCount(items []list.Item) int` to `tui/model.go` — range over `items`,
     type-assert each to `item`, count `it.t.Status == "done"` (skip anything that does not
     assert, matching `delegate.Render`'s own guard).
   - [ ] In `content()`, replace the `listTitle` line with
     `fmt.Sprintf("Tasks (%d/%d)", doneCount(m.Items()), len(m.Items()))`. This is inside
     the branch that runs after the `m.mode == modeBoard` early return, so the board title
     is untouched by construction.

## Claude verifies
- [ ] `just test` — the whole package, confirming no other test still asserts the old
      `Tasks (N)` format
- [ ] `just lint`

## User verifies
- [ ] Whole slice: `just install`, then `bp tui` and press `tab` to the list. The header
      reads `Tasks (n/m)` where `m` equals the row count from `bp task list` and `n` equals
      the number of rows showing `✓`.

## Commit (user)
`feat(tui): show done-over-total in the list pane header`