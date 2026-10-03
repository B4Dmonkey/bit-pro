---
id: BIT-46.2
title: Concurrent first opens of main.db migrate exactly once
status: todo
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

Several MCP servers and the CLI can open a fresh `main.db` at the same moment, and each runs `CreateAndMigrate` (`db/open.go`). The track decides migrations on open are serialized; a multi-process test forces the lock.

## Scope
- `db/open.go` — wrap `mate.CreateAndMigrate()` in an exclusive advisory file lock on `<data>/bit/main.db.lock`: `os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)`, `syscall.Flock(int(f.Fd()), syscall.LOCK_EX)`, `defer` `syscall.Flock(..., syscall.LOCK_UN)` + `f.Close()`. Release before `sql.Open`. Errors wrap with the lock path.
- Put the lock in its own small file `db/lock.go` (`func withLock(path string, fn func() error) error`). bp ships for macOS only (`scripts/install.sh`), so `syscall.Flock` (unix) is acceptable; no Windows variant (YAGNI).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestOpen/concurrent first opens migrate once` (subtest in `TestOpen`, `db/open_test.go`; the file was converted in BIT-46.1)
     - **Behavior:** N processes opening a fresh registry at once all succeed and the schema is applied once.
     - **Setup:** one `XDG_DATA_HOME := t.TempDir()`; spawn 8 child processes of the test binary at once (`exec.Command(os.Args[0], "-test.run=^TestHelperProcessOpen$")`, env `BIT_OPEN_HELPER=1`, `XDG_DATA_HOME=<same>`), started together, then `Wait` all. `TestHelperProcessOpen` returns immediately unless `BIT_OPEN_HELPER=1`; otherwise it calls `Open()`, runs `SELECT count(*) FROM projects`, and exits non-zero on any error (printing it). `TestHelperProcessOpen` stays its own top-level function: it is the re-exec entry point for the child processes, not a test of a unit.
     - **Assertions:** every child exits 0; afterwards `schema_migrations` count == 1.
     - **Boundary:** concurrent openers == 8 on a db file that doesn't exist yet — the first-open race; a pre-existing db is the trivial case.
   - [ ] Confirm fails: at least one child reports `table ... already exists` or `database is locked`. The race is timing-dependent: if RED doesn't reproduce, run the test with `-count=20`; if it still passes, record in the commit body that the test is a regression guard that couldn't be made to fail first.

2. **Implement (GREEN):**
   - [ ] `withLock` around `CreateAndMigrate` in `Open`.

## Claude verifies
- [ ] `go test ./db -run 'TestOpen/concurrent first opens migrate once' -count=20` passes
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): serialize registry migrations across processes`