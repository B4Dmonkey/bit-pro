---
id: BIT-29.5
title: Re-adding an enrolled path is a no-op
status: done
phase: 2
phase_label: add
---
## **Verse 2**

Enrolling a path that is already in the registry prints `already added` and stops — no prompt, no
second row. Contradicts BIT-29.4, which prompts and inserts unconditionally and so fails the
`projects.path` UNIQUE constraint on a second run.

## Scope
- `db/queries/projects.sql` — a `ProjectExists` query
- `db/orm/projects.sql.go` (generated) — regenerated, not committed
- `cmd/add.go` — the early return
- `cmd/add_test.go` — the test

The check runs before the prompt, not after. The scope's Decision is explicit that there is no
re-prompting, and asserting on exact stdout is what holds that: a prompt printed before the
`already added` line would show up in the comparison.

## TDD cycle

1. **Write test (RED):** `cmd/add_test.go`
   - [ ] `TestAddCmd_SkipsAPathAlreadyEnrolled`
     - **Behavior:** enrollment is idempotent, so `bp add .` is safe to put in a setup script or run
       twice by hand — and it says so plainly rather than failing with a constraint error the
       operator has to interpret.
     - **Setup:** the same `home`/`XDG_DATA_HOME`/`initProject(t, "BIT")` preamble as
       `TestAddCmd_EnrollsUsingTheBitPrefix`. Run `runWithStdin(t, "\n", addCmdUse, ".")` once and
       require it to succeed, then run it a second time and capture that output.
     - **Assertions:** the second run's `err` is nil and its `out == "already added\n"` — exactly
       that, so the absence of a prompt is part of the assertion. Then `db.Open()` and
       `ListProjects` has length 1.
     - **Boundary:** the `projects.path` UNIQUE constraint at its collision point — one existing row
       against the same path, versus BIT-29.4's zero.
   - [ ] Confirm fails: the second run returns a non-nil error whose message contains
         `UNIQUE constraint failed: projects.path`, and `out` is the prompt with no `already added`.

2. **Implement (GREEN):**
   - [ ] `db/queries/projects.sql`: add
         `-- name: ProjectExists :one` / `SELECT EXISTS(SELECT 1 FROM projects WHERE path = ?);`
   - [ ] `sqlc generate` — or just let `just test` do it, since it depends on `db-gen-queries`.
         Nothing to commit: `db/orm/` is gitignored. Verified signature:
         `orm.ProjectExists(ctx context.Context, path string) (bool, error)` — sqlc scans the `EXISTS`
         result straight into a `bool`.
   - [ ] `cmd/add.go`: move `db.Open()` above the prompt, and between resolving `abs` and prompting,
         call `ProjectExists`; when true, print `already added` and `return nil`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. Verse 2's whole-slice check lands on BIT-29.8.

## Commit (user)
`feat(add): report an already-enrolled path instead of failing`