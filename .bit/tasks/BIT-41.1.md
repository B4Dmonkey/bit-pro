---
id: BIT-41.1
title: Justfile reads the bit--v tag namespace
status: done
approved: true
phase: 1
phase_label: cut a version
---
## **Verse 1**

Make `bp --version` *capable* of reporting a semver, and clear the `author` warning so the
release recipe can validate strictly later. No version value is written here — the recipe cuts
the baseline (BIT-41.3), never a hand edit.

## Scope
- `Justfile` — `version :=` matches the `bit--v*` tag namespace and strips the `bit--v` prefix,
  so ldflags carries a bare semver rather than a tag name.
- `bit/.claude-plugin/plugin.json` — add `author`.

## Steps
- [ ] `Justfile`: change the `version :=` backtick to
      `git describe --tags --match 'bit--v*' --always --dirty 2>/dev/null | sed 's/^bit--v//' || echo dev`.
      The three shapes it must handle: no tag → `28418fe-dirty` (unchanged, verified today);
      at the tag → `bit--v0.1.0` → `0.1.0`; mid-work → `bit--v0.1.0-3-gabc123-dirty` →
      `0.1.0-3-gabc123-dirty`. `--match` is what keeps a future non-plugin tag out of the
      version string.
- [ ] `plugin.json`: add `"author": { "name": "josiah" }` — the object shape used by every
      versioned plugin in the local cache (`go@go-skills`, `ai@pydantic-skills`), and the name
      the enclosing `marketplace.json` already declares as `owner`.

## Claude verifies
- [ ] `claude plugin validate bit` — the `author` warning is gone and the `version` warning
      remains. The remaining warning is expected: no version exists until BIT-41.3 cuts one.
- [ ] `just install && bp --version` → still `bp version 28418fe…`. No `bit--v*` tag exists yet,
      so the sha fallback is the correct output at this bar.
- [ ] `just test` passes.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`chore(release): match bit--v tags and declare plugin author`