---
id: BIT-41
title: Versioning for bp and the bit plugin
status: doing
---
## Why

Nothing on this machine can answer "am I running current?" — not the operator, and not the
tool. `bp version` prints a git sha (`28418fe`), the installed plugin lists as
`4ebbe7cd5eff`, and neither string can be compared to anything. That is not cosmetic: the
dispatch defect chain in START-HERE.md ran for months partly because a *broken* plugin
install was indistinguishable from a *stale* one, and there was no version to check.

The exposure is real rather than theoretical. `bp` is built from local source while the
plugin installs from GitHub, so the binary can sit ahead of the pushed plugin and the
installed plugin cache can lag behind origin — they skew in both directions, silently, and
they must agree because the skills the plugin ships are the only thing that drives the CLI.

## Summary

Adopt SemVer across `bp`, the plugin, the `.bit/` format and the db schema as one lockstep
number, starting at 0.1.0, bumping minors as the track develops, and arriving at 1.0.0 when
it completes. Cutting a release stops being a hand-typed `git tag` and becomes
`just` recipes that take a bump *level*, never a version string. Then close the loop at the
consumer end: running `bp` tells the operator when a newer version is available.

## Visual aid

```
just release <level>
      |
      +-- bumps   bit/.claude-plugin/plugin.json    <- source of truth
      |             (no version yet -> establishes 0.1.0)
      +-- commits "release: v0.1.0"
      +-- claude plugin tag ------------------> git tag bit--v0.1.0  (local)
                                                       |
just release-push -------------------------------------+--> origin
                                                       |
                          git describe --match 'bit--v*'
                                       |
                            ldflags -> cmd.version
                                       |
                                 bp version -> 0.1.0
```

## Risks & unknowns

- **Unknown:** Does a pushed, tagged version actually reach a machine that already has the
  plugin installed — and if a consumer has to be told, where does it read "latest" from?
  Measured today: no `claude plugin` subcommand reports that an update is *available*.
  `list` shows what is installed, `update` just performs one and says "restart required to
  apply". So the notice in Verse 4 has no built-in source to ask.
  **Resolve by:** Verse 3 (spike). Observe an already-installed consumer against the
  `bit--v0.1.0` that Verses 1–2 cut and pushed — it currently reports the git sha
  `4ebbe7cd5eff`, so it is already behind. Answer is **yes/automatic** if `claude plugin list
  --json` starts reporting `"version": "0.1.0"` without an explicit command; **yes/manual** if
  it only changes after `claude plugin update bit@bit-pro`; **no** if it stays pinned at the
  git sha either way.
  Separately record which of (`git ls-remote --tags`, the marketplace entry, the plugin
  cache, `plugin list --json`) can be read cheaply enough to serve as "latest".
  **Downstream:** Verse 4 depends entirely — automatic updates make the notice nearly
  pointless, manual updates make it obligatory and it must name the command to run. Verses
  1 and 2 do not depend on the answer.
  **Two transitions, both inside the spike:** a consumer on the git sha moving to `0.1.0` is
  not necessarily the same event as one on `0.1.0` moving to `0.2.0` — the first could be
  detected as mere difference, the second requires comparing semver. The spike observes both
  by cutting `0.2.0` after the first observation, so the answer covers the case that will
  actually recur.
  **Artifact:** kept — `0.1.0` and `0.2.0` are this project's real releases, not throwaways.
  **De-risk before planning?** Yes — but it cannot run first. Producing a pushed, tagged
  version is exactly what Verses 1 and 2 build, so the spike sits third by necessity rather
  than by preference.

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
- **The recipe cuts its own baseline; no version is ever hand-typed.** There is no
  pre-existing `bit--v*` tag, and none gets created by hand. Verse 1's first real use of
  `just release` is what establishes 0.1.0 — with no prior version in plugin.json the
  invocation produces 0.1.0 regardless of the level passed, since there is nothing to bump.
  The monotonic guard passes trivially in that one case (no tag to be greater than) and
  binds on every release after it.
- **Minors while the track develops; 1.0.0 when it completes.** This track releases itself
  repeatedly rather than once, which is the point — the recipes get exercised several times
  before anything depends on them. Two of those releases are load-bearing and named here:
  `0.1.0`, cut by Verse 1 and published by Verse 2, and `0.2.0`, cut by Verse 3 so the spike
  can observe a version-to-version update. Further minors are fine as the work reaches
  publishable points; nothing downstream depends on how many.
