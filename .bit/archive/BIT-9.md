---
id: BIT-9
title: TUI + init cleanup
status: done
---
## Why

Two unrelated rough edges make the tool briefly hostile to the human driving it. In the
TUI, once focus moves to the detail pane you can't quit — `q` and `ctrl+c` scroll the
viewport instead of exiting, so you have to arrow back to the list first, which nobody
expects. And re-running `bit init` in a project that's already initialized re-prompts for
the task ID prefix from scratch, with no hint of what it already is and no way to just keep
it — pressing enter errors out. Neither is a design question; both have a known right
answer and just aren't done. Grouped here as one cleanup pass.

## Summary

Two independent fixes. First: let the quit keys exit the TUI even when the detail pane is
focused — the detail pane quits exactly like the list does. Second: when `bit init` runs
in an already-initialized project, read the existing config, show its prefix as the default
in the prompt, and let a bare enter reuse it.

## Phases

- [x] Phase 1 — Quit from anywhere in the TUI: the quit keys exit the TUI whether focus is
  on the list or the detail pane, instead of being swallowed by viewport scrolling. Focus
  doesn't change what quits — `q`, `ctrl+c`, and `esc` quit from the detail pane exactly as
  they do from the list.
  Touches: the key handling in the TUI update loop (`tui/model.go`, the `detailFocused`
  branch that forwards every message to the viewport).
- [x] Phase 2 — Re-run init keeps the existing prefix: running `bit init` where a config
  already exists offers that prefix as the default (`Task ID prefix (BIT): `) and treats a
  bare enter as "keep it," instead of demanding a fresh value.
  Touches: the prompt in `bit init` (`cmd/init.go`) reading the existing config
  (`task.Config()` in `task/config.go`).

## Risks & unknowns

- **Resolved:** `esc` quits the TUI from the detail pane, same as `q` and `ctrl+c` — the
  detail pane's quit behavior matches the list's exactly. Focus never changes what quits.
  (Consistent with board mode, which already treats all three as quit — `TestUpdate_BoardQuits`.)

- **Unknown:** How does the init prompt read in a test? The prompt is interactive
  (`bufio` over `cmd.InOrStdin()`), and re-run needs a pre-existing config on disk.
  **Resolve by:** follow the existing `cmd/init_test.go` pattern for feeding stdin and
  seeding a temp `.bit/`.
  **De-risk before planning?** No — it's a test-setup detail, not a design risk.