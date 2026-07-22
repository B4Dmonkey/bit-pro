---
id: BIT-7.2
title: init seeds the embedded .claude tree
status: todo
phase: 2
phase_label: seed skills
---
`bit init` reproduces the whole embedded `.claude` tree — `bit-cli.md` plus the four `bit_*/SKILL.md` — into the target repo, establishing the embed→write path. The seeded skills open with "read `.claude/bit-cli.md`", so the contract must land alongside them or the port is broken. A single-file write can't create the nested skill dirs, so `Seed` walks the embedded FS from the start.

**Scope:**
- `assets/assets.go` (new) — `package assets`; `//go:embed skills bit-cli.md`; exported `var FS embed.FS`; `func Seed(destRoot string) error` implemented as `fs.WalkDir` over `FS` (dirs → `MkdirAll`, files → `WriteFile` under `destRoot`, converting the embed's forward-slash paths with `filepath.FromSlash`).
- `assets/bit-cli.md` (new) — moved from `.claude/bit-cli.md`, every `./bin/bit` changed to `bit`. Now the source of truth.
- `assets/skills/bit_{scope,plan,do,check}/SKILL.md` (new) — moved from `.claude/skills/…`, each with `./bin/bit`→`bit`.
- `cmd/init.go` — after `SaveConfig`, call `assets.Seed(claudeDir)`.
- `cmd/root.go` — add `const claudeDir = ".claude"`.
- `.claude/bit-cli.md`, `.claude/skills/bit_*/SKILL.md` — become regenerated output (init writes them back; content switches `./bin/bit`→`bit`).
- `cmd/init_test.go` — new test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestInitCmd_SeedsClaudeTree`
     - **Behavior:** `bit init` reproduces the whole embedded `.claude` tree under the target root.
     - **Setup:** `t.Chdir(t.TempDir())`; `mustRun(t, "init", "--prefix", "BIT")`.
     - **Assertions:** each of `.claude/bit-cli.md` and `.claude/skills/bit_{scope,plan,do,check}/SKILL.md` exists and its bytes equal the embedded copy (`assets.FS` read); and `.claude/CLAUDE.md` does **not** exist (only bit-cli.md + skills are seeded).
     - **Boundary:** flat file (`bit-cli.md`, depth 1) and nested files (`skills/bit_*/SKILL.md`, depth > 1) the single-file write can't create, plus the exclusion case (CLAUDE.md absent).
   - [ ] Confirm fails: `.claude/bit-cli.md` (and the skills) missing — init writes nothing under `.claude/` today.

2. **Implement (GREEN):**
   - [ ] Move `.claude/bit-cli.md` → `assets/bit-cli.md` and the four `.claude/skills/bit_*/SKILL.md` → `assets/skills/bit_*/`; replace `./bin/bit` with `bit` throughout each (including bit-cli.md's title line).
   - [ ] Create `assets/assets.go` with the embed and `Seed` as an `fs.WalkDir` (dirs → `MkdirAll`, files → `WriteFile`), preserving relative structure under `destRoot`.
   - [ ] Call `assets.Seed(claudeDir)` from `init` after config is saved; add the `claudeDir` const in `cmd/root.go`.
   - [ ] Regenerate this repo's `.claude/` (run init here — `go run . init --prefix BIT`).

**Claude verifies:**
- [ ] `just test` green (new test passes; existing init tests unaffected).
- [ ] `just lint`, `just build` green.

**User verifies:**
- [ ] `git status`: `assets/bit-cli.md` + `assets/skills/*` added; `.claude/bit-cli.md` and `.claude/skills/*` diffs are only `./bin/bit`→`bit`, nothing else; working skills still load in Claude Code.
- [ ] Grep confirms no `./bin/bit` remains in seeded bit-cli.md or skills; source of truth is now `assets/`.
- [ ] Note the consequence: bit-pro dev now drives the installed `bit`; run `just install` after CLI changes.

**Commit (user):** `feat(init): seed the embedded .claude tree on init`