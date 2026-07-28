---
id: BIT-15
title: Rename binary from bit to bp
status: doing
---
## Why
The `bit` binary name collides with other tools in the dev environment. `bp` (bit-pro) is short, unique, and unambiguous — so `bp tui`, `bp task list`, etc. work without shadowing anything else.

## Summary
Rename the installed binary from `bit` to `bp`. The Justfile build step, the Cobra root command's `Use:` string, and the embedded skill assets (which call `bit task …`) all update in lockstep. The `.bit/` project directory and `BIT-` task ID prefix are unchanged — this is a binary name change only.

## Decisions
- **Binary name is `bp`.** Matches what the user said.
- **`.bit/` directory and `BIT-` task ID prefix are unchanged.** The rename is binary-only; project data stays the same.
- **Skill assets (`assets/skills/`, `assets/bit-cli.md`) are the source of truth.** Changes go there first, then `just install` + `bp init` re-seeds `.claude/`.

## Verses
- [x] Verse 1 — Binary runs as `bp`: `just install` builds `bp` into `$GOBIN`; `bp tui`, `bp task list`, etc. all work; tests pass against the new name.
  Touches: `Justfile` (build -o flag), `cmd/root.go` (cobra `Use:`), `cmd/root_test.go` (string assertions).
- [ ] Verse 2 — Skill assets call `bp`: all four skill SKILL.md files and `bit-cli.md` in `assets/` updated from `bit task …` to `bp task …`; binary reinstalled; `bp init` re-seeds `.claude/` with the updated contract.
  Touches: `assets/bit-cli.md`, `assets/skills/bit_scope/SKILL.md`, `assets/skills/bit_plan/SKILL.md`, `assets/skills/bit_do/SKILL.md`, `assets/skills/bit_check/SKILL.md`.