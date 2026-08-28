---
id: BIT-41.13
title: Real state replaces the hardcoded reader
status: done
approved: true
phase: 5
phase_label: stale notice
---
## **Verse 5**

Wire the two readers into the notice, so the feature stops being inert: `pluginState` swaps its
"reports nothing" default for one that resolves the home directory and this project's root and
calls `InstalledVersion` and `LatestVersion`.

"This project" needs defining, because `bp` can run from inside a `.claude/worktrees/<name>`
checkout while the install record names the outer repository — `bitdir` already performs exactly
that cut for `.bit/`, so the project root follows the same rule rather than a second one.

## Scope
- `bitdir/bitdir.go` — `Root() string`: the outer checkout when the working directory is inside
  `.claude/worktrees/`, otherwise the working directory. Reuses the existing `worktreeCut`.
- `claude/plugin.go` — `PluginState(home, projectRoot string) (installed, latest string, ok bool)`,
  reporting `ok` only when both reads succeed.
- `cmd/root.go` — `pluginState` defaults to a closure that resolves `os.UserHomeDir()` and
  `bitdir.Root()` and calls `claude.PluginState`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestRoot` (in `bitdir`, table-driven)
     - **Behavior:** the project root is the outer checkout when the working directory sits inside a
       Claude worktree, and the working directory itself otherwise — the same cut `Current()`
       already applies to `.bit/`.
     - **Setup:** `t.Chdir` into each of: a plain temp dir; `<root>/.claude/worktrees/hazy`; and
       `<root>/.claude/worktrees/outer/.claude/worktrees/inner`.
     - **Assertions:** the plain dir returns itself; both worktree cases return `<root>`.
     - **Boundary:** worktree nesting depth at 0, 1 and 2 — the same three cases
       `TestBitDir_*` already covers for `Current()`, so the two functions cannot drift.
   - [ ] `TestPluginState_ReportsThisProject`
     - **Behavior:** with both files present, the state reader returns the installed version
       recorded for this project and the latest the clone declares.
     - **Setup:** a temp home holding both fixtures from BIT-41.11 and BIT-41.12, with the install
       record's `projectPath` set to a temp project dir; installed `0.1.0`, latest `0.2.0`.
     - **Assertions:** `("0.1.0", "0.2.0", true)`.
     - **Boundary:** both reads succeeding — the only state that yields `ok`.
   - [ ] `TestPluginState_SilentWhenEitherReadFails` (table-driven)
     - **Behavior:** a half-known state is not a state — if either read fails, nothing is reported.
     - **Setup:** two temp homes: one with the install record but no marketplace clone, one with
       the clone but no install record.
     - **Assertions:** `ok` is false in both rows.
     - **Boundary:** each of the two reads at its failing edge, independently.
   - [ ] Confirm fails: `undefined: bitdir.Root` / `undefined: claude.PluginState`.

2. **Implement (GREEN):**
   - [ ] `bitdir.Root()`: `os.Getwd()`, then `worktreeCut` — returning the cut path with the
         trailing `.bit` element removed, so the two functions share one rule. Fall back to the
         working directory when the cut does not apply, and to `"."` when `Getwd` fails.
   - [ ] `claude.PluginState(home, projectRoot)`: call both readers; return
         `("", "", false)` unless both report true.
   - [ ] `cmd/root.go`: `pluginState` defaults to a closure resolving `os.UserHomeDir()` — silent on
         error — and `bitdir.Root()`, then calling `claude.PluginState`. The tests from BIT-41.8 and
         BIT-41.10 keep stubbing the var, so they never touch the real home.

## Claude verifies
- [ ] `just test` passes, including every pre-existing `cmd` test — none of them route through
      `execute`, so the new default cannot reach them.
- [ ] `just lint` passes.
- [ ] `just install`, then in this repo `bp task list` prints no notice: installed and latest are
      both `0.1.0`, and equal means silent.

## User verifies
- [ ] none — deterministic. The stale case cannot be observed until something newer is published;
      that is the whole-slice check on BIT-41.14.

## Commit (user)
`feat(version): compare the installed plugin against the marketplace clone`