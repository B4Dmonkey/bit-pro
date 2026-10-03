---
id: BIT-46.14
title: Tracks and bars carry branch and commit exactly as the caller passes them
status: todo
phase: 2
phase_label: records are JSON metadata plus markdown
---
## **Verse 2**

Tracks and bars carry a `branch` and a single `commit`. This track leaves them empty. The store writes whatever its caller passes, so BIT-47 (bar commit hash, then track landing commit) and BIT-49 (HEAD at migration) only pass values in. A test that passes values through `Create` and `Update` forces the parameters, and a test of the default forces `""`.

## Scope
- `task/task.go`: `Task` gains `Branch string`, `Commit string`, tagged `yaml:"-"`.
- `task/record.go`: `branch`, `commit`, always written (`""` when empty).
- `task/store.go`:
  - `CreateParams` gains `Branch, Commit string`, copied onto the new task.
  - `Patch` gains `Branch, Commit *string` (nil leaves the stored value alone, like the other fields).
  - `Update`: setting only `Branch`/`Commit` does **not** revoke approval. They record where work landed, not what was reviewed, so they stay out of `contentChanged`.
  - `Save` writes `t.Branch`/`t.Commit` as given (BIT-49's migrate builds `Task`s and calls `Save`).
- **Seams (exported):** `task.CreateParams{..., Branch, Commit string}`, `task.Patch{..., Branch, Commit *string}`, `(*Store).Save(*Task)` with `Task.Branch`/`Task.Commit`.
- `TestStoreSave/writes a json record beside the body` (BIT-46.12, extended in BIT-46.13) asserts the exact unmarshalled record: add `"branch":""` and `"commit":""` to its expected value.
- MCP and CLI pass nothing new (fields stay empty in this track).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreCreate/writes git fields as given` (subtest in `TestStoreCreate` holding a table)
     - **Behavior:** the store records the git state its caller hands it, and writes empty strings when given nothing.
     - **Setup:** `NewProject(dir, "BIT")`; row `no git fields` `Create(CreateParams{Title: "T"})`; row `branch and commit` `Create(CreateParams{Title: "T", Branch: "v2", Commit: "6a1d345"})`.
     - **Assertions:** row `no git fields` raw JSON has `"branch": ""` and `"commit": ""`; row `branch and commit` has `"branch": "v2"` and `"commit": "6a1d345"`.
     - **Boundary:** empty vs set, so a hardcoded `""` can't pass the second row.
   - [ ] Confirm fails: `CreateParams` has no `Branch`.

2. **Implement (GREEN):**
   - [ ] Fields, `CreateParams`, record fields.

3. **More tests (RED → GREEN):**
   - [ ] `TestStoreUpdate/sets commit without revoking approval` (subtest): approved bar, `Update(id, Patch{Commit: &sha})` → `commit == sha`, `approved` still true, title/body unchanged. *Boundary:* git-only patch × approved, the BIT-47 path.
   - [ ] Add a row `a branch change keeps approval` to the `TestStoreUpdate/approval revocation` table: `Patch{Branch: &b}` keeps approval.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): tracks and bars carry caller-supplied branch and commit`