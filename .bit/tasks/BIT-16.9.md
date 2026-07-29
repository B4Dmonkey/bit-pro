---
id: BIT-16.9
title: A fresh clone installs rather than fails
status: doing
phase: 3
phase_label: Init keeps it current
---
## **Verse 3**

A repo that has never had the plugin contradicts the previous bar: there is nothing for `plugin
update` to update, so the single-verb happy path cannot satisfy a fresh clone. That forces the
install fallback — and with it, the verse's actual promise that one command works whether or not
the repo already has the plugin.

**Why update first, then install** — this is the strategy that needs no facts nobody has. Trying
`update` first is safe under either possible behaviour of `install` when the plugin is already
present: if `install` refuses, we never call it in that case; if `install` silently no-ops, we have
already updated. Install-first is only safe if you know which of those it does, and if it no-ops
the repo would never come current, which is the entire point of the verse.

## Scope
- `claude/sync.go`
- `claude/sync_test.go`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestSyncPlugin_FallsBackToInstall`
     - **Behavior:** a repo that has never installed the plugin still ends up with it, so the caller
       never has to know which case they are in — the scope's "having one command is the point".
     - **Setup:** a recording `Runner` that returns an error for the `plugin update` call and nil
       for everything else.
     - **Assertions:** `rec.calls` equals exactly, in order, `marketplace update bit-pro`,
       `plugin update bit@bit-pro --scope project`, `plugin install bit@bit-pro --scope project`;
       `SyncPlugin` returns nil, because the fallback succeeding means the repo is current and a
       failed probe is not an error the user should see.
     - **Boundary:** failure count == 1, at the update step — the first case above the previous
       bar's zero-failure path.
   - [ ] Confirm fails: `SyncPlugin` returns the update error and `install` is never called.

2. **Implement (GREEN):**
   - [ ] On a `plugin update` failure, run `plugin install bit@bit-pro --scope project` and return
     that call's result.

3. **More tests (RED → GREEN):**
   - [ ] `TestSyncPlugin_ReportsWhenInstallAlsoFails`
     - **Behavior:** when the plugin genuinely cannot be fetched, init says so — it does not leave a
       settings file naming a plugin nobody has. This is the scope's "a written settings file is
       never to be treated as a working install," which is only true if this path reports.
     - **Setup:** a `Runner` that fails both `plugin update` and `plugin install`, with a
       distinguishable message on the install.
     - **Assertions:** `SyncPlugin` returns non-nil and the message carries the install failure, not
       the swallowed update one — the update failing is expected in a fresh clone and would be a
       misleading thing to report.
     - **Boundary:** failure count == 2 — the upper bound, where no fallback is left.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice: clone this repo into a throwaway directory — `git worktree add` is enough — and
      run `bp init` there with any prefix. Restart Claude Code in that directory with **no**
      `--plugin-dir`, and `/bit:scope` is offered. One command in a fresh repo ended with the
      current skills available, which is the whole verse.
- [ ] The `--plugin-dir` omission is the point, not a detail: a session started with it takes
      precedence over the installed plugin, so it can never be evidence that the install worked.
- [ ] Cleanup, if you care: that throwaway path now has a project-scoped install recorded against
      it. `claude plugin uninstall bit@bit-pro --scope project` from inside it, or leave the stale
      entry — it is inert.

## Commit (user)
`feat(claude): fall back to install when the plugin is absent`