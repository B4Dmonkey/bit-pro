---
id: BIT-16
title: Ship the bit_* skills as a Claude Code plugin
status: todo
---
## Why

Editing one line of a skill costs a Go rebuild and a re-seed. The skills are `//go:embed`-ed
into the binary (`assets/assets.go`), so `bit init` writes `.claude/skills/` from the *embedded*
copy — meaning the loop is edit `assets/skills/<skill>/SKILL.md`, `just install`, `bit init`.
Miss a step and the tool ships a stale skill without saying so.

The bigger cost is distribution. The pipeline is meant to be used across many repos and several
clients, and today there is no versioned way to get it there: every project gets whatever
happened to be compiled into whatever binary was installed at the time. There is no way to say
"this project is on v2 of the skills," no way to update one repo without rebuilding, and no way
for someone who clones a repo to discover the pipeline exists at all.

Claude Code's plugin system is built for exactly this — versioned, marketplace-distributed
skills — and this repo already consumes one (`go@go-skills` and friends, enabled in
`.claude/settings.json`). The skills should ship the same way.

## Summary

Move the four `bit_*` skills and the `bit-cli.md` contract doc out of the binary and into a
Claude Code plugin that lives in this repo, with the repo doubling as its own plugin
marketplace. `bit init` stops writing skills entirely; its job narrows to setting the task-ID
prefix, scaffolding `.bit/`, and writing a managed block into the project's `CLAUDE.md`. Once
the plugin is the only source of skills, `assets/` and the `//go:embed` are deleted.

The immediate win is the authoring loop: with the plugin loaded from disk, editing a skill and
running `/reload-plugins` picks up the change with no rebuild and no re-init. The durable win is
that a skill version can be installed into any repo from GitHub.

## Visual aid

```
today                                    after

assets/skills/*/SKILL.md                 <plugin>/skills/*/SKILL.md
        │ //go:embed                             │
        ▼                                        │ /reload-plugins  (dev, in-repo)
   bit binary                                    │ marketplace       (other repos)
        │ bit init (Seed)                        ▼
        ▼                                  loaded by Claude Code
 .claude/skills/*/SKILL.md
                                         bit init: prefix + .bit/ + CLAUDE.md block
 edit → just install → bit init           edit → /reload-plugins
```

## Risks & unknowns

- **Unknown:** How many plugins, and what the skills end up being called. Plugin skills in a
  `skills/` directory are namespaced as `/<plugin>:<skill>`, so one plugin named `bit` yields
  `/bit:bit_scope`. A plugin shipping exactly *one* skill can instead put `SKILL.md` at the
  plugin root and takes its invocation name from frontmatter — which is how `go@go-skills` is
  invoked as plain `go`. So four single-skill plugins would preserve today's `/bit_scope`
  exactly, at the cost of four plugins to enable and four copies of `bit-cli.md`.
  **Resolve by:** user's call on the naming. One plugin with `skills/scope|plan|do|check` reads
  best (`/bit:scope`, `/bit:plan`) and keeps one copy of the contract doc, but it renames every
  invocation the user's muscle memory depends on.
  **De-risk before planning?** Yes — it determines the directory layout every other verse
  writes into, and it is a naming choice, not something building can answer.

- **Unknown:** How a skill reads `bit-cli.md` once it's inside a plugin. Plugins are copied into
  a cache (`~/.claude/plugins/cache`), so paths pointing outside the plugin directory break.
  `${CLAUDE_PLUGIN_ROOT}` is the documented substitution for plugin-bundled paths, but it's
  documented for hook, MCP, and monitor *commands* — not for prose in a `SKILL.md` instructing
  the agent to go read a file.
  **Resolve by:** Verse 1 proves it by building — put `bit-cli.md` in the plugin, have a skill
  read it, confirm the path resolves from the cache directory.
  **De-risk before planning?** No — Verse 1 is the walking skeleton and this is the specific
  thing it proves. If the substitution doesn't work in skill prose, the fallback is a copy of
  the doc inside each skill directory, which is a known-good layout.

- **Unknown:** What `bit init` can actually assert about the plugin, given the version is a git
  commit SHA during active development rather than a semver string. There's nothing to compare
  against, so a compatibility *assertion* may not be expressible — it may only be able to report
  presence.
  **Resolve by:** decide during Verse 4 what init checks and what it says when the check fails.
  **De-risk before planning?** No — it only affects Verse 4's wording, and it depends on what
  Verse 3 shows about how a fresh clone actually behaves.

- **Risk:** `BIT-15` (rename `bit` → `bp`) is in flight and `BIT-15.2` rewrites all five files
  under `assets/` — the exact files this scope deletes. Doing them in the wrong order means
  editing files to throw them away. **Resolve by:** a sequencing call before planning — either
  finish `BIT-15.2` and accept the wasted edit, or drop `BIT-15.2` and fold the `bit` → `bp`
  rename into this scope's port, since the skill text has to be touched during the move anyway.
  **De-risk before planning?** Yes — cheap to decide, expensive to get wrong.

