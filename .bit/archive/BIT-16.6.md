---
id: BIT-16.6
title: Init writes the plugin wiring
status: done
phase: 3
phase_label: Init keeps it current
---
## **Verse 3**

Init's fourth job starts here with the half that touches no network: a repo ends up with project
settings that register the `bit-pro` marketplace and enable `bit@bit-pro`. A fixed document
satisfies this bar, which is exactly what the next one contradicts.

## Scope
- `claude/settings.go` — new package, new `WriteSettings(path string) error`.
- `claude/settings_test.go` — new.
- `cmd/init.go` — call it after `SaveConfig`, before the existing `assets.Seed`.
- `cmd/init_test.go` — the command-level assertion.

The logic lands in a new `claude` package rather than in `cmd/` because the `cmd/` layer should do
nothing but parse flags and call domain code — Cobra's own guidance, and it is what makes this
testable without driving a command. Domain-named, so `claude.WriteSettings` reads as what it is:
the Claude Code integration. Avoid naming the package `plugin`; that shadows a standard-library
package name for no benefit.

The exact shapes, both read off working config on this machine rather than inferred:

- `extraKnownMarketplaces` maps a marketplace name to an object holding a nested `source`:
  `{"bit-pro": {"source": {"source": "github", "repo": "B4Dmonkey/bit-pro"}}}`. The doubled
  `source` key is not a typo — the outer is the wrapper, the inner is the discriminator.
- `enabledPlugins` maps `plugin@marketplace` to a boolean: `{"bit@bit-pro": true}`.

## References
- `https://code.claude.com/docs/en/plugin-marketplaces` — the marketplace source types and the
  `extraKnownMarketplaces` shape. Worth a look if the schema disagrees with the shapes above.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestWriteSettings_CreatesFile`
     - **Behavior:** a repo with no `.claude/` at all ends up with a settings file that both names
       where the plugin comes from and says to load it, so a fresh clone needs nothing hand-written.
     - **Setup:** `dir := t.TempDir()`; `path := filepath.Join(dir, ".claude", "settings.json")` —
       neither the file nor its parent directory exists.
     - **Assertions:** `WriteSettings(path)` returns nil; the file parses as JSON;
       `extraKnownMarketplaces["bit-pro"]["source"]` equals `{"source": "github", "repo":
       "B4Dmonkey/bit-pro"}`; `enabledPlugins["bit@bit-pro"]` is `true`.
     - **Boundary:** the `.claude` directory at its lower bound — absent — so the `MkdirAll` path
       is exercised rather than assumed to have been done by something else.
   - [ ] Confirm fails: `undefined: claude.WriteSettings`.

2. **Implement (GREEN):**
   - [ ] `WriteSettings` creates the parent directory and writes the two keys. A fixed document is
     acceptable here — the merge is the next bar's job and inventing it now would be code no test
     demands.
   - [ ] `cmd/init.go` calls `claude.WriteSettings(filepath.Join(claudeDir, "settings.json"))`.

3. **More tests (RED → GREEN):**
   - [ ] `TestInitCmd_WritesPluginWiring`
     - **Behavior:** the wiring is something `bp init` does, not a separate command to remember —
       the scope's one-entry-point decision, asserted at the surface the user actually types.
     - **Setup:** `t.Chdir(t.TempDir())`; `mustRun(t, "init", "--prefix", "BIT")`.
     - **Assertions:** `.claude/settings.json` exists and contains the `bit@bit-pro` key.
     - **Boundary:** existing wiring count == 0 — a fresh repo, the case where init has to create
       rather than amend.

## Claude verifies
- [ ] `just test` — note this makes every existing `cmd` test write a settings file into its temp
      directory, which is harmless but will show up as new files if a test ever asserts on the
      contents of `.claude/`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(init): write the bit plugin wiring into project settings`