---
id: BIT-49.22
title: 'migrate refuses a code bp add would refuse: invalid, reserved, or held by another project'
status: done
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

The v1 prefix becomes the project code. A prefix of `FEEDBACK` would name a project dir that collides with the shared folder. A prefix another project already holds, live or removed, would fail late on the UNIQUE constraint after staging. Each case must refuse before any write, by the same rules `bp add` follows (BIT-46).

## Scope
- `migrate/migrate.go`:
  - Right after reading the prefix: `code, err := project.ValidateCode(rawPrefix)` → returns `ErrInvalidCode`/`ErrReservedCode`, wrapped with `config.toml`. `badIDs` (BIT-49.19) still receives the raw prefix, so a lowercase prefix is still refused even though `ValidateCode` uppercases it.
  - After the BIT-49.20/21 path lookup and before the file checks, `if p, ok := project.ByCode(ps, code); ok`:
    - `p.Removed` → `fmt.Errorf("code %s: %w", code, project.ErrCodeRemoved)`;
    - otherwise → `fmt.Errorf("code %s is already used by %s: %w", code, p.Path, ErrCodeTaken)`, with `var ErrCodeTaken = errors.New("code belongs to another project")`.
  - There's no code rename (scope).
- Test file touched: `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/refuses a code another project holds`
     - **Behavior:** two v1 projects with the same prefix can't both be migrated, and the second is told which folder holds the code.
     - **Setup:** sandbox. Migrate store `a` (`BIT`). Then from folder `b`, whose store also uses `prefix = "BIT"`, run `bp migrate`.
     - **Assertions:** `errors.Is(err, migrate.ErrCodeTaken)`, and the text contains `a`'s canonical path. One row. No `.migrate-*` dir, and `<data>/bit/BIT` holds only `a`'s records.
     - **Boundary:** a live code collision.
   - [ ] Confirm fails: the copy is staged, then the `ProjectDir` rename or `CreateProject` fails.

2. **Implement (GREEN):**
   - [ ] `ValidateCode` and the `ByCode` lookup.

3. **More tests (RED → GREEN):**
   - [ ] `TestMigrateCmd/refuses a removed project's code`: `a` migrated, then removed (`SetProjectRemoved`); migrate `b` with `BIT` → `errors.Is(err, project.ErrCodeRemoved)`. *Boundary:* a removed code stays reserved.
   - [ ] `TestMigrateCmd/refuses an invalid or reserved code` (table): `prefix = "FEEDBACK"` → `ErrReservedCode`; `prefix = "BIT-PRO"` → `ErrInvalidCode`. Nothing is written in either case. *Boundary:* BIT-46's code format and reserved names.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate refuses codes bp add would refuse`