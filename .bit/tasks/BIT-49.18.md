---
id: BIT-49.18
title: migrate stops and lists task files that don't parse or don't round-trip byte for byte
status: todo
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

`task.Parse` silently drops frontmatter keys it doesn't know (yaml.v3 non-strict). So a v1 task with an extra key would lose it in the copy, and verify would never notice, because it compares only the fields Parse knows. The scope's round trip, `Parse` → `Bytes` == source, catches it. A file that doesn't parse at all (CRLF, BOM, no closing delimiter) is listed the same way.

## Scope
- `migrate/check.go`: unexported `badTaskFiles(src string) []string`. For every `tasks/`, `completed/` and `archive/tasks/` `.md`: `task.Parse`. On an error, add `<rel>: <err>`. Otherwise `t.Bytes()`, and if `!bytes.Equal(out, raw)`, add `<rel>: does not round-trip`.
- `migrate.Run`: runs after `unknownFiles`, before staging. A non-empty list → `ErrTaskFiles` (`errors.New("task files migrate can't copy exactly")`) with the list.
- Every fixture so far comes from `Task.Bytes()`, so it round-trips.
- Test files touched: `migrate/check_test.go` and `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/stops on task files it can't copy exactly`
     - **Behavior:** no task loses frontmatter in the copy without the operator being told.
     - **Setup:** sandbox. A valid store plus `tasks/BIT-2.md` with a hand-written frontmatter `priority: high` line, and `completed/BIT-3.md` with CRLF line endings. Then `bp migrate`.
     - **Assertions:** `errors.Is(err, migrate.ErrTaskFiles)`. The text names `tasks/BIT-2.md` with `does not round-trip` and `completed/BIT-3.md`. Nothing is registered or written.
     - **Boundary:** an unknown key (it parses but loses data) and an unparseable file.
   - [ ] Confirm fails: `BIT-2`'s `priority` key is dropped and the migration succeeds.

2. **Implement (GREEN):**
   - [ ] `badTaskFiles`, and the check in `Run`.

3. **More tests (RED → GREEN):**
   - [ ] `TestBadTaskFiles` (table): a `Bytes()`-generated file → none; a file missing its closing `---` → listed; a file whose keys are in a different order than `Bytes()` writes → listed as `does not round-trip`. *Boundary:* a byte difference with no data loss is still refused, which is what the scope's byte-for-byte rule says.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate refuses task files that don't round-trip`