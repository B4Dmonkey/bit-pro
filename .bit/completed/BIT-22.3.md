---
id: BIT-22.3
title: Expanded footer labels the arrow keys for paging
status: done
phase: 2
phase_label: Footer
---
## **Verse 2**

The footer labels ←/→ as `focus` in both states, but while expanded those keys page between
tasks. `helpKeys()` (line 323) returns the one static `m.keys` regardless of state, so the fix
is to vary the binding's help text with `detailExpanded`.

## Scope
- `tui/model.go` — `helpKeys()`, and the `focus` binding built in `newKeyMap()`
- `tui/model_test.go` — new test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestView_FooterLabelsArrowsForCurrentState` (table-driven, two cases)
     - **Behavior:** The footer names what ←/→ actually do in the state the operator is in —
       `focus` while collapsed, `page` while expanded.
     - **Setup:** `New([]*task.Task{{ID: "BIT-2"}, {ID: "BIT-1"}})` with empty bodies so no
       rendered body text can contain either word; `Update(tea.WindowSizeMsg{Width: 80,
       Height: 24})` then `KeyTab`. Case "collapsed": no further keys. Case "expanded": one
       `KeyEnter`. Read `View().Content` in both.
     - **Assertions:** collapsed — contains `"focus"`, does not contain `"page"`; expanded —
       contains `"page"`, does not contain `"focus"`.
     - **Boundary:** `detailExpanded` in both of its two states. The collapsed row is what
       makes a constant label insufficient, so the implementation has to branch.
   - [ ] Confirm fails: the expanded case only — `View().Content` still contains `"focus"` and
     not `"page"`. The collapsed case passes from the start; that is the point of including it.

2. **Implement (GREEN):**
   - [ ] In `helpKeys()`, when `m.detailExpanded`, return a `keyMap` copy whose `focus`
     binding is rebuilt with `key.WithHelp("←/→", "page")`. Same keys, same bindings
     elsewhere — only the help text differs.

## Claude verifies
- [ ] `just test` — `TestView_HelpBarPresentAndBounded` asserts `"focus"` is present in the
      collapsed list view and must stay green untouched
- [ ] `just lint`

## User verifies
- [ ] Whole slice: in `bp tui`, press Tab, read the footer — `←/→ focus`. Press Enter — the
      footer now reads `←/→ page`, and pressing ←/→ does page between tasks. Press `?` while
      expanded; the full help shows the same `page` label.

## Commit (user)
`fix(tui): label the arrow keys for the expanded state in the help footer`