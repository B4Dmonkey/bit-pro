---
id: BIT-38.6
title: 'Contradiction: an omitted param leaves its field alone'
status: done
approved: true
phase: 1
phase_label: Scope writes
---
## **Verse 1**

Bar 1.5's handler always patches the body, so an update that never mentions the body blanks it. This test cannot pass under that reading — it is the contradiction that forces the real "an omitted param means leave unchanged" contract, and with it the rest of the tool's params.

## Scope
- `cmd/serve_mcp.go` — `taskUpdateInput` gains pointer fields for `title`, `body`, `status`, `phase`, `phase_label`; the handler forwards them as a whole `task.Patch`.
- `cmd/serve_mcp_write_test.go` — the contradicting test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskUpdateLeavesOmittedFieldsAlone` (table-driven)
     - **Behavior:** every field the caller does not send survives the update untouched, and every field it does send is written — including a phase tag, which the CLI could only express through `--phase`.
     - **Setup:** `dir := t.TempDir()`; per case, `seedTasks(t, dir, &task.Task{ID: "FOO-1.1", Title: "Contradiction forces real fan-out", Status: task.StatusTodo, Phase: 2, PhaseLabel: "Plan writes", Body: "## **Verse 2**\n\nstep detail"})`; `s := mcpSession(t, dir)`. Cases: (a) `{"id": "FOO-1.1", "status": "doing"}`; (b) `{"id": "FOO-1.1", "title": "Renamed step"}`; (c) `{"id": "FOO-1.1", "phase": 3, "phase_label": "Run writes"}`; (d) `{"id": "FOO-1.1"}` alone.
     - **Assertions:** for each case, `Load("FOO-1.1")` equals the seed with only the sent fields changed. (a) `Status` `doing` with `Body`, `Title`, `Phase`, `PhaseLabel` intact. (b) `Title` `Renamed step` with `Body` and `Status` intact. (c) `Phase` 3 and `PhaseLabel` `Run writes` with `Title` and `Body` intact. (d) byte-identical to the seed. Compare the whole `task.Task` with `reflect.DeepEqual`, matching the style of `cmd/task/update_test.go`, so an unintended change to any field fails the case.
     - **Boundary:** each param at both states of the present/absent distinction, and (d) at the count-of-set-fields lower bound of 0 — a request that names only an ID must be a no-op rather than a five-field blanking.
   - [ ] Confirm fails: case (a) reports `Body = "", want "## **Verse 2**\n\nstep detail"` — bar 1.5's handler passes `&in.Body` unconditionally, so an absent `body` writes the empty string. Cases (b) and (c) fail the same way; (d) fails on `Body` too.

2. **Implement (GREEN):**
   - [ ] Change `taskUpdateInput` to `ID string` (`json:"id"`) plus `Title`, `Body`, `Status` `*string` and `Phase *int`, `PhaseLabel *string`, each with `json:",omitempty"`. A nil pointer is what "leave unchanged" means on the wire; the SDK's inferred schema marks an `omitempty` pointer field optional and unmarshals an absent key to nil. (An explicit JSON `null` also arrives as nil, so it reads as "leave unchanged" too.)
   - [ ] Change the handler to forward the whole patch: `store.Update(in.ID, task.Patch{Title: in.Title, Body: in.Body, Status: in.Status, Phase: in.Phase, PhaseLabel: in.PhaseLabel})`. The pointers pass straight through with no conversion, which is why both layers use the same shape.
   - [ ] Extend the tool's `Description` to say that an omitted field is left unchanged.

## Claude verifies
- [ ] `just test` — 1.5's test still passes alongside the new one
- [ ] `just lint`

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(mcp): leave omitted task_update params unchanged`