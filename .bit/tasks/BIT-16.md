---
id: BIT-16
title: Ship the bit_* skills as a Claude Code plugin
status: doing
---
## Why

Editing one line of a skill costs a Go rebuild and a re-seed. The skills are `//go:embed`-ed
into the binary (`assets/assets.go`), so `bp init` writes `.claude/skills/` from the *embedded*
copy — meaning the loop is edit `assets/skills/<skill>/SKILL.md`, `just install`, `bp init`.
Miss a step and the tool ships a stale skill without saying so.

The bigger cost is distribution. The pipeline is meant to be used across many repos and several
clients, and today there is no versioned way to get it there: every project gets whatever
happened to be compiled into whatever binary was installed at the time. There is no way to say
"this project is on v2 of the skills," and no way to fix a skill in one repo without rebuilding
the binary and re-running init everywhere.

Claude Code's plugin system is built for exactly this — versioned, marketplace-distributed
skills — and this repo already consumes one (`go@go-skills` and friends, enabled in
`.claude/settings.json`). The skills should ship the same way.

## Summary

Move the four `bit_*` skills out of the binary and into a single Claude Code plugin named `bit`,
living in this repo, with the repo doubling as its own plugin marketplace. The `bit-cli.md`
contract doc stops being a file at all: `bp` grows an `instructions` subcommand that prints the
contract, and each skill runs it instead of reading a path. `bp init` stops writing skills; its
job narrows to setting the task-ID prefix, scaffolding `.bit/`, and wiring the plugin into
project settings. With the plugin as the only source of skills, `assets/` and the `//go:embed`
are deleted.

The immediate win is the authoring loop: editing a skill and running `/reload-plugins` picks up
the change with no rebuild and no re-init. The durable win is that the skills release on their
own cadence — shipping a fix to a consuming repo becomes `git push` plus one update command
there, rather than rebuilding `bp`, reinstalling it, and re-running `init` in every repo.

## Visual aid

```
today                                    after

assets/skills/*/SKILL.md                 bit/skills/{scope,plan,do,check}/SKILL.md
        │ //go:embed                             │
        ▼                                        ├─ --plugin-dir + /reload-plugins   (dev, in-repo)
   bp binary                                     └─ marketplace → cache/<sha>        (other repos)
        │ bp init (Seed)                                 │
        ▼                                                ▼
 .claude/skills/*/SKILL.md                         loaded by Claude Code

 edit → just install → bp init            edit → /reload-plugins
                                          bp init: prefix + .bit/ + plugin wiring
```

## Risks & unknowns

- **Unknown:** Whether pushing a skill change actually delivers it to a repo that has the plugin
  installed. This is the mechanism the whole scope rests on, and it is currently believed rather
  than observed: the update commands (`claude plugin marketplace update bit-pro`, then `claude
  plugin update bit@bit-pro`) and SHA-based version resolution are documented, but no push has
  been watched arriving. Separately, whether it can arrive *unprompted* is near-certainly no by
  default — the docs state auto-update is disabled by default for third-party marketplaces — and
  `"autoUpdate": true` is documented only for managed admin settings, so treat it as unverified.
  **Resolve by:** Verse 1 (spike) — a plugin holding one trivial skill, installed for real, then
  edited and pushed. *Yes* looks like `installed_plugins.json` showing a new `gitCommitSha` and
  `lastUpdated` for `bit@bit-pro`, and a session loading the edited text. *No* looks like the
  update commands reporting nothing available, or the cache still serving the old SHA.
  Deterministic either way — the command path is synchronous, so nothing waits on a background
  timer.
  **Downstream:** Verses 2, 3, and 4 all depend on the answer. Each assumes the plugin is how a
  skill reaches a repo; if a push can't reach an installed plugin, the distribution story changes
  and their shape changes with it.
  **Artifact:** kept. The plugin manifest, the marketplace manifest, and the repo-as-marketplace
  wiring are Verse 2's real layout, not scaffolding — only the trivial skill itself gets replaced.
  **De-risk before planning?** Yes, and Verse 1 *is* the de-risking. Nothing else is worth
  building until a push demonstrably reaches an installed plugin.

## Decisions

