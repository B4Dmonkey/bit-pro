---
id: BIT-49.20
title: Re-running migrate on a registered project says "already migrated" and changes nothing
status: todo
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

A second `bp migrate` today fails part-way, on the `ProjectDir` rename or the UNIQUE path. The scope's idempotency means a re-run must stop cleanly before anything else happens, even if the v1 `.bit/` has changed since.

## Scope
- `migrate/migrate.go`:
  - `Result` gains `Already bool`.
  - Right after the source `path` is known, and before any check or staging: `ps, err := project.Load(ctx, q)` (an error is returned). If `p, ok := project.ByPath(ps, path); ok && !p.Removed`, return `Result{Code: p.Code, Path: p.Path, Already: true}, nil`. BIT-49.21 handles the removed case.
- `cmd/migrate.go`: on `res.Already`, print `already migrated` and return nil (`bp add` returns nil for `already added`).
- Test file touched: `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/a re-run says already migrated and changes nothing`
     - **Behavior:** migrate is safe to run twice. The second run is a no-op.
     - **Setup:** sandbox. Migrate a valid store once. Then add `notes.txt` to `.bit/` (which would be an unknown file), snapshot every path and mtime under `<data>/bit` except `main.db*` (opening the registry may touch it and its `-wal`/`-shm` files), and run `bp migrate` again.
     - **Assertions:** the output is `already migrated\n` and `err == nil`. `ListProjects` has one row. The `<data>/bit` snapshot is identical.
     - **Boundary:** a registered path, with a source that changed after the first run.
   - [ ] Confirm fails: `ErrUnknownFiles`, or a rename error.

2. **Implement (GREEN):**
   - [ ] The registry check, and `Already` handling in the command.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate re-runs on a registered project are a no-op`