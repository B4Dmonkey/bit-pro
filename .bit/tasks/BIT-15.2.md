---
id: BIT-15.2
title: Asset files call bp
status: todo
phase: 2
phase_label: Skill assets call bp
---
## **Verse 2**

Replace every CLI invocation of `bit` with `bp` across the five embedded asset files — these are the source of truth that `bp init` seeds into `.claude/`. No unit test is possible for embedded text content; the verification is a grep after editing and a re-seed to confirm the live files update.

## Scope
- `assets/bit-cli.md` — replace `bit task`, `bit init`, `bit tui`, and prose `` `bit` `` (as CLI name) with `bp`
- `assets/skills/bit_scope/SKILL.md` — same replacement
- `assets/skills/bit_plan/SKILL.md` — same replacement
- `assets/skills/bit_do/SKILL.md` — same replacement
- `assets/skills/bit_check/SKILL.md` — same replacement

Keep unchanged: `.bit/` (directory name), `BIT-` (task ID prefix), `bit_plan`/`bit_scope`/`bit_do`/`bit_check` (skill names), `bit-cli.md` (filename), `bit-pro` (repo/module name).

## Implementation

- [ ] In each file, replace all of: `bit task`, `bit init`, `bit tui`, and `` `bit` `` when used as a prose reference to the binary — with `bp task`, `bp init`, `bp tui`, `` `bp` `` respectively.
  The distinction: "run `bit task create`" → "run `bp task create`" (rename). "`.bit/` directory" and "BIT-7" stay unchanged (project data, not binary name).
- [ ] After editing all five files: `just install` (rebuilds `bp` with the updated embedded assets)
- [ ] `bp init` — re-seeds `.claude/` with the new content

## Claude verifies
- [ ] `grep -rn 'bit task\|bit init\|bit tui' assets/` returns no results
- [ ] `just install` exits 0
- [ ] `bp init` exits 0

## User verifies
- [ ] `grep -rn 'bit task\|bit init\|bit tui' .claude/` returns no results — the live skill files reflect the rename
- [ ] `bp tui` — TUI opens (smoke-tests the whole chain: build + init + command)

## Commit (user)
`feat(assets): update embedded skill files to call bp instead of bit`