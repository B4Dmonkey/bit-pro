---
id: BIT-1
title: CLI Bootstrap
status: done
---
# CLI Bootstrap

## Why

`bit` doesn't exist yet as a runnable program (see `README.md` for the full vision).
Before any of the real product decisions (storage structure, epic/task model, kanban
board, TUI) can be scoped and built, there needs to be a Cobra project that actually
builds, runs, and can be pointed at a project to manage. This scope covers only that
bootstrapping step — it ends the moment there's a real, runnable command — not the
features that will eventually live inside it.

## Summary

Stand up a Go module with Cobra wired in, a `Justfile` as the task runner for common
dev commands (build/run/test), and a `bit init` command that creates the `.bit/`
directory future commands will read and write. No epic/task logic, storage schema, or
TUI rendering lands in this scope — bootstrapping ends once `bit init` gives us a real
on-disk target; the actual data model and commands (task CRUD, list UI, board UI) are
separate follow-up scopes (see README's roadmap).

## Phases

- [x] Phase 1 — `bit` is an installable, runnable CLI: `just build` (wrapping
  `go build`) produces a binary, and `bit --help` / `bit --version` work. This is the
  walking skeleton everything else hangs off of.
  Touches: `main.go`, `cmd/root.go`, `Justfile`
- [x] Phase 2 — a user can initialize a project for `bit` to manage: `bit init` run
  inside a repo creates a `.bit/` directory as the on-disk home future commands
  will read and write. This is the bootstrapping finish line — once this works, the
  next scope (task CRUD) builds the actual data model inside `.bit/`.
  Touches: `cmd/init.go`, new `internal/` package for shared CLI scaffolding

## Risks & unknowns

- **Unknown:** Storage *format* is decided — markdown files with YAML frontmatter (see
  README) — but the *structure* inside `.bit/` (directory layout, file naming, how
  frontmatter encodes status/relationships) isn't.
  **Resolve by:** Explore during the "task management (CRUD)" scope, once `.bit/`
  exists as a real target to build against.
  **De-risk before planning?** No — Phase 2 only needs to create the directory, not
  decide what goes inside it.

## Context
See scope: [cli-bootstrap-scope.md](./cli-bootstrap-scope.md)
Recap: `bit` doesn't exist as a runnable program yet; this plan covers both phases of the
bootstrap scope — the walking skeleton that makes `bit` installable and runnable (Phase 1),
and `bit init` creating the `.bit/` directory (Phase 2). Both are planned together because
the scope's only open risk (storage structure inside `.bit/`) explicitly doesn't block
Phase 2 — it only needs to create an empty directory, not decide what goes inside it.

## How this plan works
The entry point is the Cobra root command (`cmd.NewRootCmd()`), exercised directly in Go
tests rather than by shelling out to a built binary — that's the fastest, most direct way
to prove the command is wired correctly. Step 1 proves the root command exists and answers
`--help`. Step 2 is forced by a second test that Step 1's command can't satisfy: `--version`
requires a `Version` field Step 1 never set. Step 3 has no new Go-level behavior to test —
it wraps the already-proven behavior in the `just` task runner the scope calls for, and is
verified by actually running the built binary rather than a unit test. Step 4 (Phase 2)
adds `init` as a new subcommand on the same root command, forced by a test that a fresh
`bit` binary can't satisfy — there's no `init` subcommand registered yet.