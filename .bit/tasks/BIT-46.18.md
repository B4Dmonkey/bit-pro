---
id: BIT-46.18
title: A removed project's folder resolves to a "was removed" error
status: todo
approved: true
phase: 3
phase_label: bp remove soft-deletes a project
---
## **Verse 3**

The registry gains the removed flag. The resolver then refuses a removed project's folder with an error that points to `bp add`. A resolver test with a removed row forces the column, the query and the error.

## Scope
- `db/migrations/<initial>_create_projects.sql`: edit in place (the Decision: one initial migration until cutover) to add `removed INTEGER NOT NULL DEFAULT 0` (0/1, mapped to `bool` in Go; this avoids relying on sqlc's BOOLEAN mapping).
- `db/queries/projects.sql`: `ListProjects` also selects `removed` and still returns every row. New `SetProjectRemoved :exec` → `UPDATE projects SET removed = ? WHERE id = ?`.
- `project/resolve.go`: `Project` gains `Removed bool` (`Load` maps `row.Removed != 0`). New `var ErrRemoved = errors.New("this project was removed; run `bp add` here to revive it")`. In `Resolve`, if the longest match is removed, return `fmt.Errorf("%s: %w", dir, ErrRemoved)`. It doesn't fall through to an outer project, because the track says a removed project's folder says it was removed.
- The CLI and MCP get the error for free through `project.OpenStore`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestResolve/removed longest match is refused` (row added to the `TestResolve` table): `[{ACME, root}, {API, root/api, Removed: true}]`, dir `root/api/x` → `errors.Is(err, ErrRemoved)`, and the message contains `removed` and `` `bp add` ``.
     - **Behavior:** bp run in a removed project's folder says so instead of acting on it or on an enclosing project.
     - **Boundary:** the removed project is the longest match inside an active one; the outer project must not win.
   - [ ] Confirm fails: `Project` has no `Removed` field.

2. **Implement (GREEN):**
   - [ ] The column, the query changes, `Removed`, `ErrRemoved`.

3. **More tests (RED → GREEN):**
   - [ ] `TestSetProjectRemoved/keeps the row and toggles the flag` (subtest of a new `TestSetProjectRemoved`, `db/queries_test.go`): create, `SetProjectRemoved(1, id)` → `ListProjects` still returns the row with `Removed == 1`; set back to 0 → 0. *Boundary:* both transitions; the row survives (soft delete).
   - [ ] `TestFind/refuses a removed project` (subtest in `TestFind`, created in BIT-46.4): through the db, a removed row → `ErrRemoved`.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`
- [ ] delete any dev sandbox stores (the initial migration was edited; dbmate won't re-apply it)

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): registry tracks removed projects; resolver refuses them`