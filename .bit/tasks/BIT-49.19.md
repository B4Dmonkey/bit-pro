---
id: BIT-49.19
title: migrate refuses a .bit/ whose IDs aren't uppercase
status: todo
phase: 2
phase_label: bp migrate
---
## **Verse 2**

`task.Parse` uppercases `id` and `order` as it reads (`task/task.go:48-51`), so a mixed-case v1 store would collide in uppercase dirs on case-insensitive APFS. Inside frontmatter, a lowercase `id:` or `order:` entry already fails BIT-49.18's round trip, because `Bytes()` writes it back uppercase. The carriers outside frontmatter still need their own check: file names, feedback names, research dirs and the raw prefix.

## Scope
- `migrate/check.go`: unexported `badIDs(src, rawPrefix string) []string`. It checks:
  - the raw `prefix` from `config.toml` (BIT-49.10 kept it raw), which must equal `strings.ToUpper(rawPrefix)`;
  - task file stems in `tasks/`, `completed/` and `archive/tasks/`;
  - the track part of each feedback file name;
  - each `research/<dir>` name.

  Each ID-like value must be uppercase. Each offender becomes `<rel>: <value>`.
- `migrate.Run`: runs after `badTaskFiles`, before staging → `ErrIDs` (`errors.New(".bit/ holds IDs that aren't uppercase")`). The wrapped message adds `run v1's update/normalize.sh to uppercase them`.
- Test files touched: `migrate/check_test.go` and `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/refuses ids that are not normalised`
     - **Behavior:** a store that would collide on a case-insensitive disk never gets copied.
     - **Setup:** sandbox. `config.toml` has `prefix = "BIT"`. `tasks/bit-2.md` holds `Bytes()` output for `ID: "BIT-2"`, so it round-trips and only its file name is lowercase. Then `bp migrate`.
     - **Assertions:** `errors.Is(err, migrate.ErrIDs)`. The text names `tasks/bit-2.md`. Nothing is registered or written.
     - **Boundary:** a lowercase carrier that both `task.Parse` and the round trip would miss.
   - [ ] Confirm fails: the migration succeeds and stores `BIT-2`.

2. **Implement (GREEN):**
   - [ ] `badIDs`, and the check in `Run`.

3. **More tests (RED → GREEN):**
   - [ ] `TestBadIDs` (table):
     - `prefix = "bit"` → `config.toml: bit`.
     - `feedback/bit-1-001.md` → listed.
     - `research/bit-1/` → listed.
     - a fully uppercase `BIT-*` store → none.

     *Boundary:* each carrier.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate refuses IDs that aren't uppercase`