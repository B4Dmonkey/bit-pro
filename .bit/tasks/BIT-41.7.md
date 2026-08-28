---
id: BIT-41.7
title: bp -v reports the released version
status: todo
approved: true
phase: 4
phase_label: read a version
---
## **Verse 4**

`bp -v` reports the version `plugin.json` declares instead of the `git describe` sha it prints
today, so the number an operator reads is comparable against a published release. The build stops
consulting git for a version at all.

This bar changes shell only, so it follows the `## Steps` shape of BIT-41.1 and BIT-41.4 rather
than a TDD cycle — there is no Go seam to test-drive. `cmd/root.go` is deliberately untouched:
`var version = "dev"` stays as the no-ldflags fallback and the `bp version <x>` output template is
unchanged, so the `bare version` this verse delivers is the bare *semver* (`0.1.0`, never
`0.1.0-4-gdf72130`), not a bare line.

The manifest is read at build time through ldflags rather than embedded into the binary. Embedding
would couple the `bp` binary to the plugin's manifest file, and the two may diverge later.

## Scope
- `scripts/install.sh` — the `version=` line reads `bit/.claude-plugin/plugin.json` instead of
  running `git describe`.
- `Justfile` — delete the `version :=` line. It is dead: `install.sh` computes its own value and
  nothing in the Justfile references the variable (`grep -n version Justfile` shows the definition
  and two doc comments, no use).

## Steps
- [ ] `scripts/install.sh`: replace
      `version="$(git describe --tags --match 'v*' --always --dirty 2>/dev/null | sed 's/^v//' || echo dev)"`
      with `version="$(jq -r '.version // "dev"' bit/.claude-plugin/plugin.json)"`. `jq` is already
      a hard dependency of `scripts/release.sh`, so this adds nothing to the toolchain. The script
      already `cd`s to the repo root, so the relative path resolves.
- [ ] Leave the `-ldflags` line as it is — the mechanism is unchanged, only its input.
- [ ] `Justfile`: delete the `version :=` line at the top of the file.

Accepted cost, following from the Decision this verse implements: a build cut between releases
reports the same string as the release itself, and `go run .` / `just run` still report `dev`
because they pass no ldflags. Git answers "which commit is this" for anyone standing in the repo.

## Claude verifies
- [ ] `just install && bp --version` prints `bp version 0.1.0` — not the `bp version 7c4023e` it
      prints today.
- [ ] `bp -v` prints the same string. Cobra binds `-v` to `--version` on the root because no other
      root shorthand claims it; confirmed against the installed binary.
- [ ] `grep -c 'git describe' Justfile scripts/install.sh` reports 0 for both files.
- [ ] `just test` and `just lint` pass.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(release): bp reports the version plugin.json declares`