- **1.0.0 is a one-time major, cut when this track completes.** It is the single deliberate
  override of the trigger guideline — a completed track would otherwise be a minor, and no
  migration is involved. The reason is that 1.0.0 is a claim about the versioning machinery
  itself: until this track lands, no version on this project can be compared to anything, so
  a stable number would be asserting something untrue. It happens exactly once; every release
  after it follows the guideline again, so BIT-39 lands as 1.1.0.
- **`plugin.json` is the source of truth; the git tag is the derived marker.** It is what a
  consumer sees when installed from an untagged commit, and `claude plugin tag` already
  reads it to build the tag name. `.claude-plugin/marketplace.json` carries no version
  field, so plugin.json is the only file a release writes.
- **The recipe takes a level, never a version string.** `major` / `minor` / `patch` only, so
  `0.2.0 -> 0.1.0` is unrepresentable rather than merely discouraged. Backed by a monotonic
  guard: the computed next version must be strictly greater than the highest existing
  `bit--v*` tag.
- **Never pass `-f` to `claude plugin tag`.** Its dirty-tree and tag-exists checks are the
  protection; `-f` is the only thing that removes them.
- **Creating a tag and pushing it are separate recipes.** A local tag is trivially
  deletable; a pushed one is not. The irreversible step gets its own deliberate invocation.
- **Dirty means tracked-and-uncommitted; untracked files are ignored.** Measured: this is
  already `claude plugin tag`'s exact behaviour — an untracked file passes, a modified
  tracked file is refused with "Uncommitted changes affecting this release". **But its check
  is scoped to the plugin directory only** — a dirty `cmd/root.go` sails straight through.
  Under lockstep that is a hole, so the recipes add a repo-wide guard with the same
  tracked/untracked semantics.
- **Add an `author` field to plugin.json.** Measured: with `version` present, the lone
  remaining warning is `author`, and `validate --strict` exits 1 on it. The marketplace
  already declares owner `josiah`, so this costs nothing and lets the recipe validate
  strictly.
- **The version notice must never block, delay, or fail `bp`.** Whatever source it reads,
  an offline machine, a slow network, or a missing remote produces silence — not an error,
  not a hang. `claude plugin list --json` was measured at ~0.3s, which is the upper bound of
  what is acceptable even on a session-shaped entry point.
- **Skew comparison uses the base version, not the describe string.** A mid-work build
  reports `0.1.0-3-gabc123-dirty` while plugin.json reads `0.1.0`; those agree.

## Verses

- [x] Verse 1 — Cut a version without typing one: `just release <level>` bumps plugin.json,
  commits, and creates the local tag. Its first real use is this project's baseline — run it
  and `bp version` reports `0.1.0` instead of a git sha. Refuses to go backwards and refuses
  a dirty tree anywhere in the repo.
  Touches: `Justfile`, `bit/.claude-plugin/plugin.json`, `cmd/root.go`.

- [ ] Verse 2 — Publish a release deliberately: `just release-push` sends the tag to origin
  as a separate, guarded step, refusing when tracked changes are uncommitted and ignoring
  untracked files. Its first real use publishes `bit--v0.1.0`, the baseline Verse 3 observes
  against.
  Touches: `Justfile`.

- [ ] Verse 3 (spike) — Settle what "update" actually means end to end: observe whether an
  already-installed consumer picks up the pushed `0.1.0` on its own, only on command, or not
  at all, then cut and push `0.2.0` and observe the same for a version-to-version bump — and
  record what a consumer can cheaply read to learn "latest".
  Touches: `.claude-plugin/marketplace.json`, the installed plugin cache, `claude plugin
  list --json`.

- [ ] Verse 4 — Know you are behind without going looking: running bare `bp` prints a notice
  when a newer version is available, shaped by whatever Verse 3 found, and stays silent
  otherwise.
  Touches: `cmd/root.go`.

## References

- `START-HERE.md` — the 2026-08-26 dispatch design session. Its "Versioning" measurements
  are the evidence behind Verses 1–2, and its open question ("Does Claude auto-detect a
  version bump once we start versioning? Unverified") is the same unknown Verse 3 settles.