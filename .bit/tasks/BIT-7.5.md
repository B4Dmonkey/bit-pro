---
id: BIT-7.5
title: Retire just build now that init installs bit on PATH
status: done
phase: 2
phase_label: seed skills
---
Retires the now-vestigial `just build` recipe and its `bin/bit` artifact. Once 7.2 migrates `bit-cli.md` and the four skills to call `bit` on PATH, nothing in the repo drives `./bin/bit` anymore, so the recipe that built it is dead weight — dropping it here (not earlier) keeps every prior commit self-consistent. No Go test: this is build-tooling cleanup, not application logic, so the check is that the suite/build still pass and no `just build` / `./bin/bit` references survive in the code and docs (a unit test would exercise `just`, not our code — YAGNI).

**Scope:**
- `Justfile` — remove the `build:` recipe; keep `run`, `test`, `lint`, and the `install:` recipe from 7.1.
- `bin/` — gitignored, so the stale `bin/bit` is a local-only artifact; delete it by hand (`rm -rf bin/`). No tracked change.
- `.claude/settings.local.json` — gitignored (local, not committed): add `Bash(bit task *)` and drop the `./bin/bit` entries so the migrated workflow drives `bit` on PATH without a permission prompt. Local tidy, not part of the commit.

**Implement (cleanup — no TDD):**
- [ ] Delete the `build:` recipe from the `Justfile`.
- [ ] `rm -rf bin/` locally.
- [ ] Add `Bash(bit task *)` to `.claude/settings.local.json` (local convenience).

**Claude verifies:**
- [ ] `just install` still builds `bit` into the Go bin dir and `bit --help` runs.
- [ ] `just test`, `just lint` green.
- [ ] `git grep -n "just build" -- ':!.bit'` and `git grep -n "\./bin/bit" -- ':!.bit'` return nothing — confirms `bit-cli.md` and the skills are fully migrated and no live doc still points at the retired path. The `:!.bit` exclusion is required: `.bit/tasks/*.md` are historical records of completed work whose bodies legitimately quote the old strings, and they are never hand-edited.

**User verifies:**
- [ ] From a clean checkout: `just install`, then a `bit_*` workflow step drives `bit` on PATH with no `bin/bit` present. Comfortable that `just build` is gone for good (quick iteration is `just run`; PATH is `just install`).

**Commit (user):** `chore(build): retire just build now that init installs bit on PATH`