---
id: BIT-41
title: Versioning for bp and the bit plugin
status: doing
---
## Why

Nothing on this machine can answer "am I running current?" — not the operator, and not the
tool. Before this track, `bp version` printed a git sha (`28418fe`) and the installed plugin
listed as `4ebbe7cd5eff`; neither string could be compared to anything. That matters because
`bp` is built from local source while the plugin installs from GitHub, so the binary can sit
ahead of the pushed plugin and the installed plugin cache can lag behind origin — they skew in
both directions, silently, and they must agree, since the skills the plugin ships are the only
thing that drives the CLI.

The exposure is not hypothetical, but it is not the shape this scope originally claimed. It was
written believing the versioning gap is what stalled the BIT-39 dispatch chain. Verse 3 measured
that and it is false — the stale plugin cache and the current one were byte-identical apart from
`plugin.json`, and both already carried the agents BIT-39 needed. That defect was a Claude Code
bug in project-scope installs, since fixed upstream. What remains is the original, narrower
problem, and it is real on its own: a plugin that only ever updates when told will sit stale
indefinitely, and today nothing tells you.

## Summary

Adopt SemVer across `bp`, the plugin, the `.bit/` format and the db schema as one lockstep
number, starting at 0.1.0, bumping minors as the track develops, and arriving at 1.0.0 when it
completes. Cutting a release stops being a hand-typed `git tag` and becomes `just` recipes that
take a bump *level*, never a version string. Then close the loop at the consumer end: `bp`
reports its own version as a number comparable to a published release, and tells the operator
when the plugin backing it is stale, naming the one command that fixes it.

## Visual aid

```
PRODUCER                                    CONSUMER (any project)
just release <level>
      |                                     bp <anything>
      +-- bumps plugin.json  <- truth             |
      +-- commits                                 +-- reads installed_plugins.json
      +-- tags -> git tag v0.1.0 (local)          |     -> installed here? which version?
                       |                          +-- reads marketplace clone plugin.json
just release-push -----+--> origin                |     -> latest seen (free, offline)
                       |                          |
              plugin.json -> ldflags              +-- fires detached marketplace refresh
                       |                          |     (never waited on; helps next run)
                 bp -v -> 0.1.0                   +-- behind  -> notice naming the fix command
                                                  +-- current -> silence
                                                  +-- absent / unreadable -> silence
```

## Decisions

- **SemVer, lockstep, one number.** One version covers `bp`, the plugin, the `.bit/` file
  format and the db schema. Chosen over CalVer because bit writes persistent state into
  every project it touches, and only SemVer carries a "this needs migrating" signal.
- **The bump level has a default reading, and the owner overrides it.** A completed track is
  a minor. Major means an existing `.bit/` or db needs migrating. Patch is a fix landing
  outside a track. This is a guideline for whoever runs the recipe, not something it checks — the
  recipes enforce direction and cleanliness (monotonic, clean tree, level-only), never whether
  the level fits the reason. The owner's judgement overrides it; reaching 1.0.0, below, is the
  first such case.
- **The recipe cuts its own baseline; no version is ever hand-typed.** There was no
  pre-existing `v*` tag, and none was created by hand. Verse 1's first real use of
  `just release` established 0.1.0 — with no prior version in plugin.json the invocation
  produces 0.1.0 regardless of the level passed, since there is nothing to bump. The monotonic
  guard passed trivially in that one case and binds on every release after it.
- **Minors while the track develops; 1.0.0 when it completes.** This track releases itself
  repeatedly rather than once, which is the point — the recipes get exercised several times
  before anything depends on them. `0.1.0` is load-bearing and named here: cut by Verse 1,
  published by Verse 2, and the baseline Verse 3 observed against. Further minors are fine as
  the work reaches publishable points; nothing downstream depends on how many.
- **1.0.0 is a one-time major, cut when this track completes.** It is the single deliberate
  override of the trigger guideline — a completed track would otherwise be a minor, and no
  migration is involved. The reason is that 1.0.0 is a claim about the versioning machinery
  itself: until this track lands, no version on this project can be compared to anything, so
  a stable number would be asserting something untrue. It happens exactly once; every release
  after it follows the guideline again, so BIT-39 lands as 1.1.0.
- **`plugin.json` is the source of truth; the git tag is the derived marker.** It is what a
  consumer sees when installed from an untagged commit, and the release recipe reads it to
  build the tag name. `.claude-plugin/marketplace.json` carries no version field, so plugin.json
  is the only file a release writes.
- **The recipe takes a level, never a version string.** `major` / `minor` / `patch` only, so
  `0.2.0 -> 0.1.0` is unrepresentable rather than merely discouraged. Backed by a monotonic
  guard: the computed next version must be strictly greater than the highest existing `v*` tag.
