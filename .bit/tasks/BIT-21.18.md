---
id: BIT-21.18
title: A lowercase --parent no longer hides a track's bars
status: done
phase: 2
phase_label: Recurrence
---
## **Verse 2**

The `--parent` filter is the last place a command still decides for itself which bars belong to
a track, and it decides using the raw flag value — so this step moves that decision behind
`Store`, where normalization already lives. What forces it is a test that passes the flag in
lowercase, exactly the way `/bit:do bit-21` types an ID.

## Scope
- `task/store.go` — export `Children(parent string) ([]*Task, error)` wrapping the private
  `children`, which already normalizes its argument.
- `cmd/task_list.go` — call `Children` when `--parent` is set instead of filtering `List()`
  output inline; the `strings` import goes with the filter.
- `cmd/task_list_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskListCmd_LowercaseParentStillListsTheBars`
     - **Behavior:** a track's bars are found by identity, so the case of the `--parent` value
       cannot hide them. This is the read-path twin of BIT-21.11 (the guard) and BIT-21.13 (the
       write path).
     - **Setup:** `initProject(t, "BIT")`; `createTask` for the track, then
       `task create --parent BIT-1` twice for two bars. Run `task list --parent bit-1` —
       lowercase flag against correctly-cased data, so the only wrong-case input is the flag.
     - **Assertions:** the output holds exactly two rows, naming `BIT-1.1` and `BIT-1.2`; it does
       not contain the track's own row (`BIT-1\t`).
     - **Boundary:** the flag's case is the only variable — `--parent BIT-1` passes today, so
       this is the lowercase end of the same input, and an empty list is the failure it produces.
   - [ ] Confirm fails: the command exits 0 and prints nothing, because `BIT-1.1` does not carry
     the prefix `bit-1.`

2. **Implement (GREEN):**
   - [ ] Add exported `Children(parent string) ([]*Task, error)` to `task/store.go`, returning
     `s.children(parent)`.
   - [ ] In `cmd/task_list.go`, choose the slice before the loop — `List()` when `parent` is
     empty, `Children(parent)` otherwise — then delete the `HasPrefix` continue and the
     now-unused `strings` import.
   - [ ] Expect one deliberate behaviour change: `children` matches *direct* bars via
     `barParent`, where `HasPrefix` also matched deeper IDs — which is what the flag's own help
     ("list only this task's direct bars") already claims. Ordering is unaffected, since
     `children` filters `List()`'s output and bars stay in step order.

## Claude verifies
- [ ] `just test` passes, including the two existing filter tests unchanged —
  `TestTaskListCmd_FiltersToParentBars` and `TestTaskListCmd_ParentWithNoBars`
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(task): list a track's bars through Store`
