---
id: BIT-46.20
title: bp list leaves out removed projects
status: todo
phase: 3
phase_label: bp remove soft-deletes a project
---
## **Verse 3**

A removed project shouldn't look active. `ListProjects` returns every row (resolver, add and BIT-49's migrate need removed rows), so `bp list` filters in Go. A list test with one removed row forces the filter.

## Scope
- `cmd/list.go`: skip rows with `Removed != 0`.
- `cmd/list_test.go` (converted in BIT-45.4 to the `TestListCmd` table with rows `three projects` and `no database yet`).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestListCmd/hides removed projects` (row added to the `TestListCmd` table; the table gains a field naming the seeded codes to mark removed, empty for the existing rows)
     - **Behavior:** `bp list` shows only active projects.
     - **Setup:** sandboxed `HOME`/`XDG_DATA_HOME`; `seedProject` `{/tmp/ace, ACE}` and `{/tmp/mid, MID}`; `SetProjectRemoved(1, <MID id>)`; `bp list`.
     - **Assertions:** `normalizeSpaces(out) == "ACE /tmp/ace"`.
     - **Boundary:** one active, one removed.
   - [ ] Confirm fails: output also has `MID /tmp/mid`.

2. **Implement (GREEN):**
   - [ ] The filter.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): bp list hides removed projects`