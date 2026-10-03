---
id: BIT-49.24
title: A first migration ensures the global Claude wiring; a re-run doesn't
status: todo
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

The scope says the migration that registers a project ensures the global wiring (BIT-48), and a re-run doesn't. A recording runner that sees BIT-48's four calls on the first run and none on the second forces the call, and forces it to come after registration.

## Scope
- `cmd/migrate.go`: `newMigrateCmd(run claude.Runner)`. After a successful `migrate.Run` with `!res.Already`, print the `migrated` line, then `return ensureGlobalWiring(cmd, run)` (BIT-48, `cmd/wiring.go`). It returns that helper's error, which already prints the failing step and the commands. The project stays registered, so the exit is non-zero (BIT-48's rule). On `res.Already` there's no wiring. BIT-49.25 inserts the cleanup line between the two.
- `cmd/root.go`: `rootCmd.AddCommand(newMigrateCmd(run))`, with `run` from `newRootCmd(run claude.Runner)` (one parameter since BIT-45.1), the same as `newAddCmd(run)`.
- Test file touched: `cmd/migrate_test.go`. The expected calls are built from `claude.GlobalWiring()` (BIT-48), the same way BIT-48's add tests do. The sandboxed `HOME` keeps `ensureGlobalWiring` off the real `~/.claude.json`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/first migration ensures the global wiring`
     - **Behavior:** migrating the first v1 project sets bit up for the whole machine, the same as `bp add` does for a new project.
     - **Setup:** sandboxed `HOME`/`XDG_DATA_HOME`. A recording `claude.Runner` passed through `newRootCmd`, and a valid store at `dir`. Then `bp migrate`.
     - **Assertions:** the runner recorded exactly `claude.GlobalWiring()`'s four argvs in order (the MCP step is included, because the sandboxed `~/.claude.json` has no `mcpServers.bit`). The row exists, and the output starts with `migrated BIT`.
     - **Boundary:** a first registration.
   - [ ] Confirm fails: the runner recorded no calls.

2. **Implement (GREEN):**
   - [ ] The `run` parameter, the root wiring, and the `ensureGlobalWiring` call.

3. **More tests (RED → GREEN):**
   - [ ] `TestMigrateCmd/a re-run does not wire again`: a second `bp migrate` → `already migrated`, and the runner recorded nothing new. *Boundary:* the operator's 2026-10-01 rule.
   - [ ] `TestMigrateCmd/a wiring failure exits non-zero and keeps the project`: the runner fails the third call → `err != nil`, the row exists, and the store dir exists. *Boundary:* a failed wiring step after registration.
   - [ ] `TestMigrateCmd/a refused migration does not wire`: BIT-49.17's unknown-files case → the runner recorded nothing. *Boundary:* the wiring only follows a successful registration.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): first migration ensures the global Claude wiring`