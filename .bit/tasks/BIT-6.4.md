---
id: BIT-6.4
title: '`--parent` filters list to a track''s bars'
status: done
phase: 3
phase_label: Read one track's bars
---
## Step 4 (Phase 3 — Read one track's bars) — `--parent` filters list to a track's bars
**Status:** ✅ Done — verified 2026-07-20

The roll-up (derive a track's status from its bars' statuses) has to dump every task and
grep the ID prefix. This adds `--parent <id>` to `list`, narrowing output to that track's
direct children. Without the flag, `list` is unchanged.

**Scope:**
- `cmd/task_list.go` — add a `--parent` string flag; when set, keep only tasks whose ID has the `parent + "."` prefix (the parent's direct bars), preserving the existing sort and columns.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskListCmd_FiltersToParentBars` in `cmd/task_list_test.go`
     - **Behavior:** `--parent BIT-1` shows only BIT-1's bars — not the BIT-1 track itself, and not another track's bars — so the roll-up reads exactly one track's children.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "One", "...")`; `createTask(t, "Two", "...")`; create bars under BIT-1 (`One.1`, `One.2`) and under BIT-2 (`Two.1`) via `mustRun(..., "--parent", ...)`; `out := mustRun(t, "task", "list", "--parent", "BIT-1")`.
     - **Assertions:** `out == "BIT-1.1\ttodo\tOne.1\t\nBIT-1.2\ttodo\tOne.2\t\n"` (bars only, ascending, existing four-column format).
     - **Boundary:** a parent with N>1 bars amid a sibling track's bars — exercises both the include (own bars) and exclude (track row + other track's bars) paths.
   - [x] `TestTaskListCmd_ParentWithNoBars` in `cmd/task_list_test.go`
     - **Behavior:** filtering to a track that has no bars (or doesn't exist) yields empty output rather than the full list.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Lonely", "...")`; `out := mustRun(t, "task", "list", "--parent", "BIT-9")`.
     - **Assertions:** `out == ""`.
     - **Boundary:** zero matches — proves the filter fails closed (empty), never falls back to listing everything.
   - [x] Confirm fails: `--parent` is an unknown flag on `list` (`cobra.NoArgs`, no flags), so `mustRun` fails.

2. **Implement (GREEN):**
   - [x] Add `var parent string` + `cmd.Flags().StringVarP(&parent, "parent", "p", "", "list only this task's direct bars")`.
   - [x] In the print loop, when `parent != ""`, skip any `t` where `!strings.HasPrefix(t.ID, parent+".")`. Add the `strings` import.

**Claude verifies:**
- [x] `just test` — new tests pass; `TestTaskListCmd_ShowsNewestFirst` and the other unfiltered list tests stay green.
- [x] `just lint`

**User verifies:**
- [x] The four-column line format (`ID\tstatus\ttitle\tphase`) is what the roll-up will parse for each bar's status.

**Commit (user):** `feat(task): add task list --parent to show one track's bars`