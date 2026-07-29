---
id: BIT-18.2
title: Force contradicts the hardcoded guard
status: todo
phase: 1
phase_label: Filed as completed
---
## **Verse 1**

The previous bar hardcodes `force` to `false`, so a track with a bar still `todo` cannot be
completed at all. That's the common case in practice — a track is often signed off with a
bar left unchecked — so `complete` needs the same escape hatch `archive` had. A test that
completes such a track contradicts the hardcoded `false`, and its twin pins the guard that
`--force` overrides.

## Scope
- `cmd/task_complete.go` — add `var force bool` and the `--force`/`-f` flag, pass it to
  `Complete`.
- `cmd/task_complete_test.go` — the two-state test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCompleteCmd_ForceCompletesUnfinished` in `cmd/task_complete_test.go`, a
     table over the two states of the flag.
     - **Behavior:** the all-bars-`done` guard applies to completing, and `--force`
       overrides it — so an unfinished track is refused by default but can still be filed
       when the human means it.
     - **Setup:** for each case, `initProject(t, "BIT")`; `createTask(t, "Track", "A track
       with an unfinished bar.")`; one bar via `mustRun(t, "task", "create", "Bar",
       "--parent", "BIT-1", "--description", "Still todo.")` left `todo`; then
       `run(t, "task", "complete", "BIT-1")` or
       `run(t, "task", "complete", "BIT-1", "--force")`.
     - **Assertions:** without `--force`, `errors.As(err, &unfinished)` holds for
       `*task.UnfinishedBarsError`, `unfinished.Bars` contains `BIT-1.1`, and both
       `.bit/tasks/BIT-1.md` and `.bit/tasks/BIT-1.1.md` still stat clean. With `--force`,
       `err` is nil and both files are under `.bit/completed/` with neither left in
       `.bit/tasks/`.
     - **Boundary:** the `force` flag at both of its two values, against a bar-status
       distribution that sits exactly on the guard's threshold (one bar, not `done`).
   - [ ] Confirm fails: the `--force` case errors with `unknown flag: --force`. The
     no-force case already passes — it rides on the guard the shared `relocateTree` brought
     over, and it is here to pin that the guard still applies to the new verb.

2. **Implement (GREEN):**
   - [ ] `cmd/task_complete.go`: `cmd.Flags().BoolVarP(&force, "force", "f", false,
     "complete even when bars are unfinished")`, and `RunE` passes `force` instead of
     `false`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(task): let --force complete a track with unfinished bars`