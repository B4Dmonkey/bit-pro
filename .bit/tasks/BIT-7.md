---
id: BIT-7
title: Install bit and bootstrap any project with the skills
status: doing
---
## Why

Today bit only runs as `./bin/bit` from this repo's root, and the `bit_*` skills that
make the scope → plan → do workflow usable live inside this repo's `.claude/`. There is
no way to use bit — or its skills — to track work in another project. The whole point of
the tool is to track development next to the code it changes; until bit installs and can
seed a fresh project, it can only ever track its own development.

## Summary

Make bit a locally-installed tool (a binary on `PATH` via `just install`) and teach `bit
init` to write the `bit_*` skills and `bit-cli.md` into any project's `.claude/`, from a
single copy of those files embedded in the binary at build time. `init` is idempotent —
re-running it refreshes the skills — so there is one source of truth and staleness is
fixed by re-running, not by hand. This repo dogfoods the same path: its own
`.claude/skills/` becomes generated output of the same seeding. (This mirrors Backlog.md,
which bundles `src/guidelines/*.md` into its binary and writes them out idempotently on
`init`.)

## Phases

- [x] Phase 1 — bit runs anywhere: `just install` builds `bit` into your Go bin dir, so
  you can `cd` into any repo and run `bit init` / `bit task …` / `bit tui` by hand — no
  `./bin/` prefix, no needing to be in this repo. Touches: `justfile` (an install target),
  README install docs.

- [ ] Phase 2 — any project is bootstrapped (and refreshed) with the skills: `bit init`
  writes the `bit_*` skills + `bit-cli.md` into the target repo's `.claude/` from copies
  embedded in the binary, idempotently, and the seeded skills call `bit` (on `PATH`) not
  `./bin/bit`. This repo regenerates its own `.claude/skills/` the same way. Touches:
  `cmd/init.go`, the embedded skill sources, and `.claude/bit-cli.md` + `.claude/skills/bit_*`
  (their references switch to `bit`).

## Decided — don't re-litigate

- **Distribution is a local `just install`** into the Go bin dir — no publishing to `go
  install`/GitHub, no release/tag process. (Confirm `$(go env GOPATH)/bin` is on `PATH`.)
- **`init` always (re)writes the skills** — a plain idempotent overwrite. Version-aware
  "update only when changed / track a skill version" is deferred (see risks); for now,
  re-running `init` is the refresh mechanism.
- **One source of truth: the copy embedded in the binary.** This repo's `.claude/skills/`
  is generated output of the same seeding, so bit-pro dogfoods. Consequence: the skills +
  `bit-cli.md` reference `bit` on `PATH`, so developing bit-pro means `just install` after
  a CLI change to pick it up in the skills' workflow.
- **Only the `bit_*` skills + `bit-cli.md` are seeded** — not this project's bit-pro-specific
  `CLAUDE.md` Go-skills note, which doesn't belong in other repos.
- **Skill sources move under the Go tree** (Backlog.md-style — e.g. an embed package) so
  `go:embed` can reach them; `init` writes them out into any project's `.claude/skills/`,
  and this repo's `.claude/skills/` becomes generated output of that same step. That Go-tree
  location becomes where the skills are edited from now on. (Exact directory is a plan-time
  layout choice.)

## Risks & unknowns

- **Unknown:** Migrating this repo's live skills without breaking them. If sources move out
  of `.claude/skills/`, the repo's working skills must be regenerated in the same change so
  Claude Code still finds them.
  **Resolve by:** land the source move and the init-regeneration together; verify
  `.claude/skills/` is unchanged (or intended-changed) afterward.
  **De-risk before planning?** No — handled as an ordering constraint in the plan.

- **Unknown (deferred):** Skill version metadata — knowing when a project's seeded skills
  are stale vs. the installed `bit`. Backlog.md added exactly this.
  **Resolve by:** a later scope; idempotent overwrite is enough for now.
  **De-risk before planning?** No — explicitly out of scope for this work.

- **Unknown (minor):** A plain `go build`/`go install` reports version `dev` (the `justfile`
  injects it via ldflags).
  **Resolve by:** accept `dev` (versioning is nice-to-have) or read
  `runtime/debug.ReadBuildInfo`.
  **De-risk before planning?** No.