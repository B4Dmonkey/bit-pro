---
id: BIT-48.4
title: A failed wiring step keeps the project registered, prints the claude commands and exits non-zero
status: todo
approved: true
phase: 1
phase_label: bp add sets bit up for the whole machine
---
## **Verse 1**

A re-run of `bp add` won't re-wire a registered project, so when a step fails the only recovery is the commands themselves. A runner that fails the install forces `ensureGlobalWiring` to print them; it already returns the error, so the project-stays-registered and non-zero assertions pin what BIT-48.3 built.

## Scope
- `cmd/wiring.go` `ensureGlobalWiring`: on a non-nil error from `EnsureGlobal`, print to `cmd.ErrOrStderr()`:
  ```
  Run these to finish setting up bit in Claude Code:
    claude plugin marketplace add B4Dmonkey/bit-pro
    claude plugin marketplace update bit-pro
    claude plugin install bit@bit-pro --scope user
    claude mcp add -s user bit -- bp serve mcp
  ```
  (each line `"  " + strings.Join(argv, " ")` over `claude.GlobalWiring()`, all four), then return the error unchanged. `main.go` prints it as `Error: claude wiring step N of 4: ...` and exits 1.
- `cmd/add_test.go`: the new case joins `TestAddCmd` as a `t.Run` subtest (its failing runner and multi-step assertions differ from the other cases). The file is already in the one-`TestAddCmd` shape from BIT-46, so no conversion step.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAddCmd/wiring failure keeps the project and prints the commands`
     - **Behavior:** if Claude setup fails, the project is still registered, the operator gets the exact commands to run by hand, and the command fails so scripts notice.
     - **Setup:** temp `HOME`, `XDG_DATA_HOME=""`, fresh `t.TempDir()`; runner records calls and returns `boom := errors.New("plugin bit not found")` when `args[1] == "install"`; stdin `FOO\n`; `bp add <dir>`.
     - **Assertions:** `errors.Is(err, boom)`; output contains `added FOO <canonical dir>`; output contains `Run these to finish setting up bit in Claude Code:` followed by the four `  claude ...` lines, in order; exactly three calls recorded; `ListProjects` returns one row, code `FOO`; a second `bp add <dir>` prints `already added` and records no calls.
     - **Boundary:** failure at step 3 of 4 × a fresh path — the registration survives and a re-run doesn't retry the wiring.
   - [ ] Confirm fails: the output has no `Run these` block.

2. **Implement (GREEN):**
   - [ ] The print-on-error branch in `ensureGlobalWiring`.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none here (the verse's end-to-end check is on its last bar)

## Commit (user)
`feat(bit): print the claude commands when bp add's wiring fails`