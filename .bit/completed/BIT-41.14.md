---
id: BIT-41.14
title: Every run kicks a detached marketplace refresh
status: done
approved: true
phase: 5
phase_label: stale notice
---
## **Verse 5**

Closes the loop. The marketplace clone is the free offline source of "latest", but it only moves
when `claude plugin marketplace update` runs — left alone it matches the install forever and the
notice never fires. So `bp` fires the refresh itself on every run: detached, output discarded,
never waited on. A refresh benefits the *next* run, never this one, which is what keeps the hot
path free.

No timestamp, no once-a-day guard, no stamp file. Two runs refreshing the same clone at once,
an offline machine, a missing `claude` binary — every one of them is silent by design.

This bar completes the verse, so it carries the whole-slice check.

## Scope
- `claude/plugin.go` — `RefreshMarketplace()` plus an internal `start(name string, args ...string)`
  seam the test can drive with a harmless command.
- `cmd/root.go` — a `refreshMarketplace` package var defaulting to `claude.RefreshMarketplace`;
  `execute` fires it for every command `suppressed` does not exempt.
- `cmd/cmd_test.go` — `runSplit` stubs `refreshMarketplace` so the package's tests never shell out.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStart_DoesNotWaitForTheChild`
     - **Behavior:** the refresh returns as soon as the child is launched, rather than waiting for
       it to finish — the property the whole never-block decision rests on.
     - **Setup:** call the internal seam with `sleep`, `3`; record elapsed time with `time.Now()`.
     - **Assertions:** returns a nil error in under 1 second.
     - **Boundary:** the child's 3s runtime against a 1s bound — a `Wait()` implementation takes 3s
       and fails; anything that only launches takes milliseconds.
   - [ ] `TestStart_MissingBinaryIsSilent`
     - **Behavior:** a machine with no `claude` on PATH gets silence, not an error.
     - **Setup:** call the seam with a command name that does not exist.
     - **Assertions:** returns nil; does not panic.
     - **Boundary:** the launch failing at its earliest point — the lower bound of the child
       lifecycle.
   - [ ] `TestExecute_FiresTheRefresh`
     - **Behavior:** an ordinary command kicks the refresh exactly once.
     - **Setup:** `refreshMarketplace` stubbed to a counter; `pluginState` left at "nothing known";
       run `task list` through `runSplit`.
     - **Assertions:** the counter is 1; stderr is still empty.
     - **Boundary:** one invocation per run — not zero (the refresh is unconditional for
       non-exempt commands) and not more than one.
   - [ ] Confirm fails: `undefined: claude.RefreshMarketplace`; then, before `execute` calls it, the
         counter is 0.

2. **Implement (GREEN):**
   - [ ] `start(name string, args ...string) error`: `exec.Command`, leave `Stdout`/`Stderr` nil so
         the child's output is discarded, `cmd.Start()`, and **no** `Wait` — the child is reparented
         when `bp` exits. Swallow the start error.
   - [ ] `RefreshMarketplace()` calls
         `start("claude", "plugin", "marketplace", "update", "bit-pro")`, reusing the marketplace
         name const.
   - [ ] `cmd/root.go`: `var refreshMarketplace = claude.RefreshMarketplace`; `execute` calls it
         after resolving the command and before writing the notice, on the same non-suppressed path
         — so `tui` and `serve mcp` are wholly untouched by this verse, as the scope requires.
   - [ ] `runSplit` sets `refreshMarketplace` to a no-op unless the test replaced it.

## Claude verifies
- [ ] `just test` passes.
- [ ] `just lint` passes.
- [ ] `just install`, then `bp task list` in this repo: silent (installed and latest are both
      `0.1.0`), and it returns as fast as it does today — the refresh is not being waited on.

## User verifies
- [ ] Whole slice — make the machine look stale, without publishing anything:
      1. `sed -i '' 's/"version": "0.1.0"/"version": "9.9.9"/' ~/.claude/plugins/marketplaces/bit-pro/bit/.claude-plugin/plugin.json`
      2. `bp task list` — the list prints as normal, and stderr carries exactly
         `bp: bit plugin 0.1.0 → 9.9.9 available — run: claude plugin update bit@bit-pro --scope project`
      3. `bp task list` again — silent this time, because run 2's *predecessor* fired the detached
         refresh, which pulled the clone back to what origin says. That is the design in one
         observation: the refresh helps the next run, never its own.
      4. `bp tui` — no stray notice line anywhere in the frame; quit with `q`.
      5. If step 3 did not restore it (offline, say), reset the clone by hand:
         `git -C ~/.claude/plugins/marketplaces/bit-pro checkout -- bit/.claude-plugin/plugin.json`
      6. `bp task list | cat` — stdout is only the list, so a shell capture is still clean.

## Commit (user)
`feat(version): refresh the marketplace clone in the background on every run`