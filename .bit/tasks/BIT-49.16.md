---
id: BIT-49.16
title: migrate stages and verifies the copy, so a mismatch leaves nothing behind
status: todo
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

Until now, migrate writes straight into the live store. A source whose copy can't match it exactly must stop the migration with nothing written and nothing registered. One example is a hand-numbered v1 note `BIT-1-1.md` with no `BIT-1-001.md` beside it: `ImportNote` writes it as `BIT-1-001`, so the copy succeeds but its name differs from the source. That forces stage, then verify, then commit, then register last.

## Scope
- `task/store.go`:
  - `func (s *Store) LoadFrom(p Place, id string) (*Task, error)`, the read twin of `SaveTo`. `Load(id)` becomes `LoadFrom(Active, id)`.
  - `func (s *Store) IDs(p Place) ([]string, error)`: the stems of `*.json` in `placeDir(p)`, sorted. A missing dir gives an empty list.
- `migrate/migrate.go`:
  - **Stage:** `data, err := store.Dir()`; `os.MkdirAll(data, 0o755)` (a first migration may run before the data dir exists); `stage, err := os.MkdirTemp(data, ".migrate-")`, which is on the same filesystem, and a leading `.` can never be a project code. Each error is returned. `defer os.RemoveAll(stage)`. The staging store is `task.NewProject(filepath.Join(stage, code), code).WithDataRoot(stage)`, and every write so far goes to it.
  - **Verify** (new unexported `verify(src, s, ...) []string`, returning one problem line per mismatch as a path relative to `.bit/`):
    - Tasks, per place: `IDs(place)` equals the set of source file stems. For each, `LoadFrom` equals the parsed source on `ID, Title, Status, Approved, Phase, PhaseLabel, Order, Body` (bytes); `Branch` and `Commit` equal the session head's branch and SHA (the source has neither); and `Project == code`.
    - Feedback: `ListNotes("")` equals the source stems, and each `ReadNote` is byte-equal to the source.
    - Research: per track dir, `ResearchTopics` equals the source stems, and each `ReadResearch` is byte-equal.
    - Retro: `ListRetro()` names equal `retroName(code, stem)` for each source, and each `ReadRetro` body is byte-equal. Export the rule as `task.RetroName` so migrate computes the expected name the same way.
  - Any problem → `fmt.Errorf("%w:\n  %s", ErrVerify, strings.Join(problems, "\n  "))` with `var ErrVerify = errors.New("migrated copy doesn't match the source")`.
  - **Commit:** move each file in `stage/feedback/` to `data/feedback/`, and `stage/retro/` to `data/retro/`, with `os.Rename` (`MkdirAll` first). `os.Rename` replaces an existing file, so each destination is `os.Lstat`ed first and an existing one stops the commit with an error naming it. Then `os.Rename(stage/<code>, store.ProjectDir(code))`. Then `CreateProject`, last of all. Destination names come from `os.ReadDir` and go through `pathologize.Join` (fileflow-pathologize). `fileflow.Move` isn't used, because it would silently suffix a clash (BIT-46.12's reasoning).
- Test files touched: `cmd/migrate_test.go` and `migrate/migrate_test.go`, both already in shape.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/a copy that doesn't match leaves nothing behind`
     - **Behavior:** a migration either lands whole and verified, or changes nothing at all.
     - **Setup:** sandbox. A v1 store with `tasks/BIT-1.md` and a hand-numbered `feedback/BIT-1-1.md` (no `BIT-1-001.md`, so `ImportNote` doesn't clash and the mismatch reaches verify). Then `bp migrate`.
     - **Assertions:**
       - `errors.Is(err, migrate.ErrVerify)`, and the error text contains `feedback/BIT-1-1.md`.
       - `ListProjects` is empty.
       - `<data>/bit/BIT` doesn't exist, and `<data>/bit/feedback` has no `BIT-*` file.
       - `<data>/bit` has no `.migrate-*` entry.
       - The source is unchanged.
     - **Boundary:** a source the store can't reproduce byte for byte.
   - [ ] Confirm fails: the migration succeeds, registers `BIT` and leaves `feedback/BIT-1-001.md` in the live store.

2. **Implement (GREEN):**
   - [ ] `LoadFrom`, `IDs`, `RetroName`, plus staging, `verify` and the commit sequence.

3. **More tests (RED → GREEN):**
   - [ ] Every earlier `TestMigrateCmd` and `TestRun` case still passes through staging. *Boundary:* the happy path is unchanged.
   - [ ] `TestRun/registers only after the files are in place`: on success, `<data>/bit/BIT/tasks/BIT-1.json` exists, the row exists, and no `.migrate-*` remains. *Boundary:* the commit order.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate stages and verifies before it commits and registers`