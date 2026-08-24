---
id: BIT-38.2
title: Store.Create owns ID minting and order insertion
status: done
approved: true
phase: 1
phase_label: Scope writes
---
## **Verse 1**

`cmd/task/create.go:37` holds ID minting and order insertion, so an MCP handler creating a task would have to re-implement both. This moves that sequence into `task.Store` and leaves the Cobra command as a thin caller. Forced by a new store-level test: `Store.Create` does not exist yet.

## Scope
- `task/store.go` — new `CreateParams` type and `(*Store).Create` method, beside `Save` and `SetApproved`.
- `task/store_test.go` — new tests for `Create`.
- `cmd/task/create.go` — `runCreate` becomes: build `CreateParams`, call `Create`, print the returned task's ID.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreCreate` (table-driven)
     - **Behavior:** `Store.Create` mints the next ID, writes the file, and maintains the parent's explicit order — the whole sequence `runCreate` performs today, callable without Cobra.
     - **Setup:** `s := New(t.TempDir())`; `s.SaveConfig(&Config{Prefix: "BIT"})`. Cases: (a) no parent, empty store → `CreateParams{Title: "Track", Body: "why\n\nbecause"}`; (b) no parent with `BIT-1` already saved; (c) parent `BIT-1` present, `CreateParams{Title: "Bar", Parent: "BIT-1", Phase: 2, PhaseLabel: "Plan writes"}`; (d) parent `BIT-1` with `Order: []string{"BIT-1.1", "BIT-1.2"}` and both bars saved, `CreateParams{Title: "Bar", Parent: "BIT-1", After: "BIT-1.1"}`.
     - **Assertions:** returned `*Task` has ID `BIT-1` / `BIT-2` / `BIT-1.1` / `BIT-1.3` respectively; `Status` is `StatusTodo`; `Title`, `Body`, `Phase`, `PhaseLabel` are what was passed; `s.Load(id)` returns the same values, so the file was written. For (d) `s.Load("BIT-1").Order` is `["BIT-1.1", "BIT-1.3", "BIT-1.2"]`.
     - **Boundary:** `Parent` at both values of its only meaningful distinction (empty → prefix minting via `NextID`, non-empty → dotted minting via `NextChildID`), and `After` in both states (empty → append semantics, set → splice). Case (b) puts the existing-task count at 1 rather than 0, which is what separates real minting from an always-`-1` return.
   - [ ] Confirm fails: `s.Create undefined (type *Store has no field or method Create)`
   - [ ] `TestStoreCreate_RejectsUnknownParent` — `CreateParams{Parent: "BIT-9"}` against an empty store returns a non-nil error wrapping `fs.ErrNotExist` (this is `NextChildID`'s existing guarantee; the test pins that `Create` does not swallow it).

2. **Implement (GREEN):**
   - [ ] `type CreateParams struct { Title, Body, Parent, After string; Phase int; PhaseLabel string }` in `task/store.go`.
   - [ ] `func (s *Store) Create(p CreateParams) (*Task, error)` performing, in this order (matching `runCreate` today — the `After` failure path must still run before any file is written, so a bad anchor leaves no orphan task file):
       1. mint: `p.Parent != ""` → `s.NextChildID(p.Parent)`; else `s.Config()` then `s.NextID(cfg.Prefix)`. Wrap nothing new — both already carry context.
       2. if `p.After != ""` → `s.InsertAfter(p.Parent, id, p.After)`, returning on error.
       3. build `&Task{ID: id, Title: p.Title, Status: StatusTodo, Phase: p.Phase, PhaseLabel: p.PhaseLabel, Body: p.Body}` and `s.Save` it.
       4. if `p.Parent != "" && p.After == ""` → `s.AppendToOrder(p.Parent, id)`.
       5. return the task.
   - [ ] Rewrite `runCreate` to build `CreateParams` from its arguments, call `s.Create`, and `fmt.Fprintln(cmd.OutOrStdout(), t.ID)`. Delete the minting, `InsertAfter`, `Save`, and `AppendToOrder` calls from `cmd/task/create.go` — leaving a second copy is the drift this bar exists to remove.

## Claude verifies
- [ ] `just test` — new store tests pass, and `cmd/task/create_test.go` passes unchanged (it is what proves the CLI path still behaves identically)
- [ ] `just lint`
- [ ] `grep -n "NextChildID\|NextID\|AppendToOrder\|InsertAfter" cmd/task/create.go` returns nothing

## User verifies
- [ ] none — deterministic

## Commit (user)
`refactor(task): move ID minting and order insertion into Store.Create`