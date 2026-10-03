---
id: BIT-46.21
title: bp add on a removed project's path revives it with its archive intact
status: done
approved: true
phase: 3
phase_label: bp remove soft-deletes a project
---
## **Verse 3**

`bp add` on a removed project's path clears the flag. It doesn't prompt for a code, doesn't run wiring, and leaves the archived tracks in `archive/tasks/`, as after `bp task delete`. Today that path prints `already added`, which is the contradiction that forces the revive branch.

## Scope
- `project/resolve.go`: extract the exact-path, case-insensitive lookup from `cmd/add.go` into `func ByPath(projects []Project, path string) (Project, bool)` (`strings.EqualFold` on canonical paths). **Seam for BIT-49:** migrate's "already migrated" and "removed path → refuse, point to `bp add`" checks.
- `cmd/add.go`: after `project.Load`, `if p, ok := project.ByPath(projects, path); ok`:
  - `p.Removed` → `SetProjectRemoved(0, p.ID)`, print `revived <CODE> <path>`, return nil (no prompt, no wiring).
  - otherwise `already added`, as before.
  - Both branches come before the `.bit/` check, as today. Keep it a separate early return; BIT-48 inserts wiring only on the fresh branch.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAddCmd/revives a removed project` (subtest in `TestAddCmd`, holding a table whose rows differ only in how the path is typed; first row `same case`)
     - **Behavior:** re-adding a removed project brings it back under its old code, with its archived work still archived and its IDs still reserved.
     - **Setup:** `dir := initProject(t, "BIT")`; create `BIT-1`, `BIT-2`; `bp remove` with `"y\n"`; recording runner; `bp add <dir>` with empty stdin.
     - **Assertions:** output `revived BIT <canonical dir>\n` (no `Project code` prompt); runner recorded no calls; the row's `Removed == false`; `bp task list` succeeds with no tracks; `archive/tasks/BIT-1.json` still exists; `bp task create "Next"` prints `BIT-3`.
     - **Boundary:** a removed row × the same path; ID minting after revive skips the archived 1–2.
   - [ ] Confirm fails: output `already added`, and `bp task list` errors with `ErrRemoved`.

2. **Implement (GREEN):**
   - [ ] `ByPath`; the revive branch.

3. **More tests (RED → GREEN):**
   - [ ] row `different case` in the `TestAddCmd/revives a removed project` table: the same with the path typed in a different case → revived. *Boundary:* `ByPath` is case-insensitive.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none here (the verse's end-to-end check is on its last bar)

## Commit (user)
`feat(bit): bp add revives a removed project`