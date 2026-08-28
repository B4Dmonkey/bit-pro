---
id: BIT-41.8
title: A behind plugin prints the notice on stderr
status: todo
approved: true
phase: 5
phase_label: stale notice
---
## **Verse 5**

The outer end of the notice: given a stale plugin state, `bp` writes the settled one-line notice to
**stderr** and leaves stdout untouched. Where the state comes from is hardcoded here — the state
reader returns "nothing known" until BIT-41.13 wires the real readers in — so this bar proves the
wording, the stream and the placement, and nothing else.

The notice is emitted from `Execute()`, not from `PersistentPreRunE`. Cobra 1.10.2's
`(*Command).execute` returns on `--version`, on `--help` and on a non-runnable root *before* it
walks the persistent pre-run hooks, so a hook would leave `bp -v` silent — and `bp -v` is the one
command an operator runs to ask "am I current?". `ExecuteContextC` returns the resolved command,
which BIT-41.10 uses to exempt `tui` and `serve mcp`.

## Scope
- `cmd/root.go` — `Execute()` splits into a thin wrapper plus `execute(ctx, root)`; a package-level
  `pluginState` var (the `serveTick` / `serveRunner` precedent in `cmd/serve.go`) defaulting to a
  reader that reports nothing; a `notice(installed, latest string) string` formatter.
- `cmd/cmd_test.go` — a `runSplit` helper that captures stdout and stderr **separately** and drives
  `execute`. The existing helpers merge both streams into one buffer, so they cannot see which
  stream a line went to; they keep calling `root.Execute()` and are unaffected by this bar.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestExecute_BehindPluginWritesNoticeToStderr`
     - **Behavior:** when the state reader reports the installed plugin is behind the latest, `bp`
       writes the settled notice to stderr, and stdout carries only the command's own output.
     - **Setup:** `initProject(t, "BIT")` and one created task; `pluginState` stubbed to return
       `("0.1.0", "0.2.0", true)`; run `task list` through `runSplit`.
     - **Assertions:** stderr is exactly
       `bp: bit plugin 0.1.0 → 0.2.0 available — run: claude plugin update bit@bit-pro --scope project\n`;
       stdout contains `BIT-1` and does not contain `bp: bit plugin`.
     - **Boundary:** `installed` strictly below `latest` — the one state of the ordering relation
       that fires. The stdout assertion is the boundary on the output stream: `bp instructions`
       guarantees `ID=$(bp task create …)` holds exactly the ID and that `task read --body`
       round-trips byte-for-byte, so a notice on stdout would break both.
   - [ ] `TestExecute_NoPluginStateIsSilent`
     - **Behavior:** when the reader reports nothing known, `bp` says nothing — the production
       default at this bar, and the "absent → silence" decision.
     - **Setup:** same project; `pluginState` left at its default.
     - **Assertions:** stderr is empty; stdout still contains `BIT-1`.
     - **Boundary:** the reader's `ok` flag false — the lower bound of what is known.
   - [ ] Confirm fails: compile error, `undefined: execute` / `undefined: pluginState`. Once it
         compiles but before the notice is written, the first test fails with an empty stderr.

2. **Implement (GREEN):**
   - [ ] `var pluginState = func() (installed, latest string, ok bool) { return "", "", false }`.
   - [ ] `notice(installed, latest string) string` returning the settled line, both versions
         substituted, no trailing newline.
   - [ ] `execute(ctx context.Context, root *cobra.Command) error`: call `root.ExecuteContextC(ctx)`;
         then if `pluginState()` reports `ok` and `installed != latest`, `fmt.Fprintln` the notice to
         the resolved command's `ErrOrStderr()`, falling back to `root` when the returned command is
         nil. Return the command's error unchanged — the notice must never alter the exit path.
   - [ ] `Execute()` becomes `ctx, stop := signalContext(); defer stop(); return execute(ctx, NewRootCmd())`.
   - [ ] `runSplit(t, args ...string) (stdout, stderr string, err error)` in `cmd_test.go`: build the
         root with the same fake runners the other helpers use, `SetOut`/`SetErr` to two separate
         buffers, and call `execute`.

`installed != latest` is deliberately the wrong comparison — BIT-41.9 contradicts it.

## Claude verifies
- [ ] `just test` passes.
- [ ] `just lint` passes.
- [ ] Every pre-existing test in `cmd` still passes untouched — they call `root.Execute()`, not
      `execute`, so the new stderr write cannot reach their merged buffers.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(version): write the stale-plugin notice to stderr`