---
id: BIT-41.10
title: The full-screen commands are exempt
status: todo
approved: true
phase: 5
phase_label: stale notice
---
## **Verse 5**

`bp tui` and `bp serve mcp` own their output stream for the life of the process — the TUI renders
full-screen, the MCP server speaks a protocol over stdio — so an advisory line written around them
lands in the middle of a render or a frame. This bar carves both out, and the same predicate later
gates the detached refresh in BIT-41.14, so those two commands end up wholly untouched by this
verse.

The exemption is marked on the commands themselves rather than matched by name in `execute`, so a
future full-screen command opts out where it is defined instead of in a list somewhere else.

## Scope
- `cmd/root.go` — a `quietAnnotation` const and a `suppressed(cmd *cobra.Command) bool`; `execute`
  returns before the notice when it reports true.
- `cmd/tui.go`, `cmd/serve_mcp.go` — carry the annotation.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestSuppressed_FullScreenCommands` (table-driven)
     - **Behavior:** the notice is suppressed for exactly the two commands that own their output
       stream, and for nothing else — including the exempt command's own sibling.
     - **Setup:** build the root with `newRootCmd`, resolve each command with `root.Find(args)`:
       `["tui"]`, `["serve", "mcp"]`, `["serve", "daemon"]`, `["task", "list"]`, and the root itself.
     - **Assertions:** `suppressed` is true for `tui` and `serve mcp`; false for `serve daemon`,
       `task list` and the root.
     - **Boundary:** the exemption set at both edges — both members, plus `serve daemon` as the
       sibling that proves the carve-out is not "the whole `serve` subtree", plus an ordinary
       command and the root as controls.
   - [ ] `TestExecute_SuppressedCommandWritesNoNotice`
     - **Behavior:** a suppressed command produces no notice even when the state is stale.
     - **Setup:** `pluginState` stubbed to `("0.1.0", "0.2.0", true)`; call `execute` with a root
       whose resolved command carries the annotation but a no-op `RunE`, registered by the test —
       neither `tui` nor `serve mcp` can be run inside a test, so the wiring is exercised through a
       stand-in that carries the same annotation.
     - **Assertions:** stderr is empty.
     - **Boundary:** the annotation present vs. absent — BIT-41.8's stale test is the absent case.
   - [ ] Confirm fails: `undefined: suppressed`; then, before the annotations are added, the `tui`
         and `serve mcp` rows fail.

2. **Implement (GREEN):**
   - [ ] `const quietAnnotation = "bit.quiet"` and
         `func suppressed(cmd *cobra.Command) bool` — nil-safe, reads
         `cmd.Annotations[quietAnnotation]`.
   - [ ] Add `Annotations: map[string]string{quietAnnotation: "true"}` to the `tui` command and to
         the `serve mcp` command.
   - [ ] `execute`: return the command's error without touching the notice when `suppressed(cmd)`.

## Claude verifies
- [ ] `just test` passes.
- [ ] `just lint` passes.

## User verifies
- [ ] none — deterministic. The observable half of this (no stray line in a TUI frame) needs a
      genuinely stale plugin, so it rides with the whole-slice check on BIT-41.14.

## Commit (user)
`feat(version): exempt the full-screen commands from the notice`