- **The plugin is the only source of skills.** `assets/` and the `//go:embed` are deleted rather
  than kept as a fallback: two copies of a `SKILL.md` drift, and embedding the plugin tree — which
  does work, since a `directory`-source marketplace needs no network — keeps the skills' version
  welded to the binary, which is the coupling this scope exists to remove.
- **One plugin named `bit`, holding four skills under `skills/`.** Invocations become `/bit:scope`,
  `/bit:plan`, `/bit:do`, `/bit:check`; plugin skills under `skills/` are namespaced
  `/<plugin>:<skill>`, so the skill directories drop their now-redundant `bit_` prefix. One plugin
  means one thing to enable. The cost is that today's `/bit_scope` muscle memory changes.
- **The plugin source stays in this repo, and the repo is its own marketplace.** A
  `.claude-plugin/marketplace.json` at the repo root lists the `bit` plugin with a relative
  `source` of `./bit` — the layout `spf13/go-skills` uses. `bit-pro` is its own git repo, so a
  plain `github` marketplace source works. Keeping the source here means a CLI flag change and
  the skill that calls it land in the same commit. The repo being private is not a constraint on
  distribution: the marketplace fetch authenticates with the account's own git credentials, and
  every consuming machine is that same account. The repo goes public later, which only widens it.
- **The CLI contract is printed by `bp instructions`, not shipped as a file.** Each skill's single
  "read `.claude/bit-cli.md`" line becomes "run `bp instructions`". That removes the only
  cross-file reference in the skill set, so nothing has to resolve a path out of the plugin cache,
  and it closes version skew: a skill asks the binary for its own contract, so the two cannot
  disagree about verb and flag names. Rejected: a bundled copy of the doc per skill directory
  (four checked-in copies that drift from each other and from the binary), and having init keep
  seeding the file (keeps skill-facing content in the binary, so the embed can never fully die).
- **The delivery mechanism is proven before anything is ported.** Verse 1 ships a trivial skill
  purely to watch a push arrive, because the port is wasted effort if pushes don't propagate, and
  a trivial payload makes a failure unambiguous — nothing else is in the frame to blame.
- **v1 ports the skill text unchanged** — two edits only: the directories drop `bit_`, and the
  contract pointer line changes. `BIT-15` already renamed `bit` → `bp` throughout `assets/`, so
  the text moving across is current. Improving the skills is separate work.
- **`bp init` has three jobs:** set the task-ID prefix, scaffold `.bit/`, and wire + install the
  plugin. It no longer writes skills.
- **Init writes the wiring *and* runs the install, because wiring alone installs nothing.** The
  wiring is project-scoped JSON carrying no skill content, so a binary can write it without
  coupling versions: `extraKnownMarketplaces` (an object keyed by marketplace name, each with a
  nested `source` object — `{"source": "github", "repo": "B4Dmonkey/bit-pro"}`) plus
  `enabledPlugins` (`"bit@bit-pro": true`). But those two keys only *register* the catalog and
  *enable* the plugin; neither fetches a file. Installing is a third step that copies the plugin
  into `~/.claude/plugins/cache/<marketplace>/<plugin>/<version>/`, and its only documented
  trigger is a prompt when the user trusts the repository folder. So init also runs `claude plugin
  install bit@bit-pro --scope project` and surfaces its error if it fails — `claude` is assumed
  present, since a repo using these skills is a repo running Claude Code. A written settings file
  is never to be treated as a working install.
- **The dev loop is `claude --plugin-dir ./bit` plus `/reload-plugins`.** `--plugin-dir` loads a
  directory for the session only and takes precedence over an installed plugin of the same name,
  so this repo develops against its working tree while other repos run the released version. It is
  load-bearing, not a convenience: an installed plugin is served from a cache directory keyed by
  commit SHA, so `/reload-plugins` alone would reload the cached copy and never see the edit.
- **Omit `version` from the manifest while the skills are under active development.** Without it
  the version resolves to the git commit SHA, so every pushed commit is an available update; an
  explicit `version` means nobody gets a fix until it's bumped, which is wrong for something
  iterated on this often. `claude plugin validate` warns and passes. Switch to semver when the
  skills stabilize.
