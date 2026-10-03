---
id: BIT-46.3
title: bp add refuses a code that can't name a store dir
status: done
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

A code names `~/.local/share/bit/<CODE>/` and prefixes every ID, and v1's `bp add` only uppercases it (`cmd/add.go:84-102`, `task.NormalizeID`). This bar adds the code rule as an exported function in a new `project` package, which BIT-49's migrate reuses.

## Scope
- new `project/code.go` (package `project`):
  - `func ValidateCode(code string) (string, error)` — trims, uppercases, then requires `^[A-Z][A-Z0-9]*$` and rejects `FEEDBACK` and `RETRO`. Returns the normalized code.
  - sentinels `ErrInvalidCode`, `ErrReservedCode`; returned wrapped as `fmt.Errorf("project code %q: %w", code, Err...)`.
  - **Seam for BIT-49:** migrate calls `project.ValidateCode` on a v1 prefix.
- `cmd/add.go` — replace the empty check (`"project code cannot be empty"`) + `task.NormalizeID(code)` with `project.ValidateCode`; an invalid code returns the error before any wiring or insert.
- `project/code_test.go`, `cmd/add_test.go`.

## TDD cycle

0. **Convert existing tests (pure restructure, same cases and assertions, still green):**
   - [ ] `cmd/add_test.go`: every test becomes a subtest of `TestAddCmd`: `TestAddCmd/enrolls using the bit prefix`, `TestAddCmd/initialises a project without bit`, `TestAddCmd/uppercases a typed code` (its `lowercase`/`uppercase` rows stay a table inside it), `TestAddCmd/rejects an empty code`, `TestAddCmd/skips a path already enrolled`.

1. **Write test (RED):**
   - [ ] `TestValidateCode` (table, `project/code_test.go`), one row per case:
     - **Behavior:** only codes that are safe as a dir name and an ID prefix get through, normalized to uppercase.
     - **Rows / Assertions:** `lowercase is uppercased` `"bit"` → `"BIT", nil`; `digits after a letter` `"ACME2"` → `"ACME2", nil`; `surrounding space is trimmed` `"  ex "` → `"EX", nil`; `empty` `""` → `ErrInvalidCode`; `leading digit` `"2BIT"` → `ErrInvalidCode`; `hyphen` `"BIT-PRO"` → `ErrInvalidCode`; `underscore` `"BIT_1"` → `ErrInvalidCode`; `traversal` `"../X"` → `ErrInvalidCode`; `reserved feedback in lowercase` `"feedback"` → `ErrReservedCode`; `reserved retro in mixed case` `"Retro"` → `ErrReservedCode` (check with `errors.Is`).
     - **Boundary:** first char letter vs digit; `-`/`_`/`/` outside the class; reserved names checked *after* uppercasing (lowercase `feedback` is still reserved, because APFS is case-insensitive); empty string is the lower bound.
   - [ ] Confirm fails: package `project` doesn't exist.

2. **Implement (GREEN):**
   - [ ] `ValidateCode` with a package-level `regexp.MustCompile`.

3. **More tests (RED → GREEN):**
   - [ ] `TestAddCmd/refuses an invalid code` (subtest, `cmd/add_test.go`)
     - **Behavior:** `bp add` won't register a code that fails the rule, and runs no wiring.
     - **Setup:** sandboxed `HOME`, `XDG_DATA_HOME=""`; `t.Chdir(t.TempDir())` (no `.bit/`); recording runner; stdin `"bit-pro\n"`; `bp add .`.
     - **Assertions:** `errors.Is(err, project.ErrInvalidCode)`; `ListProjects` empty; runner recorded no calls; no `.claude/settings.json` written.
     - **Boundary:** a code with `-` — refused before side effects.
   - [ ] Update `TestAddCmd/rejects an empty code`: it asserts the exact string `"project code cannot be empty"`; assert `errors.Is(err, project.ErrInvalidCode)` instead (empty is now a rule violation, not its own message). `TestAddCmd/uppercases a typed code` stays green unchanged.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): validate project codes (letters and digits, reserved names)`