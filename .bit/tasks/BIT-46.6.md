---
id: BIT-46.6
title: A task store built with a project code mints tracks without config.toml
status: todo
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

`task.Store` is only `{root}` (`task/store.go:25-31`) and `Create` reads the prefix from `config.toml` (`:216-221`). The track drops `config.toml`, so the store has to take the code from the registry. This bar adds the constructor and keeps a transitional fallback so every existing caller stays green. The fallback is deleted with `config.go` in the last bar of this verse.

## Scope
- `task/store.go`:
  - `type Store struct { root, code string }`.
  - `func NewProject(root, code string) *Store` — **seam**: the only way production code builds a store from now on (`project.OpenStore` uses it next). `New(root)` stays as `NewProject(root, "")` for tests and stores that never mint tracks.
  - `Create`, top-level branch: `prefix := s.code`; if empty, fall back to `s.Config()` as today; then `s.NextID(prefix)`.
- `task/store_test.go`.

## TDD cycle

0. **Convert existing tests (pure restructure, same cases and assertions, still green):**
   - [ ] `task/store_test.go`, one top-level test per `Store` method, each flat test becoming a subtest named by its suffix in lowercase words; existing tables keep their rows:
     - `TestStorePath/contains untrusted id`
     - `TestStoreRelocate/{moves file out of list, cascades to bars, refuses with unfinished bars, force overrides guard, contains untrusted id, drops bar from parent order, leaves legacy order unmaterialized}`
     - `TestStoreNextID`: its existing rows, plus subtests `reserves archived ids`, `reserves completed ids`
     - `TestStoreNextChildID/{errors when parent missing, mints when parent exists, reserves archived children, reserves completed children}`
     - `TestStoreLoad/{round trips a saved task, errors on unknown id}` (`TestStoreSaveLoad_RoundTrips` becomes `round trips a saved task`)
     - `TestStoreList/{empty when no tasks dir, orders bars by explicit order}`
     - `TestStoreMove/{resequences, rejects, rejects anchor pair}`
     - `TestStoreConfig/{round trips, errors when absent}`
     - `TestStoreCreate`: its existing rows, plus subtest `rejects unknown parent`
     - `TestStoreUpdate/{applies only set fields, approval revocation}` (`approval revocation` keeps its rows as a table)
     - `TestCompareIDs` is already in shape.

1. **Write test (RED):**
   - [ ] `TestStoreCreate/mints from the project code` (subtest in `TestStoreCreate`)
     - **Behavior:** a store that knows its project's code mints track IDs from it, with no `config.toml` on disk.
     - **Setup:** `s := NewProject(t.TempDir(), "EX")`; `s.Create(CreateParams{Title: "First track"})`.
     - **Assertions:** `t.ID == "EX-1"`; `<root>/config.toml` does not exist.
     - **Boundary:** no `config.toml` at all — the case that errors today.
   - [ ] Confirm fails: `NewProject` undefined.

2. **Implement (GREEN):**
   - [ ] `NewProject` and the `code` field; `Create` uses the code first.

3. **More tests (RED → GREEN):**
   - [ ] Turn `TestStoreCreate/mints from the project code` into a table inside that subtest: `"EX"` → `EX-1`, and a second create in the same store → `EX-2`; a different store with `"ZZ"` → `ZZ-1`. *Boundary:* two distinct codes, so a hardcoded prefix can't pass; count 1 → 2 for minting.
   - [ ] `TestStoreCreate`'s existing rows (they seed `SaveConfig`, `:739-822`) stay green — the fallback path.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): task store takes its project code`