---
id: BIT-48.1
title: EnsureGlobal runs the four user-scope claude commands in order and stops at the first failure
status: done
approved: true
phase: 1
phase_label: bp add sets bit up for the whole machine
---
## **Verse 1**

The four wiring commands live in one place in `claude/`, so `bp add`, BIT-49's `bp migrate`, the failure message and the README all quote the same lines. This bar is additive: nothing calls it yet. A test against a home with no `.claude.json` forces the command list and the run order. A failing step forces the stop and the step-numbered error.

## Scope
- new `claude/global.go`:
  - `func GlobalWiring() [][]string`: returns a fresh slice of four full argvs, each starting with `claude`:
    1. `claude plugin marketplace add B4Dmonkey/bit-pro`
    2. `claude plugin marketplace update bit-pro`
    3. `claude plugin install bit@bit-pro --scope user`
    4. `claude mcp add -s user bit -- bp serve mcp`
    Use `pluginKey` and `marketplaceName` (still in `claude/settings.go` until BIT-48.5 moves them). `bp` stays by name, never `os.Executable()`.
  - `func EnsureGlobal(ctx context.Context, run Runner, home string) error`: runs each argv in order via `run(ctx, argv[0], argv[1:]...)`. On the first error it stops and returns `fmt.Errorf("claude wiring step %d of %d: %w", i+1, len(steps), err)`. `home` is unused in this bar; the next bar uses it.
- new `claude/global_test.go`: one top-level `TestEnsureGlobal`, which BIT-48.2 extends. Reuse `recorder`/`newRecorder` from `claude/sync_test.go` (same package; this bar doesn't edit `sync_test.go`). Pin the four argvs as literals in the test, not via `GlobalWiring()`, so the test catches a changed command. Where a token repeats, use the test consts `claude/sync_test.go` already declares in this package (`claudeBin`, `pluginSubCmd`, `bitProPlugin`, `updateSubCmd`, `scopeFlag`, `mcpSubCmd`, `bitServer`) rather than new literals; BIT-48.5 moves the ones used here into `global_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestEnsureGlobal/runs the four commands in order` (a `t.Run` subtest)
     - **Behavior:** on a machine with no bit wiring, one call sets up the marketplace, refreshes it, installs the plugin at user scope and registers the MCP server at user scope.
     - **Setup:** `home := t.TempDir()` (no `.claude.json`); `rec := newRecorder(nil)`; `EnsureGlobal(t.Context(), rec.Run, home)`.
     - **Assertions:** nil error; `rec.calls` equals the four literal argvs above, in order; `GlobalWiring()` equals the same four.
     - **Boundary:** every step succeeds — the full four-call path.
   - [ ] Confirm fails: `EnsureGlobal` / `GlobalWiring` undefined.

2. **Implement (GREEN):**
   - [ ] `GlobalWiring` and the loop in `EnsureGlobal`.

3. **More tests (RED → GREEN):**
   - [ ] `TestEnsureGlobal/stops at the first failing step` (a `t.Run` subtest; its recorder fails a step and its assertions differ from the first case)
     - **Behavior:** a failed step is reported by number, and later steps don't run against a half-set-up machine.
     - **Setup:** `boom := errors.New("plugin bit not found")`; `newRecorder(map[int]error{2: boom})` (the install, index 2).
     - **Assertions:** `errors.Is(err, boom)`; `err.Error()` contains `step 3 of 4`; `len(rec.calls) == 3`.
     - **Boundary:** failure mid-sequence (step 3 of 4) — neither the first nor the last.

## Claude verifies
- [ ] `just lint` and `just test` pass (if `goconst` flags the repeated `claude`/`plugin` literals in `global.go`, hoist them to unexported consts there, with names that don't clash with the `sync_test.go` consts listed above, since both files are in package `claude`)

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): claude.EnsureGlobal runs the user-scope wiring commands`