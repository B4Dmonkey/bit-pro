---
id: BIT-49.21
title: migrate on a removed project's folder refuses and points to bp add
status: done
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

A removed project's path is still a registry row (`bp remove` is soft). BIT-49.20's check skips removed rows, so a migrate there would try to register the path a second time. The scope says to refuse and point to `bp add`, which revives.

## Scope
- `migrate/migrate.go`: in the BIT-49.20 lookup, `ok && p.Removed` → `fmt.Errorf("%s: %w", path, project.ErrRemoved)`. `ErrRemoved`'s own text already says ``run `bp add` here to revive it`` (BIT-46.18), so nothing is appended.
- Test file touched: `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/refuses a removed project's folder`
     - **Behavior:** reviving a removed project goes through `bp add`, never a second registration.
     - **Setup:** sandbox. Migrate a valid store at `dir`, then `SetProjectRemoved` on its row (BIT-46's query). Then `bp migrate` again.
     - **Assertions:** `errors.Is(err, project.ErrRemoved)`, and the text contains `bp add`. `ListProjects` still has one row, `removed`. `<data>/bit` is unchanged, `main.db*` aside (as in BIT-49.20).
     - **Boundary:** a registered path whose row is removed.
   - [ ] Confirm fails: the run proceeds and fails on the rename or the UNIQUE path, or registers again.

2. **Implement (GREEN):**
   - [ ] The removed branch.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate refuses a removed project's folder`