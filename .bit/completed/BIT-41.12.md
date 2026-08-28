---
id: BIT-41.12
title: The latest version comes from the marketplace clone
status: done
approved: true
phase: 5
phase_label: stale notice
---
## **Verse 5**

The second local read: what the latest published version is. The marketplace clone at
`~/.claude/plugins/marketplaces/bit-pro/bit/.claude-plugin/plugin.json` is a plain file on disk —
measured at 0.00s and working offline — which is why the scope's decision reads "latest" from it
rather than from `git ls-remote`, which blocked for 75 seconds against a black-holed route.

Same failure posture as BIT-41.11: any missing, unreadable or version-less file reports "not
known", and the caller stays silent.

## Scope
- `claude/plugin.go` — `LatestVersion(home string) (string, bool)`.
- `claude/plugin_test.go` — fixtures in `t.TempDir()`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestLatestVersion` (table-driven)
     - **Behavior:** the latest version is the one the marketplace clone's plugin manifest declares,
       and anything else reports "not known".
     - **Setup:** a temp home containing
       `.claude/plugins/marketplaces/bit-pro/bit/.claude-plugin/plugin.json` with
       `{"name":"bit","version":"0.2.0"}`; separate temp homes for the negative rows — no file, a
       malformed file (`{`), and a manifest with every field but `version`.
     - **Assertions:** `("0.2.0", true)` for the populated home; `("", false)` for each negative
       home.
     - **Boundary:** the manifest across its range — present with a version, present without one
       (the pre-versioning shape this project's own manifest had until BIT-41.3), malformed, and
       absent (a machine that never registered the marketplace).
   - [ ] Confirm fails: `undefined: claude.LatestVersion`.

2. **Implement (GREEN):**
   - [ ] `LatestVersion(home string) (string, bool)`: read
         `filepath.Join(home, ".claude", "plugins", "marketplaces", "bit-pro", "bit",
         ".claude-plugin", "plugin.json")`; return `("", false)` on any read or unmarshal error, or
         when the decoded `version` is empty.
   - [ ] Reuse the marketplace name const rather than repeating the `bit-pro` literal.

## Claude verifies
- [ ] `just test` passes.
- [ ] `just lint` passes.
- [ ] The real clone still matches the fixture's shape:
      `cat ~/.claude/plugins/marketplaces/bit-pro/bit/.claude-plugin/plugin.json` shows a `version`
      field.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(version): read the latest published version from the marketplace clone`