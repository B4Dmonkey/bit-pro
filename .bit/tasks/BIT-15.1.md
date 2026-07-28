---
id: BIT-15.1
title: 'Tests drive Use: bp in root command'
status: done
phase: 1
phase_label: Binary runs as bp
---
## **Verse 1**

Update the two root-command tests to assert `"bp"` — they fail because `Use: "bit"` — then flip `Use` and the Justfile output name to make them pass.

## Scope
- `cmd/root_test.go` — update string assertions from `"bit"` to `"bp"` (tests + one log message)
- `cmd/root.go` — `Use: "bit"` → `Use: "bp"`; update Short description if it contains the literal word `bit`
- `Justfile` — `-o "$dir/bit"` → `-o "$dir/bp"`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestRootCmd_Help` — change `strings.Contains(out, "bit")` to `strings.Contains(out, "bp")`, and update the error string literal to `"bp"`
     - **Behavior:** help output includes the binary name as `bp`
     - **Setup:** `mustRun(t, "--help")` — no project dir needed
     - **Assertions:** `strings.Contains(out, "bp")` is true
     - **Boundary:** exact string match on the binary name — the one value cobra prints from `Use`
   - [ ] `TestRootCmd_Version` — change `want := "bit version " + version + "\n"` to `"bp version " + version + "\n"`
     - **Behavior:** `--version` prints `bp version <tag>`
     - **Setup:** `mustRun(t, "--version")`
     - **Assertions:** `out == "bp version dev\n"`
     - **Boundary:** exact equality — verifies the full version string, not just a substring
   - [ ] Also update `cmd_test.go:32`: `t.Fatalf("bit %s returned error: …"` → `"bp %s returned error: …"` (log message, not an assertion; include it to stay consistent)
   - [ ] Confirm fails: `TestRootCmd_Version` → `got "bit version dev\n", want "bp version dev\n"`; `TestRootCmd_Help` → `help output missing command name "bp"`

2. **Implement (GREEN):**
   - [ ] `cmd/root.go`: `Use: "bit"` → `Use: "bp"`
   - [ ] `cmd/root.go`: update `Short` if it contains the literal `bit` (check the string — rename in place if so)
   - [ ] `Justfile`: `-o "$dir/bit"` → `-o "$dir/bp"`

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] `just install` then `bp --help` — output shows `bp` as the command name, not `bit`
- [ ] `bp --version` — prints `bp version <tag>`

## Commit (user)
`feat(cmd): rename binary from bit to bp`