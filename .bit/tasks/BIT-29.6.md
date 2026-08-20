---
id: BIT-29.6
title: A typed code beats the default and is stored uppercase
status: todo
approved: true
phase: 2
phase_label: add
---
## **Verse 2**

A code typed at the prompt beats the offered default and is stored uppercase. Contradicts BIT-29.4,
which returns the typed string verbatim — so a lowercase code would be stored and printed lowercase
while `bp init` would have uppercased the same characters.

## Scope
- `task/store.go` — rename `normalizeID` to `NormalizeID`
- `task/config.go` — the two call sites follow the rename
- `cmd/add.go` — normalise the code
- `cmd/add_test.go` — the test

The scope decides the code is normalised "exactly as `bp init` normalises the prefix
(`task.normalizeID`)", so that "the two can never disagree about casing". A second `strings.ToUpper`
in `cmd/add.go` would satisfy the letter and not the reason — two routines that agree today. So the
existing one is exported and shared: `normalizeID` becomes `NormalizeID`, and its eight call sites
inside `task/` follow the rename mechanically.

## TDD cycle

1. **Write test (RED):** `cmd/add_test.go`
   - [ ] `TestAddCmd_UppercasesATypedCode` (table-driven over the typed input: `"foo"` and `"FOO"`)
     - **Behavior:** a project code and a task prefix are the same identifier wearing two names, so
       they cannot disagree about casing — whatever the operator types, `bp list` and the task IDs
       in `.bit/` read the same.
     - **Setup:** the same `home`/`XDG_DATA_HOME`/`initProject(t, "BIT")` preamble, and
       `want, err := filepath.Abs(".")` after the chdir. Run
       `runWithStdin(t, tt.typed+"\n", addCmdUse, ".")`.
     - **Assertions:** in both rows, `out == "Project code (BIT): added FOO " + want + "\n"` — so the
       printed code is the *stored* one, per the scope's Decision, not an echo of what was typed. And
       `ListProjects` has one row with `Code == "FOO"`.
     - **Boundary:** the typed-input branch of the prompt — non-empty input, which must win over the
       non-empty default the same setup offers. Casing is exercised on both sides of
       `strings.ToUpper`'s fixed point: an input that must change and one that must not.
   - [ ] Confirm fails: the `"foo"` row fails on both assertions — `out` reads `added foo …` and the
         stored `Code` is `"foo"`. The `"FOO"` row passes, which is expected and is why the lowercase
         row is in the table.

2. **Implement (GREEN):**
   - [ ] `task/store.go`: rename `normalizeID` to `NormalizeID` and update its call sites in
         `store.go` and `config.go`. Package-internal rename plus one new exported name — no
         behaviour change, so `task`'s own tests stay green untouched.
   - [ ] `cmd/add.go`: `code = task.NormalizeID(code)` after `readProjectCode` returns, before both
         the insert and the printed line.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. Verse 2's whole-slice check lands on BIT-29.8.

## Commit (user)
`feat(add): store a typed project code uppercased`