- **The tag is `v<version>`, and the recipe cuts it with `git tag -a`, not `claude plugin
  tag`.** That command mints `{name}--v{version}` — `bit--v0.1.0` here — with no flag to
  change the shape, and the plugin-name prefix buys nothing on a repo publishing one plugin
  while making every tag read oddly. Dropping it costs three things, each already covered:
  its dirty-tree check (the recipes' own repo-wide guard is strictly broader), its
  tag-exists check (`git tag` fails on an existing tag anyway), and its check that
  plugin.json agrees with the enclosing marketplace entry (marketplace.json carries no
  version field, so there is nothing to disagree with). `claude plugin validate bit
  --strict` stays in the recipe. The tag stays annotated, message `bit <version>`, matching
  what `claude plugin tag` produced.
- **Creating a tag and pushing it are separate recipes.** A local tag is trivially
  deletable; a pushed one is not. The irreversible step gets its own deliberate invocation.
- **Dirty means tracked-and-uncommitted; untracked files are ignored.** Measured on
  `claude plugin tag` before it was dropped — an untracked file passes, a modified tracked
  file is refused — and **its check was scoped to the plugin directory only**, so a dirty
  `cmd/root.go` sailed straight through. Under lockstep that is a hole, so the recipes own a
  repo-wide guard (`git diff-index --quiet HEAD --`) with the same tracked/untracked
  semantics. It is now the only dirty check in the path.
- **Add an `author` field to plugin.json.** Measured: with `version` present, the lone
  remaining warning is `author`, and `validate --strict` exits 1 on it. The marketplace
  already declares owner `josiah`, so this costs nothing and lets the recipe validate strictly.
- **The version notice must never block, delay, or fail `bp`.** An offline machine, a slow
  network, or a missing remote produces silence — not an error, not a hang. `claude plugin list
  --json` was measured at ~0.3s, the upper bound of what is acceptable on a session-shaped
  entry point.
- **`bp -v` prints the bare version — `0.1.0`, never `0.1.0-4-gdf72130`.** The commit
  distance and sha that `git describe` appends are build-time trivia; the number an operator
  reads has one job, which is to be comparable against a published version, and the tail only
  makes that harder. So the printed string is the version `plugin.json` declares — already
  this scope's source of truth — which means what `bp` reports and what the release declares
  cannot drift apart by construction. Accepted cost: a build between releases reports the same
  string as the release itself. That was never the question this track exists to answer, and
  git still answers it for anyone standing in the repo. Verse 4 delivers this; it also settles
  skew comparison for Verse 5, since with only base versions in play there is no describe
  string to normalise away.

- **The notice reads `bp: bit plugin <installed> → <latest> available — run: claude plugin
  update bit@bit-pro --scope project`.** One line, on stderr, with both versions substituted.
  Settled here so the plan asserts a fixed string rather than inventing one. It carries the
  three things an operator needs and nothing else: which tool is speaking, how far behind they
  are, and the exact command — including `--scope project`, without which the command fails
  outright on a project-scope install. **stderr, not stdout,** is forced rather than chosen:
  `bp instructions` guarantees `ID=$(bp task create …)` holds exactly the ID and that
  `bp task read --body` round-trips byte-for-byte, and a notice on stdout would break both.

- **The notice never fires from `bp tui` or `bp serve mcp`.** Every other command prints it.
  Those two own their output stream for the life of the process — the TUI renders full-screen,
  and the MCP server speaks a protocol over stdio — so an advisory write from a persistent
  pre-run lands in the middle of a render or a frame. The detached marketplace refresh is
  skipped alongside it, so those two commands are wholly untouched by this verse.

- **Absent is out of scope; only *behind* is detected.** The notice assumes the project is
  correctly set up — `bp init` has run and the plugin is installed. If no install record for
  this project exists, `bp` stays silent rather than guessing a fix, because the two causes
  (never initialised vs. a lost install record) need different commands and telling them
  apart is not worth the surface. This narrows what Verse 3 measured; see below.

### Settled by Verse 3 (measured 2026-08-27, recorded in BIT-41.5)

- **Updates are manual, never automatic — so the notice is obligatory.** An installed consumer
  does not pick up a pushed, tagged version on its own. Measured: with `v0.1.0` on origin, a
  re-read reported the old git sha, a fresh session start reported the old git sha, and
  `claude plugin marketplace update` alone reported the old git sha. Only
  `claude plugin update` moved it. Nothing in the system tells an operator they are behind, so
  a machine can sit stale indefinitely — this one had for five days.
- **The notice must name `claude plugin update bit@bit-pro --scope project`, with the flag.**
  Measured: without `--scope project`, the command defaults to user scope and fails outright on
  a project-scope install (`Plugin "bit" is not installed at scope user`). A notice naming the
  bare command would hand the operator an error.
- **Two conditions are distinguishable, but only *behind* is in scope.** *Behind* — installed
  here but older than latest — is what the notice detects. *Absent* — no install record for
  this project at all, which is enabled-but-not-loaded and presents as missing agents and
  skills — was also observed, and was the failure that actually cost time. It is deliberately
  excluded (see the Decision above): its real cause here was the upstream bug below, not
  anything this track can fix, and the correct remedy differs by cause.
- **"Latest" is read locally, never from the network on the hot path.** Timed, warm: the
  marketplace clone's `bit/.claude-plugin/plugin.json` and the plugin cache directory name are
  both 0.00s and work offline; `claude plugin list --json` is 0.28s and offline but reports only
  what is *installed*; `git ls-remote --tags origin 'v*'` is 0.18s warm but **blocked 75s against
  a black-holed route**. Git 2.50.1 documents no connect-timeout setting — only
  `http.lowSpeedLimit`/`http.lowSpeedTime`, which govern transfer stalls — so any network read
  must be bounded by the caller (`exec.CommandContext` with a deadline; verified to cut off at
  2.03s). Given the never-block decision, the hot path reads local files only.
- **Every `bp` run refreshes the marketplace clone, unconditionally.** The clone is the free
  offline source of "latest", but it only moves when `claude plugin marketplace update` runs, so
  left alone it would match the install forever and the *behind* notice would never fire. So
  `bp` fires the refresh itself on every invocation — detached, output discarded, never waited
  on. No timestamp, no once-a-day guard, no stamp file to own: simplicity is the point, and a
  redundant refresh costs nothing the operator can perceive. It follows from the never-block
  decision that every failure mode here is silent — offline, no remote, or two runs refreshing
  the same clone at once. The notice always compares against whatever the clone said when the
  run started; a refresh benefits the next run, never this one.

- **A versioned plugin caches under its version.** `~/.claude/plugins/cache/bit-pro/bit/0.1.0`
  replaced `.../4ebbe7cd5eff` on update, matching the pydantic-skills precedent. The directory
  name is therefore a usable signal, though `installed_plugins.json` is the authority.
- **Versioning was never what blocked BIT-39.** Measured: the stale cache and the current one
  differed only in `plugin.json`; `agents/bot.md` and `agents/bot-dev.md` were present in both.
  The real cause was upstream bug anthropics/claude-code#27257 — a project-scope install in one
  project made the same plugin unininstallable at project scope in another, leaving it enabled
  but not loaded. Fixed in Claude Code 2.1.248, confirmed by installing into `tools/example` and
  resolving `bit:bot-dev` there. `bp init` / `bp add` already issue the correct command sequence
  and need no change for this. Recorded so the Why above is not re-derived from the old belief.
- **A local plugin directory can be loaded without publishing.** `claude --plugin-dir <path>`
  loads a plugin straight from a directory for one session, bypassing marketplace, cache and
  install records — verified resolving `bit:bot-dev` in `tools/example` from unpushed bit-pro
  source. Out of scope here, but it is the development path that removes the push-to-test loop.

## Verses

- [x] Verse 1 — Cut a version without typing one: `just release <level>` bumps plugin.json,
  commits, and creates the local tag. Its first real use is this project's baseline — run it
  and `bp version` reports `0.1.0` instead of a git sha. Refuses to go backwards and refuses
  a dirty tree anywhere in the repo.
  Touches: `Justfile`, `bit/.claude-plugin/plugin.json`, `cmd/root.go`.

- [x] Verse 2 — Publish a release deliberately: `just release-push` sends the tag to origin
  as a separate, guarded step, refusing when tracked changes are uncommitted and ignoring
  untracked files. Its first real use publishes `v0.1.0`, the baseline Verse 3 observes
  against.
  Touches: `Justfile`.

- [x] Verse 3 (spike) — Settle what "update" actually means end to end: observe whether an
  already-installed consumer picks up the pushed `0.1.0` on its own, only on command, or not
  at all, and record what a consumer can cheaply read to learn "latest". Answered: manual
  only. Findings are the "Settled by Verse 3" decisions above.
  Touches: the installed plugin cache, `claude plugin list --json`.

- [x] Verse 4 — Read a version that means something: `bp -v` reports the number
  `plugin.json` declares — `0.1.0` — instead of the `git describe` sha it prints today, so what
  the operator reads can be compared against what the release published. The build stops
  deriving a version from git at all.
  Touches: `scripts/install.sh`, `Justfile`, `cmd/root.go`.

- [x] Verse 5 — Know when the plugin behind you is stale: running `bp` in any project prints a
  short notice when the installed bit plugin is behind the latest published, naming the exact
  command to fix it, and stays silent otherwise. The comparison reads local files only; the
  same run also kicks off a detached marketplace refresh it never waits on, so the next run is
  accurate. Never blocks, delays, or fails `bp`.
  Touches: `cmd/root.go`, `claude/`.

- [ ] Verse 6 — Declare the machinery stable: cut and publish `v1.0.0` with the same recipes
  the earlier verses built, so the project's version is one an operator can trust and compare.
  This is the deliberate major described in Decisions, and it is the last thing the track does.
  Touches: `Justfile`, `bit/.claude-plugin/plugin.json`.

## References

- `START-HERE.md` — the 2026-08-26 dispatch design session. Its "Versioning" measurements are
  the evidence behind Verses 1–2. Its open question ("Does Claude auto-detect a version bump
  once we start versioning? Unverified") was settled by Verse 3: it does not.
- `BIT-41.5` body, `## Observed` — the full verbatim record of the Verse 3 spike: every command,
  its output, the timings, and the offline measurements. The source for the decisions above.
- anthropics/claude-code#27257 — the project-scope install bug that actually blocked BIT-39,
  fixed in 2.1.248. Cited so this scope is not credited with fixing it.