- **Non-goal: a `bp` health/status command.** With init performing the install, what's left is a
  missing `.bit/` — which init fixes and which every `bp task` command already fails loudly about
  — and plugin state, which Claude Code reports on itself. Revisit if Verse 3 shows the
  fresh-clone experience is actually confusing.
- **Non-goal: a managed `CLAUDE.md` block.** Init has never written one, and an enabled plugin
  already surfaces the four skills with their descriptions attached. The part a description can't
  carry — the scope → plan → do order, and never hand-editing `.bit/tasks/*.md` — is
  always-loaded-context work that `BIT-17` already owns, where it gets written against evidence
  about which rules actually get broken.
- **Non-goal: an MCP server, though nothing here forecloses one.** A plugin can ship `.mcp.json`
  beside `skills/`, so `bp mcp` as a stdio subprocess would be a strict addition to this
  packaging — same plugin, same install. It waits because it would not replace the skills (MCP
  carries tools, not methodology), it would be a third API over `task/` beside the CLI and the
  TUI, it costs context on every turn where a skill costs nothing until invoked, and it breaks the
  `TRACK=$(bp task create …)` shell-capture idiom `bit_plan` uses to create many bars at once. Its
  one real win is schema validation: an enum makes a silent `doen`-style status typo
  unrepresentable. Revisit when `BIT-17`'s recurring-cause counts show contract drift or status
  typos actually repeating.
- **Non-goals: hooks, and the per-repo QA verb config.** The TOML mapping canonical verbs (`test`,
  `test-one`, `lint`, `fmt`, `build`) to project commands is real future work that hooks depend
  on, but v1 is skills-only.

## Verses

- [x] Verse 1 (spike) — **Does pushing a skill change deliver it to an installed plugin?**: this repo carries a
  `bit` plugin holding one trivial skill and publishes itself as a marketplace, so the plugin can
  be installed for real — and then editing that skill, pushing, and updating shows the new version
  arriving. The payload is deliberately trivial because the mechanism is what's under test; the
  existing `bit_*` skills keep working from `.claude/skills/` throughout, so nothing is at risk
  while this is proven. If this fails, the rest of the scope does not hold.
  Touches: new `bit/` directory at the repo root (`.claude-plugin/plugin.json`, `skills/`), and
  `.claude-plugin/marketplace.json` at the repo root.
- [ ] Verse 2 — **The real pipeline ships through the plugin**: the four skills move into the
  plugin and are invoked as `/bit:scope` and friends, learning the CLI contract by asking `bp` for
  it rather than reading a path. Editing one and running `/reload-plugins` under `claude
  --plugin-dir ./bit` takes effect with no rebuild and no re-init.
  Touches: `bit/skills/`, content copied from `assets/`; a new `instructions` command under
  `cmd/`.
- [ ] Verse 3 — **`bp init` leaves a repo ready to use the pipeline**: init writes the marketplace
  and enabled-plugin entries into checked-in project settings and installs the plugin, so one
  command in a fresh repo — or a fresh clone of one — ends with the skills available.
  Touches: `cmd/init.go`, `.claude/settings.json`.
- [ ] Verse 4 — **There is one source of truth for the skills**: `assets/` and the `//go:embed`
  are gone, so init can no longer ship a stale embedded copy and there is nowhere for a second
  version of a `SKILL.md` to hide.
  Touches: `cmd/init.go`, `cmd/init_test.go`, deletion of `assets/`.

## References

- `https://code.claude.com/docs/en/plugins` — plugin structure (which directories sit at the
  plugin root vs inside `.claude-plugin/`), skill namespacing, `--plugin-dir`, and that
  `.mcp.json` sits beside `skills/`. Informs Verses 1 and 2.
- `https://code.claude.com/docs/en/skills` — the substitution table (`${CLAUDE_SKILL_DIR}` is the
  skill's own subdirectory, not the plugin root) and the supporting-files pattern for shipping a
  doc alongside a skill. Informs Verse 2.
- `https://code.claude.com/docs/en/plugin-marketplaces` — marketplace manifest, source types,
  version resolution, `extraKnownMarketplaces`, and the auto-update defaults. Informs Verses 1
  and 3.
- `~/.claude/plugins/marketplaces/go-skills/` — a working single-repo-as-marketplace on disk; the
  layout exemplar. Informs Verse 1.