## Decisions

- **The plugin becomes the only source of skills; `assets/` and the `//go:embed` are deleted.**
  Two copies of the same `SKILL.md` would drift, and the embedded copy is the one nobody can
  update without a rebuild.
- **`bit init` stops writing skills.** It keeps three jobs: set the task-ID prefix, scaffold
  `.bit/`, and write a managed (delimited, idempotent) block into `CLAUDE.md` telling a fresh
  agent the pipeline exists and how to orient.
- **The plugin source stays in this repo, and the repo is its own marketplace.** A
  `.claude-plugin/marketplace.json` at the repo root lists the plugin with a relative `source`
  path, which is the layout `spf13/go-skills` uses. Keeping it here means a CLI flag change and
  the skill that calls it land in the same commit.
- **v1 ports the existing skill content unchanged.** The four `bit_*` skills plus `bit-cli.md`,
  same text. Improving the skills is separate work. Note that "unchanged" cannot include the
  invocation name — plugin skills are namespaced, so at minimum how a skill is called changes.
- **Live reload works; the dev loop is edit → `/reload-plugins`.** `/reload-plugins` reloads
  plugins, skills, agents, and hooks without restarting the session, and
  `claude --plugin-dir ./<plugin>` loads the in-repo copy directly — a local `--plugin-dir` copy
  takes precedence over an installed one of the same name, so this repo can develop against its
  own plugin while other repos run the released version. This replaces `just install` +
  `bit init` for skill edits.
- **Omit `version` from the manifest while the skills are under active development.** With no
  `version` in `plugin.json` or the marketplace entry, the version resolves to the git commit
  SHA and every pushed commit is an available update. Setting an explicit `version` means users
  get nothing until it's bumped, which is wrong for something iterated on this often. Switch to
  explicit semver when the skills stabilize.
- **A cloned repo carries the plugin reference in checked-in project settings.**
  `extraKnownMarketplaces` (marketplace name → `{"source": {"source": "github", "repo": "..."}}`)
  and `enabledPlugins` in `.claude/settings.json` are both project-scoped and committed. This
  repo already uses `enabledPlugins` this way for the Go skills.
- **Enabling is not installing, and `bit init` cannot paper over that.** Project settings that
  enable a plugin from an external source do not install it; Claude Code reports the plugin as
  not installed and prints the `claude plugin install` command to run. So the honest target is
  that a fresh clone gets an accurate, actionable message — not a silent guarantee.
- **A `bit`-side health/status command is out of scope.** Claude Code already reports the plugin
  half of the half-installed state, including the command to fix it. The only gap left is a
  missing `.bit/`, which `bit init` fixes and which every `bit task` command already fails
  loudly about. Revisit if Verse 3 shows the fresh-clone experience is actually confusing.
- **Non-goals: hooks, and the per-repo QA verb config.** The TOML mapping canonical verbs
  (`test`, `test-one`, `lint`, `fmt`, `build`) to project commands is real future work that hooks
  depend on, but v1 is skills-only.

## Verses

- [ ] Verse 1 — **A skill edit takes effect without rebuilding `bit`**: the plugin exists in the
  repo with the four skills and `bit-cli.md` ported over, loads via `claude --plugin-dir`, and an
  edit to a `SKILL.md` shows up after `/reload-plugins`. Proves the reload loop and, critically,
  that a skill can still read the contract doc from inside a plugin.
  Touches: new plugin directory at the repo root (`.claude-plugin/plugin.json`, `skills/`),
  content copied from `assets/`.
- [ ] Verse 2 — **The pipeline installs into a different repo from GitHub**: the repo publishes
  itself as a marketplace, so `/plugin marketplace add B4Dmonkey/bit-pro` followed by an install
  puts the skills in an unrelated project with no `bit` rebuild. This is the actual point of the
  work — distribution.
  Touches: `.claude-plugin/marketplace.json` at the repo root.
- [ ] Verse 3 — **Cloning a bit-managed repo points you at the pipeline**: checked-in project
  settings name the marketplace and enable the plugin, so a fresh clone either loads the skills
  or says exactly what to run.
  Touches: `.claude/settings.json`.
- [ ] Verse 4 — **`bit init` no longer ships skills, so there is one source of truth**: init sets
  the prefix, scaffolds `.bit/`, and writes the managed `CLAUDE.md` block; `assets/` and the
  `//go:embed` are gone, so it's no longer possible to be running a stale embedded copy.
  Touches: `cmd/init.go`, `cmd/init_test.go`, deletion of `assets/`.