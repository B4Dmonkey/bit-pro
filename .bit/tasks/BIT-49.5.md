---
id: BIT-49.5
title: retro_write stores a proposal record under a code-prefixed name and replaces a re-run
status: todo
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

retro's output moves from a plain file to a record in the shared `retro/`. Two writes, one name that lacks the code and one that already starts with it, can't both pass with a hardcoded prefix. That forces the scope's prefix rule.

## Scope
- new `task/retro.go`:
  - const `retroSubdir = "retro"`; `retroDir()` = `filepath.Join(s.data, retroSubdir)`. An empty `s.data` errors, as in BIT-49.1.
  - unexported `retroRecord`: `project, name, created_at, updated_at, commits, content`. Every field is written, `commits` is `[]` when empty, and `content` is `<name>.md`.
  - unexported `retroName(code, name string) string`: return `name` if `strings.HasPrefix(name, code+"-")`, else `code + "-" + name`. The match is case-sensitive.
  - `func (s *Store) WriteRetro(name, body string, head Commit) (string, error)`. It validates `name` like `validateTopic` (not empty, no `..`, no separator), applies `retroName`, and builds paths with `pathologize.Join`. It writes the `.md`, then the `.json`. On an existing record it keeps `created_at` and sets `commits = appendCommit(existing.Commits, head)` (BIT-46.16). It bumps `updated_at` with `s.now`, and returns the stored name.
  - **Seam for BIT-47:** it passes the real `head` instead of `Commit{}`. BIT-49.14's migrate reuses `WriteRetro`.
- `cmd/serve_mcp.go`: const `retroWriteTool = "retro_write"`; `retroWriteInput{Name, Body string}`, `retroWriteOutput{Name string \`json:"name"\`}`. The handler passes `task.Commit{}`. The description says it writes one proposals record for the current project, that the server prefixes the project code unless the name already starts with it, that writing an existing name replaces it, and that it returns the stored name. It names no path.
- `cmd/testconst_test.go`: `testCodePrefixed = "prefixes the project code"`.
- Test files touched: new `task/retro_test.go`, new `cmd/serve_mcp_retro_test.go`, and `cmd/serve_mcp_test.go` (already converted by BIT-46.9).

## TDD cycle

0. **Restructure:** none. `cmd/serve_mcp_test.go` was converted by BIT-46.9; the other two files are new.

1. **Write test (RED):**
   - [ ] `TestRetroWriteHandler/stores a proposal under the code-prefixed name` (`cmd/serve_mcp_retro_test.go`)
     - **Behavior:** a proposals record is stored in the shared folder under a name no other project can collide with.
     - **Setup:** sandbox; `dir` registered as `BIT` (seeded as in BIT-49.3); `retro_write {name: "album-proposals", body: "## Proposal 1\n\n**Pattern:** ...\n"}`.
     - **Assertions:** `name == "BIT-album-proposals"`. `$XDG_DATA_HOME/bit/retro/BIT-album-proposals.md` holds the body. The `.json` has `"project": "BIT"`, `"name": "BIT-album-proposals"`, `"commits": []` and `"content": "BIT-album-proposals.md"`.
     - **Boundary:** a name with no code prefix.
   - [ ] Confirm fails: the tool isn't registered.

2. **Implement (GREEN):**
   - [ ] `task/retro.go` and the tool.

3. **More tests (RED → GREEN):**
   - [ ] `TestRetroName` (table, `task/retro_test.go`): `("BIT", "album-proposals") → "BIT-album-proposals"`; `("BIT", "BIT-49-proposals") → "BIT-49-proposals"`; `("ACME", "ACME-1-ACME-4-proposals") → unchanged`; `("BIT", "bit-49-proposals") → "BIT-bit-49-proposals"`. *Boundary:* already prefixed or not, and case sensitivity. Fixtures use neutral codes, never a client project's.
   - [ ] `TestStoreWriteRetro/replaces a re-run and keeps created_at` (`task/retro_test.go`; store-method tests are `TestStore<Method>`, as BIT-46's `TestStoreWriteResearch` and `TestStoreAddNote`): write at T1, then write the same name at T2 with a new body. The `.md` holds the new body, `created_at == T1` and `updated_at == T2`. *Boundary:* the second write.
   - [ ] `TestStoreWriteRetro/refuses a path-like name` (table): `""`, `"../x"`, `"a/b"` → error, and no file is written. *Boundary:* untrusted name.
   - [ ] A row in `TestMCPToolDescriptions/carry the domain`: `retro_write` wants `testCodePrefixed`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): retro_write stores proposals as records in the shared retro folder`