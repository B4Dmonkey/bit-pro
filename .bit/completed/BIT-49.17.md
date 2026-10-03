---
id: BIT-49.17
title: migrate stops and lists unknown files under .bit/ before writing anything
status: done
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

Anything under `.bit/` that isn't a known v1 record would otherwise be dropped without a word. That includes the pre-2026-07-29 flat `archive/<ID>.md`, which the scope says counts as unknown. A store holding such files must stop with all of them listed and nothing staged.

## Scope
- `migrate/check.go` (new): unexported `unknownFiles(src string) ([]string, error)`. It walks `src` with `filepath.WalkDir`, keeps regular files only (empty dirs are ignored), and classifies each path relative to `src`. Known means:
  - `config.toml`
  - `tasks/*.md`, `completed/*.md`, `archive/tasks/*.md`
  - `feedback/*.md` whose name matches `^.+-\d+-\d+\.md$`
  - `research/<dir>/*.md`, exactly one level deep
  - `retro/*-proposals.md`

  Everything else is unknown, including `archive/<ID>.md`, dotfiles such as `.DS_Store`, and non-`.md` files.
- `migrate.Run`: `unknownFiles` runs right after the config is read, before staging. A non-empty list → `fmt.Errorf("%w:\n  %s", ErrUnknownFiles, ...)` with `var ErrUnknownFiles = errors.New(".bit/ holds files migrate doesn't know")`, sorted.
- Test files touched: `migrate/migrate_test.go`, plus `cmd/migrate_test.go` for the CLI case.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/stops on unknown files`
     - **Behavior:** the operator sees every file migrate would otherwise drop, and nothing is written.
     - **Setup:** sandbox. A valid v1 store plus `archive/BIT-3.md` (the flat layout), `notes.txt`, `research/BIT-1/diagram.png` and `.DS_Store`. Then `bp migrate`.
     - **Assertions:**
       - `errors.Is(err, migrate.ErrUnknownFiles)`, and the text lists `.DS_Store`, `archive/BIT-3.md`, `notes.txt` and `research/BIT-1/diagram.png` in that sorted order.
       - `ListProjects` is empty, and `<data>/bit` has no `BIT` dir and no `.migrate-*`.
     - **Boundary:** the flat-archive layout, a dotfile, and a non-`.md` file inside a known dir.
   - [ ] Confirm fails: the migration succeeds and drops the files.

2. **Implement (GREEN):**
   - [ ] `unknownFiles`, and the check in `Run`.

3. **More tests (RED → GREEN):**
   - [ ] `TestUnknownFiles` (table, `migrate/check_test.go`): every known shape → no entry. `research/BIT-1/sub/x.md` (too deep) → listed. `feedback/notes.md` (no seq) → listed. `retro/summary.md` (no `-proposals`) → listed. An empty `completed/` dir → no entry. *Boundary:* each known pattern's edge.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate stops on unknown files under .bit/`