---
id: BIT-38.4
title: Store.Update owns approval revocation
status: done
approved: true
phase: 1
phase_label: Scope writes
---
## **Verse 1**

Approval revocation lives in `cmd/task/update.go:26-54`, keyed on Cobra's `Changed()`. It is the riskiest rule in the scope — an approved bar has to stay approved for a whole automated run — so it moves down before any tool depends on it. Forced by a new store-level test: `Store.Update` does not exist yet.

## Scope
- `task/store.go` — new `Patch` type and `(*Store).Update` method.
- `task/store_test.go` — new tests for `Update`.
- `cmd/task/update.go` — `RunE` becomes: build a `Patch` from `Changed()`, call `Update`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreUpdate_AppliesOnlySetFields` (table-driven)
     - **Behavior:** a `nil` field in a `Patch` leaves the stored value alone; a non-nil field writes it, including an explicitly empty string. This is the store's version of Cobra's `Changed()` check.
     - **Setup:** `s := New(t.TempDir())`, seed `&Task{ID: "BIT-1", Title: "Old title", Status: StatusTodo, Phase: 3, PhaseLabel: "Old label", Body: "Old body."}`. Cases: (a) zero `Patch{}`; (b) `Patch{Title: ptr("New title")}`; (c) `Patch{Body: ptr("New body."), Status: ptr(StatusDoing)}`; (d) `Patch{Title: ptr("")}`; (e) `Patch{Phase: ptr(0), PhaseLabel: ptr("")}`.
     - **Assertions:** the returned `*Task` and a fresh `s.Load("BIT-1")` agree, and each equals the seed with exactly the patched fields changed. (a) is byte-identical to the seed. (d) has `Title` `""` and `Body` `"Old body."`. (e) has `Phase` 0 and `PhaseLabel` `""` with `Title` untouched.
     - **Boundary:** every patch field at both states of the nil/non-nil distinction that is the whole contract. (d) and (e) sit at the dangerous end — an empty string and a zero int are indistinguishable from "unset" under a value-typed patch, which is exactly the accident that would blank a body or zero a phase tag.
   - [ ] Confirm fails: `s.Update undefined (type *Store has no field or method Update)`
   - [ ] `TestStoreUpdate_ApprovalRevocation` (table-driven)
     - **Behavior:** the rule, verbatim: a change to title, body, phase, or phase-label revokes approval; a status write of `todo` revokes it; a forward status move keeps it.
     - **Setup:** seed `&Task{ID: "BIT-1", Title: "T", Status: StatusDoing, Approved: true}` per case. Cases and wanted `Approved`: `Patch{Title: ptr("x")}` → false; `Patch{Body: ptr("x")}` → false; `Patch{Phase: ptr(2)}` → false; `Patch{PhaseLabel: ptr("x")}` → false; `Patch{Status: ptr(StatusTodo)}` → false; `Patch{Status: ptr(StatusDone)}` → **true**; `Patch{}` → true. Plus one case seeded `Approved: false` with `Patch{Title: ptr("x")}` → false (the rule must not somehow set it).
     - **Assertions:** returned task's `Approved` and the reloaded task's `Approved` both equal the wanted value.
     - **Boundary:** `Status` across all three enum values — `todo` (the one that revokes), `doing`/`done` (forward moves that must not), and absent. `Approved` in both its states, so "revoke" is distinguished from "assign false".

2. **Implement (GREEN):**
   - [ ] `type Patch struct { Title, Body, Status *string; Phase *int; PhaseLabel *string }` in `task/store.go`.
   - [ ] `func (s *Store) Update(id string, p Patch) (*Task, error)`: `s.Load(id)`; apply each non-nil field; compute `anyChanged := p.Title != nil || p.Body != nil || p.Phase != nil || p.PhaseLabel != nil` and `sentBack := p.Status != nil && *p.Status == StatusTodo`; `if t.Approved && (anyChanged || sentBack) { t.Approved = false }`; `s.Save(t)`; return `t`.
   - [ ] Rewrite `newUpdateCmd`'s `RunE` to build a `Patch` — one `if cmd.Flags().Changed("<flag>")` per field setting the pointer — and call `s.Update(args[0], p)`, discarding the returned task. Delete the field application and the revocation block from `cmd/task/update.go`.
   - [ ] Add a small `func ptr[T any](v T) *T` test helper in `task/store_test.go` if the package has none.

## Claude verifies
- [ ] `just test` — new store tests pass, and `cmd/task/update_test.go` (including its `explicitly empty title is applied` case and the approval cases) passes unchanged
- [ ] `just lint`
- [ ] `grep -n "Approved" cmd/task/update.go` returns nothing

## User verifies
- [ ] none — deterministic

## Commit (user)
`refactor(task): move approval revocation into Store.Update`