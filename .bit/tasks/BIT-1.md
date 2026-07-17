---
id: BIT-1
title: CLI Bootstrap
status: todo
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
