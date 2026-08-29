---
id: BIT-43.10
title: bp instructions and the embedded contract retire
status: todo
phase: 4
phase_label: Domain on descriptions
---
## **Verse 4**

With the domain relocated onto the tool descriptions and no skill calling it, `bp instructions`
teaches nothing that is not now enumerable in the tool list. This bar deletes the command and the
158-line contract it prints.

**A note on the reference doc:** `mcp-notes.md` step 5 says `bp instructions` retires "with the
other Claude-only commands in step 7", and marks step 7 optional. **The scope's Verse 4 governs and
supersedes that** — the command goes now, not maybe-later. Update the step 5 and step 7 bullets in
`mcp-notes.md` as part of this bar so the note stops contradicting the work.

**What forces this bar:** nothing in Verses 1–3 could delete the command, because a skill still
calling it would break. The previous two bars removed the last reason it exists.

## Scope
- `cmd/instructions.go` — deleted
- `cmd/instructions_test.go` — deleted (both its tests describe a command that is gone)
- `cmd/root.go:145` — `rootCmd.AddCommand(newInstructionsCmd())` removed
- `assets/bit-cli.md` and `assets/assets.go` — the whole `assets/` package deleted. `assets.go`'s
  only content is `//go:embed bit-cli.md`, so it cannot survive the markdown file. Verified: the
  only importers of `github.com/B4Dmonkey/bit-pro/assets` are `cmd/instructions.go` and
  `cmd/instructions_test.go`, both of which this bar deletes — nothing else breaks.
- `cmd/root_test.go` — new test
- `mcp-notes.md` — steps 5 and 7 corrected

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestRootCmd_HasNoInstructionsCommand` in `cmd/root_test.go`
     - **Behavior:** the contract command is gone from the CLI surface, so an operator (or a stale
       skill) invoking it gets a legible failure rather than 158 lines of retired guidance.
     - **Setup:** `_, err := run(t, "instructions")` — the helper at `cmd/cmd_test.go:14`, which
       builds a root command and returns its combined output and error.
     - **Assertions:** `err` is non-nil. Confirmed empirically that cobra produces
       `unknown command "instructions" for "bp"` with exit 1 for an unregistered subcommand on
       this root (`SilenceUsage`/`SilenceErrors` are set but the error still returns), so also
       assert the message contains `unknown command`.
     - **Boundary:** `instructions` is the removed member of the registered-subcommand set — the
       transition from registered to absent. The neighbouring rows in `newRootCmd` stay registered,
       which the existing root tests already cover.
   - [ ] Confirm fails: the command is still registered, so `err` is nil and the assertion reports
         a nil error where one was wanted. **Not** a compile failure — if the package does not
         build, something was deleted before the test was written.

2. **Implement (GREEN):**
   - [ ] Delete `cmd/instructions.go` and `cmd/instructions_test.go`
   - [ ] Remove the `newInstructionsCmd()` registration from `cmd/root.go`
   - [ ] Delete the `assets/` directory (`assets.go` and `bit-cli.md`)
   - [ ] Correct `mcp-notes.md`: step 5's third bullet no longer defers to step 7, and step 7's
         inventory no longer lists `instructions` among the commands still to be removed

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes
- [ ] `go build ./...` succeeds with no reference to the deleted `assets` package
- [ ] `just install`

## User verifies
- [ ] `bp --help` no longer lists `instructions`, and `bp instructions` prints
      `Error: unknown command "instructions" for "bp"`.
- [ ] **Whole slice:** open the tool list in a session wired to the `bit` MCP server and read
      `task_read`'s and `task_update`'s descriptions. Everything the retired contract taught about
      tracks, bars, approval and rollup is legible there — the verse's capability is that the
      domain survived the deletion.

## Commit (user)
`refactor(cli): retire bp instructions and the embedded contract`