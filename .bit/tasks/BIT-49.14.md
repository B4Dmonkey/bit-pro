---
id: BIT-49.14
title: migrate stores v1 retro proposals under the same code-prefix rule as retro_write
status: todo
phase: 2
phase_label: bp migrate
---
## **Verse 2**

A v1 `.bit/retro/<track-or-album>-proposals.md` must become a shared record that learn can find. Two source names, one already prefixed and one not, force the copy through `WriteRetro`'s prefix rule instead of a raw rename.

## Scope
- `migrate/migrate.go`: for each `src/retro/*-proposals.md`, call `s.WriteRetro(strings.TrimSuffix(name, ".md"), string(raw), task.Commit{})`.
- No `task` change (BIT-49.5).
- Test file touched: `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/stores retro proposals under the prefix rule`
     - **Behavior:** proposals written in v1 show up in `retro_list` with the project's code, never doubled.
     - **Setup:** sandbox. The v1 store has `tasks/BIT-12.md`, `retro/album-proposals.md` and `retro/BIT-12-proposals.md`, each with a `## Proposal 1` body. Then `bp migrate`, then `retro_list {}` from an MCP session at `dir` in the same sandbox.
     - **Assertions:** proposals == `[{BIT-12-proposals, BIT}, {BIT-album-proposals, BIT}]`. `<data>/bit/retro/BIT-album-proposals.md` is byte-equal to `album-proposals.md`.
     - **Boundary:** with and without the code prefix (the rule the scope gives migrate).
   - [ ] Confirm fails: `retro_list` is empty.

2. **Implement (GREEN):**
   - [ ] The retro loop.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate stores v1 retro proposals as shared records`