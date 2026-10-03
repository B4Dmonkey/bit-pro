---
id: BIT-46.5
title: bp add registers the canonical path and refuses an unregistered .bit/ folder
status: todo
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

`bp add` becomes v2 registration. It checks "already registered" first, then refuses a folder that still has a v1 `.bit/` (pointing to `bp migrate`), and stores the canonical path so the resolver and the UNIQUE constraint agree. The rewritten add tests force each branch. The order stays wire, then register (`cmd/add.go:66-73`); BIT-48 reverses it.

## Scope
- `cmd/add.go` `newAddCmd`:
  1. `path, err := project.CanonicalPath(args[0])` (replaces `filepath.Abs`).
  2. `db.Open`, `project.Load`. If any row's `Path` `strings.EqualFold`s `path` (exact path, not containment), print `already added` (v1 text) and return nil. No wiring.
  3. If `<path>/.bit` exists, return `fmt.Errorf("%s: %w", path, project.ErrNeedsMigrate)`. Check only this path; the walk-up is for the resolver's message.
  4. `readProjectCode(cmd)`: the `existing` parameter and the `.bit/config.toml` default go away. The prompt is plain `Project code: `.
  5. `project.ValidateCode`, then `writeClaudeWiring(cmd, run, path)` (the `.bit` stat guard goes; step 3 already refused), then `CreateProject{Path: path, Code: code}`.
  - Keep the three outcomes as separate early returns (registered no-op / refusal / fresh), so BIT-48 can insert its wiring change and BIT-46 Verse 3 can add revive.
- `db/queries/projects.sql` — delete `ProjectExists` (no caller left); delete its stale `db/orm` output by hand.
- `cmd/add.go` `Short` — leave as BIT-45 reworded it.
- `cmd/add_test.go` — rewrite the `TestAddCmd` subtests that use `initProject` (which creates `.bit/` via `bp init` today) to use a fresh `t.TempDir()`; expected paths become `project.CanonicalPath(dir)` (`/private/var/...`; `cmd/add_test.go:51-58` compared to `filepath.Abs`). The file was converted to `TestAddCmd` subtests in BIT-46.3.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAddCmd/refuses an unregistered folder with bit` (subtest; replaces `TestAddCmd/enrolls using the bit prefix`)
     - **Behavior:** a v1 project can't be registered by `bp add`; the operator is sent to `bp migrate`.
     - **Setup:** sandboxed `HOME`, `XDG_DATA_HOME=""`; `t.Chdir(dir)`; `os.MkdirAll(".bit", 0o755)` + write `.bit/config.toml` `prefix = "BIT"`; recording runner; stdin `"BIT\n"`; `bp add .`.
     - **Assertions:** `errors.Is(err, project.ErrNeedsMigrate)`; output does not contain `Project code`; `ListProjects` empty; runner recorded no calls.
     - **Boundary:** `.bit/` present × not registered — the refusal branch.
   - [ ] Confirm fails: today `bp add` reads the config default and registers `BIT`.

2. **Implement (GREEN):**
   - [ ] Steps 1–3 above.

3. **More tests (RED → GREEN):**
   - [ ] `TestAddCmd/uppercases a typed code` (fresh dir; its rows stay a table): output `"Project code: Bringing the bit plugin current...\nRegistering bit MCP server...\nbit MCP server registered (local scope).\nadded FOO <canonical dir>\n"` (the order is still wire, then register, so `writeClaudeWiring`'s three lines print between the prompt and `added`; BIT-48.3 replaces this expectation); stored `Path == canonical dir`. *Boundary:* the stored path is the `EvalSymlinks` form even though `.` was typed under `/var`. Forces steps 1 and 4.
   - [ ] `TestAddCmd/skips a path already enrolled` (fresh dir): `mkdir <tmp>/Repo`; `bp add <tmp>/Repo` (stdin `FOO`), then `bp add <tmp>/repo` → output `already added\n`, one row, and the second run made no runner calls. *Boundary:* the same folder typed in a different case (APFS); exact-case is the trivial case.
   - [ ] `TestAddCmd/allows a path inside a registered path` (new subtest): `bp add <tmp>` (`ACME`), then `bp add <tmp>/api` (`API`) → two rows. *Boundary:* nested registration (`acme/` holding repos).
   - [ ] `TestAddCmd/initialises a project without bit`: keep its wiring assertions; update the path expectation to canonical.
   - [ ] `TestAddCmd/refuses a code already taken` (new subtest): `bp add <tmp>/a` (`FOO`), then `bp add <tmp>/b` (`FOO`) → non-nil error, one row. *Boundary:* the UNIQUE `code` surfaces through `bp add` (message is the wrapped `registering <path>: ...` error from today's code).

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): bp add registers canonical paths and refuses v1 .bit/ folders`