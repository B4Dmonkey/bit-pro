# config.toml

**Checked:** Q4 — keys, readers, what could move to a projects table.

## What exists
`task/config.go`: `type Config struct { Prefix string \`toml:"prefix"\` }`. **One key.** bit-pro's file: `prefix = "BIT"`. `SaveConfig` rewrites the whole file with only `prefix` (any other key a user adds is dropped).

Readers/writers:
- `task.Store.Create` → `s.Config()` to get the prefix for a new track ID (`task/store.go:216`). Bars don't need it (derived from parent).
- `cmd/init.go` writes it (and reads it for the interactive default).
- `cmd/add.go:51` reads it as the default for the registry `code`.
- `update/normalize.sh` uppercases it.

## Duplication already present
`projects.code` (registry) and `config.toml prefix` are two sources for the same thing. `bp add` defaults code to prefix but lets the user type something else — they can diverge today. The daemon logs by `p.Code`; IDs are minted from `prefix`.

## Verdict
- `prefix` → becomes `projects.code` (the operator's `<code>` dir name). One source of truth; nothing else in config.toml.
- Nothing **must** stay per-repo in v1 terms. The only per-repo need in v2 is *which project this repo belongs to* — answered by the db (repos table) or, if repo-local discovery without registration is wanted, a tiny marker file. Unknown/decision: whether such a marker is wanted (it would be git-trackable and let a fresh clone self-identify).
- Per-repo Claude wiring (`.claude/settings.json` plugin enablement, `claude mcp add` local-scope server) is separate from config.toml and stays per-repo regardless.
