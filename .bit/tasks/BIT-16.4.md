---
id: BIT-16.4
title: bp answers for its own contract
status: done
phase: 2
phase_label: The plugin ships the skills
---
## **Verse 2**

The skills are about to stop reading `.claude/bit-cli.md` and start asking the binary for the
contract, so the binary has to be able to answer first. Porting the skills ahead of this would
ship four skills whose first instruction is a command that does not exist.

## Scope
- `cmd/instructions.go` — new command; prints the embedded contract to `cmd.OutOrStdout()`.
- `cmd/instructions_test.go` — new.
- `cmd/root.go` — register it on the root command.

`assets/bit-cli.md` and its `//go:embed` stay exactly as they are. Verse 4 narrows the embed
directive to this one file; it does not remove it.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestInstructionsCmd_PrintsContract`
     - **Behavior:** `bp instructions` emits the CLI contract byte-for-byte from the copy compiled
       into this binary, so a skill that asks the binary for its contract cannot be told something
       the binary does not implement. This is the whole point of the command over a shipped file.
     - **Setup:** `out := mustRun(t, "instructions")`; `want, err := assets.FS.ReadFile("bit-cli.md")`.
     - **Assertions:** `out == string(want)` — exact equality, no summary header and no added
       trailing newline. Follow `task read --body`'s `fmt.Fprint` style so the bytes round-trip.
     - **Boundary:** the full embedded document rather than a substring — a `strings.Contains`
       assertion would pass against a hardcoded first paragraph, and exact equality is what makes
       reading the embed the only way to green.
   - [ ] Confirm fails: `unknown command "instructions" for "bp"`.

2. **Implement (GREEN):**
   - [ ] `newInstructionsCmd()` with `Args: cobra.NoArgs`, `RunE` reading
     `assets.FS.ReadFile("bit-cli.md")` and writing it to `cmd.OutOrStdout()`; wrap a read error
     with context.
   - [ ] `rootCmd.AddCommand(newInstructionsCmd())` in `NewRootCmd`.

3. **More tests (RED → GREEN):**
   - [ ] `TestInstructionsCmd_RejectsArgs`
     - **Behavior:** the command takes no arguments, so a mistyped invocation fails loudly instead
       of silently printing the contract and ignoring what was asked for.
     - **Setup:** `run(t, "instructions", "garbage")`.
     - **Assertions:** returned error is non-nil.
     - **Boundary:** argument count == 1, one above the only valid count of 0.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just run instructions | head -3` prints the contract's opening lines

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(cmd): add bp instructions to print the CLI contract`