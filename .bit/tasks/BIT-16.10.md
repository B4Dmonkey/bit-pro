---
id: BIT-16.10
title: Init can no longer ship a stale skill copy
status: todo
phase: 4
phase_label: One source of truth
---
## **Verse 4**

A test asserting that init leaves no `.claude/skills/` cannot pass while `assets.Seed` still runs,
and the only way to green it is to delete the seeding path outright. That is what makes this a
red-green cycle rather than tidying: the deletion is forced by an invariant, and the test stays
behind as the guard that init can never grow a second copy of a `SKILL.md` again.

## Scope
- `cmd/init_test.go` — add `TestInitCmd_WritesNoSkills`; delete `TestInitCmd_SeedsClaudeTree` and
  `TestInitCmd_ReseedRefreshes`, and the `assets` import they were the only users of.
- `cmd/init.go` — drop the `assets.Seed(claudeDir)` call and the import.
- `assets/assets.go` — delete `Seed`; narrow the directive to `//go:embed bit-cli.md`.
- `assets/skills/` — deleted.
- `.claude/skills/bit_scope/`, `bit_plan/`, `bit_do/`, `bit_check/` — deleted from this repo. The
  plugin is the source now, and leaving these means `/bit_scope` and `/bit:scope` both answer with
  possibly different text, which is the exact ambiguity this verse exists to remove.
- `.claude/bit-cli.md` — deleted; `bp instructions` prints it and nothing reads the file any more.

`claudeDir` in `cmd/root.go` stays — Verse 3 still uses it to locate `settings.json`.

The contract doc stays embedded. Only `skills` leaves the embed directive; `bit-cli.md` is what
`bp instructions` reads, and the existing `TestInstructionsCmd_PrintsContract` is the guard that
narrowing the directive did not take it with it.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestInitCmd_WritesNoSkills`
     - **Behavior:** init no longer writes skill content anywhere, so a repo cannot end up with an
       older `SKILL.md` shadowing the plugin's — and its remaining `.claude/` job, the wiring, is
       untouched by the removal.
     - **Setup:** `t.Chdir(t.TempDir())`; `mustRun(t, "init", "--prefix", "BIT")`.
     - **Assertions:** `os.Stat(".claude/skills")` gives `fs.ErrNotExist`; `os.Stat(".claude/bit-cli.md")`
       gives `fs.ErrNotExist`; `.claude/settings.json` still exists.
     - **Boundary:** seeded-file count == 0, the lower bound — measured against the five files
       `TestInitCmd_SeedsClaudeTree` asserted right up until this bar deletes it.
   - [ ] Confirm fails: `.claude/skills` exists, because `assets.Seed` still runs. If it fails
     because `settings.json` is missing instead, Verse 3 regressed and that is a different problem.

2. **Implement (GREEN):**
   - [ ] Remove the `Seed` call from `init.go`, delete `Seed` from `assets/assets.go`, narrow the
     embed directive, delete `assets/skills/`, and delete the two obsolete tests.
   - [ ] Delete this repo's `.claude/skills/bit_*` and `.claude/bit-cli.md`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install` — the build has to survive the embed change; a directive naming a deleted
      directory is a compile error, which is the good kind of failure here
- [ ] `just run instructions | head -3` still prints the contract
- [ ] `grep -rn "assets\." --include='*.go' .` shows only the `bit-cli.md` read in
      `cmd/instructions.go` and its test
- [ ] `find . -name SKILL.md -not -path './bit/*'` prints nothing — one home for every skill

## User verifies
- [ ] Restart Claude Code in this repo with no `--plugin-dir`. `/bit_scope` and friends are gone;
      `/bit:scope`, `/bit:plan`, `/bit:do`, `/bit:check` are offered and are being served from the
      installed plugin rather than the working tree.
- [ ] Whole slice: run `/bit:scope` on something small. It drives `bp` correctly having learned the
      contract from `bp instructions`, with no `.claude/bit-cli.md` on disk to fall back on.
- [ ] Worth knowing before you start: this deletes the local skills the pipeline currently runs
      from, so run it when you are ready to switch over rather than mid-task.

## Commit (user)
`refactor(assets): drop the embedded skill tree and init's seeding`