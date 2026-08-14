---
id: BIT-21.17
title: The prefix stays canonical on both sides of config.toml
status: todo
phase: 2
phase_label: Recurrence
---
## **Verse 2**

The last gap, and the one no migration script can close: the prefix a project is born with. A
project initialised with a lowercase prefix after the migration mints lowercase IDs forever, in a
project the script never saw. Both directions are covered here because they are one property —
the prefix is uppercase regardless of what was typed or what is on disk.

`cmd/init.go` needs no change of its own: it passes the typed prefix to `SaveConfig` and reads
the existing one back through `Config`, so normalizing in `task/config.go` covers both the
`--prefix` flag and the interactive prompt.

## Scope
- `task/config.go` — uppercase `Prefix` in `SaveConfig` (write) and in `Config` (read).
- `cmd/init_test.go`, `cmd/task_create_test.go` — the new tests.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestInitCmd_UppercasesTheStoredPrefix`
     - **Behavior:** a project is born canonical whatever the operator types.
     - **Setup:** table-driven over the two input routes the command already supports —
       `init --prefix foo`, and `init` with `foo\n` on stdin. Then `task create "first"`.
     - **Assertions:** `task.New(".bit").Config()` returns `Prefix == "FOO"`; the file
       `.bit/config.toml` itself contains `FOO`; `task create` prints `FOO-1` and writes
       `.bit/tasks/FOO-1.md`.
     - **Boundary:** an all-lowercase prefix, the input furthest from canonical, exercised through
       both entry points. Asserting the minted ID as well as the config proves the normalization
       reaches the thing that actually matters.
   - [ ] `TestTaskCreateCmd_UppercasesACorruptPrefixOnRead`
     - **Behavior:** a hand-edited config cannot reintroduce lowercase IDs.
     - **Setup:** `initProject(t, "BIT")`; overwrite `.bit/config.toml` with `prefix = "bit"`.
       Run `task create "first"`.
     - **Assertions:** stdout is `BIT-1`; `.bit/tasks/BIT-1.md` exists with `id: BIT-1`.
     - **Boundary:** the read side in isolation — the write side is already correct, so only
       normalizing on read can make this pass.
   - [ ] Confirm fails: the first stores `prefix = "foo"` and mints `foo-1`; the second mints
     `bit-1`

2. **Implement (GREEN):**
   - [ ] Uppercase `cfg.Prefix` in `SaveConfig` before marshalling, and in `Config` after
     decoding.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes
- [ ] `bash update/normalize_test.sh` still exits 0 — the migration and the code agree on the
  same canonical form

## User verifies
- [ ] Whole slice: in a throwaway directory run `bp init --prefix foo`, then reproduce all four
  symptoms with lowercase arguments — `bp task create "t"`, `bp task create "b" --parent foo-1`,
  `bp task complete foo-1` with that bar still `todo`, and `bp feedback add foo-1 -d "x"` twice.
  Observe: the prefix stored is `FOO`, the second create returns `FOO-1.2` and leaves `FOO-1.1`
  intact, the complete refuses with `unfinished bars FOO-1.1`, and `.bit/feedback/` ends up with
  two files whose contents both survive.

## Commit (user)
`fix(task): keep the config prefix canonical on read and write`