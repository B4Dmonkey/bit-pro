---
id: BIT-16.8
title: Init brings the plugin current
status: done
phase: 3
phase_label: Init keeps it current
---
## **Verse 3**

The wiring only *registers* the marketplace and *enables* the plugin — neither key fetches a file,
so after the last two bars a repo has settings naming a plugin it does not have. This is where init
actually calls Claude Code, and it is the first `os/exec` anywhere in this repo, so building the
seam is part of the step rather than something to retrofit.

## Scope
- `claude/sync.go` — new: `type Runner func(ctx context.Context, name string, args ...string) error`,
  a real `ExecRunner`, and `SyncPlugin(ctx context.Context, run Runner) error`.
- `claude/sync_test.go` — new: a recording fake `Runner`.
- `cmd/root.go` — `NewRootCmd()` keeps its signature and delegates to an unexported
  `newRootCmd(run claude.Runner)` passing `claude.ExecRunner`; `newInitCmd` takes the runner.
- `cmd/init.go` — call `claude.SyncPlugin(cmd.Context(), run)` after the wiring.
- `cmd/cmd_test.go` — `runWithStdin` builds the tree through `newRootCmd` with a fake runner.

That last file is load-bearing, not cleanup. Every `cmd` test calls `init` through `initProject`,
so without the fake the whole suite would shell out to the real `claude plugin install` and mutate
the developer's own machine on `just test`. A func type is the entire seam — resist promoting it to
an interface until a second implementation actually exists.

**Trap:** `TestInitCmd_PromptShowsExistingPrefix` asserts init's output contains no `(` in the
fresh-project case, because that is how it distinguishes the two prompt forms. Any progress line
this bar adds must avoid parentheses or that test fails for a reason that has nothing to do with
this work.

## References
- `https://code.claude.com/docs/en/plugin-marketplaces` — `--scope` semantics and how a versionless
  plugin resolves to a commit SHA, which is what makes every push an available update.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestSyncPlugin_RefreshesThenUpdates`
     - **Behavior:** the steady-state path is refresh-then-update at project scope — pull the
       catalog clone, then move the install to the newest commit. This is the exact sequence the
       Verse 1 spike watched work end to end, so init reproduces it rather than inventing its own.
     - **Setup:** a recording `Runner` returning nil for every call; `SyncPlugin(t.Context(), rec.Run)`.
     - **Assertions:** `rec.calls` equals exactly, in order,
       `["claude", "plugin", "marketplace", "update", "bit-pro"]` then
       `["claude", "plugin", "update", "bit@bit-pro", "--scope", "project"]`; `SyncPlugin` returns
       nil. `--scope project` is not optional — both commands default to user scope and would
       otherwise write into `~/.claude/settings.json` instead of the repo's.
     - **Boundary:** failure count == 0 — the all-succeed lower bound. The next bar takes it up.
   - [ ] Confirm fails: `undefined: claude.SyncPlugin`.

2. **Implement (GREEN):**
   - [ ] `SyncPlugin` runs the two commands in order, returning early and wrapping any failure with
     which command failed.
   - [ ] `ExecRunner` uses `exec.CommandContext` and `CombinedOutput`, wrapping a failure with the
     captured output so `claude`'s own complaint reaches the user instead of a bare exit status.
   - [ ] `newRootCmd(run)` / `newInitCmd(run)` threading, and `runWithStdin` switched over to a
     nil-returning fake.
   - [ ] One status line to `cmd.OutOrStdout()` before the sync — the commands can take seconds and
     a silent stall reads as a hang. No parentheses, per the trap above.

3. **More tests (RED → GREEN):**
   - [ ] `TestInitCmd_SyncsThePlugin`
     - **Behavior:** bringing the plugin current is part of `bp init` — the single-entry-point
       decision, asserted at the command surface.
     - **Setup:** fresh temp dir; root built with a recording runner; `init --prefix BIT`.
     - **Assertions:** the same two calls, in the same order.
     - **Boundary:** init invocation count == 1 — one command performs the whole path, wiring
       included.
   - [ ] `TestSyncPlugin_StopsWhenTheCatalogRefreshFails`
     - **Behavior:** if the catalog cannot be refreshed there is nothing current to move to, so init
       reports instead of quietly installing whatever stale version is already cached.
     - **Setup:** a `Runner` that fails the `marketplace update` call.
     - **Assertions:** the error is non-nil; `len(rec.calls) == 1` — nothing after it ran.
     - **Boundary:** failure at position 0 — the earliest possible step, which proves the sequence
       short-circuits rather than pressing on.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `grep -rn "os/exec" --include='*.go' .` shows one use, in `claude/sync.go` — nothing leaked
      into `cmd/`, and no test grew a real subprocess

## User verifies
- [ ] `just install`, then `bp init` in this repo, accepting the existing `BIT` prefix. It prints
      its status line, exits 0, and `git diff .claude/settings.json` shows the marketplace entry
      added plus the one-time key reordering — nothing removed.
- [ ] **Do the passive auto-update check before this one.** `bp init` runs `claude plugin
      marketplace update bit-pro`, which refreshes the catalog clone and destroys the standing
      observation of whether a push propagates on its own. Check
      `git -C ~/.claude/plugins/marketplaces/bit-pro log --oneline -1` first if that answer is
      still wanted.

## Commit (user)
`feat(init): bring the bit plugin current on every init`