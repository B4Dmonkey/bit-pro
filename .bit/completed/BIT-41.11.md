---
id: BIT-41.11
title: The installed version comes from this project's record
status: done
approved: true
phase: 5
phase_label: stale notice
---
## **Verse 5**

The first of the two local reads: which version of the bit plugin is installed *for this project*.
`~/.claude/plugins/installed_plugins.json` records one entry per install, and the real file on this
machine holds two `bit@bit-pro` entries at different `projectPath`s — so picking the right one is
the whole job, not an afterthought. Reading a file is what keeps the hot path free and offline,
per the never-block decision.

Every failure is silence, never an error: the notice is advisory, and a missing or unreadable
record must not change what `bp` does.

## Scope
- `claude/plugin.go` (new) — `InstalledVersion(home, projectRoot string) (string, bool)`.
- `claude/plugin_test.go` (new) — fixtures written into `t.TempDir()`, matching the inline-JSON
  style of `claude/settings_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestInstalledVersion` (table-driven)
     - **Behavior:** the version reported is the one recorded for the given project root, and every
       shape that cannot answer that question reports "not known" rather than an error or a guess.
     - **Setup:** a temp home containing
       `.claude/plugins/installed_plugins.json` with a `plugins` object whose `bit@bit-pro` key
       holds two records — `{"scope":"project","projectPath":"/p/a","version":"0.1.0"}` and
       `{"scope":"project","projectPath":"/p/b","version":"0.2.0"}` — plus an unrelated
       `go@go-skills` key, mirroring the real file's shape. Separate temp homes for the negative
       rows: one with no file at all, one whose file is `{`, one whose `bit@bit-pro` record has no
       `projectPath` (a user-scope install), and one with an empty `plugins` object.
     - **Assertions:** `/p/a` → `("0.1.0", true)`; `/p/b` → `("0.2.0", true)`; `/p/c` →
       `("", false)`; and `("", false)` for each of the four negative homes.
     - **Boundary:** the `projectPath` match across its three states — this project, another
       project, no matching project (the scope's "absent → silence" decision); plus the file's own
       range: absent, malformed, present-but-empty, and present-but-not-project-scoped.
   - [ ] Confirm fails: `undefined: claude.InstalledVersion`.

2. **Implement (GREEN):**
   - [ ] `InstalledVersion(home, projectRoot string) (string, bool)`: read
         `filepath.Join(home, ".claude", "plugins", "installed_plugins.json")`; on any read or
         unmarshal error return `("", false)`. Decode into a struct with
         `Plugins map[string][]struct{ ProjectPath, Version string }`; return the first record under
         `bit@bit-pro` whose `ProjectPath` equals `filepath.Clean(projectRoot)` — comparing cleaned
         paths on both sides — and `("", false)` when none matches.
   - [ ] Name the plugin key as a package const alongside the existing `bit-pro` marketplace
         literal in `claude/settings.go`, so the two places that spell `bit@bit-pro` agree.

## Claude verifies
- [ ] `just test` passes.
- [ ] `just lint` passes.
- [ ] Sanity read against the real file, which is what the fixture is modelled on:
      `bp` is not involved — `python3 -c "import json;print(json.load(open('$HOME/.claude/plugins/installed_plugins.json'))['plugins']['bit@bit-pro'])"`
      shows the two project-scope records the fixture mirrors.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(version): read the installed plugin version for this project`