---
id: BIT-46.12
title: A task is stored as an <ID>.json record beside an unchanged <ID>.md body
status: todo
phase: 2
phase_label: records are JSON metadata plus markdown
---
## **Verse 2**

Tasks move from frontmatter `.md` files to a JSON metadata file plus the body as plain markdown. A test that reads the `.json` on disk can't pass against frontmatter, so it forces the format. `Load`/`Save`/`List`, ID minting and every move change together, because they all read the same files. Research and feedback follow in later bars. `task.Parse`/`Task.Bytes` (v1 frontmatter) stay unchanged: BIT-49's migrate reads v1 `.bit/` with them and checks a byte-for-byte round trip.

## Scope
- new `task/record.go`: unexported `taskRecord` with JSON tags `id, title, status, approved, phase, phase_label, order, content`, no `omitempty`. `order` is written as `[]` when empty, never `null`. `parent` isn't stored. `content` is the `.md` file's base name (`<ID>.md`). Marshal with `json.MarshalIndent(rec, "", "  ")` plus a trailing newline.
- `task/store.go`:
  - `Path(id)` returns the `.json` path (`pathologize.Join(tasksDir, ID+".json")`). New unexported `bodyPath(dir, id)` gives the `.md`. Same change for `archivePath`/`completedPath`.
  - `Save`: write the `.md` (body bytes), then the `.json`.
  - `Load`: read the `.json`; normalize `id` and each `order` entry with `NormalizeID`, as `Parse` does today (the lowercase-ID tests below depend on it); read the body from `pathologize.Join(dir, content)`. No separate check on `content`: only bp writes it, and `Join` keeps it inside the dir.
  - `List`: glob `*.json` (`:467`). `NextChildID`/`NextID`: glob and regex on `.json` (`:614-626`), so a stray `.md` reserves nothing. `NextChildID` stats `Path(parent)` (`.json`).
  - `relocateInto`: move the `.md`, then the `.json`, with `os.Rename` (both stay under one project dir on one filesystem). `fileflow.Move` would silently suffix a clashing ID, so it's not used here.
- `task/feedback.go` `trackExists` (`:35-43`): it already uses `Path`/`completedPath`/`archivePath`, which now return `.json`.
- Tests (all converted in earlier bars: BIT-45.1, BIT-46.6, .8, .9):
  - `task/store_test.go`: seeds that write `.md` files into `archive/tasks/` or `completed/` (the reserves cases of `TestStoreNextID` and `TestStoreNextChildID`) write `.json` records instead. `TestStorePath/contains untrusted id` and the archive-path rows in `TestStoreRelocate/contains untrusted id` expect `.json`.
  - Hand-written frontmatter seeds become hand-written `.json` records plus a `.md` body, keeping the lowercase IDs they test: `writeRawTask` (`cmd/task/complete_test.go`, `cmd/task/update_test.go`) and the raw track in `cmd/task/list_test.go` (`:101`).
  - Task-file stats and reads that name `<ID>.md` check `<ID>.json`: `cmd/task/{complete,create,delete,update}_test.go`, `cmd/task_test.go`, `TestRunMCPServer/resolves a subfolder root through the registry` in `cmd/serve_mcp_test.go` (BIT-46.9), `cmd/feedback_add_test.go:181-185`, `cmd/serve_mcp_write_test.go:586-710`. Reads that check frontmatter text (`id: BIT-1.2`, `title: First bar`) read the `.json` field instead. `ReadDir` counts of `tasks/`/`completed/` double (pairs), or filter `*.json`.
  - `TestTaskReadCmd/omits phase when absent` drops its "no `phase` key in the file" assertion: every field is now written (`"phase": 0`). Its output assertion stays.
  - The `contains path traversal id` fixtures (`cmd/task/read_test.go`, `delete_test.go`, placed in BIT-46.8) become `README.json`, since `Path` now targets `.json`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreSave/writes a json record beside the body` (subtest of a new `TestStoreSave`, `task/store_test.go`)
     - **Behavior:** a saved task is a JSON metadata file with an unchanged markdown body beside it.
     - **Setup:** `s := NewProject(t.TempDir(), "BIT")`; `s.Save(&Task{ID: "BIT-7", Title: "Ship it", Status: "todo", Body: "## Why\n\n- [ ] a checkbox\n"})`.
     - **Assertions:** `tasks/BIT-7.md` bytes == the body exactly. `tasks/BIT-7.json` unmarshals to `{"id":"BIT-7","title":"Ship it","status":"todo","approved":false,"phase":0,"phase_label":"","order":[],"content":"BIT-7.md"}`. The raw JSON contains `"order": []` (not `null`) and ends with `"}\n"`.
     - **Boundary:** zero values (`approved` false, `phase` 0, empty `order`) are still written, the "every field is written" rule.
   - [ ] Confirm fails: `BIT-7.json` doesn't exist (frontmatter `.md` is written).

2. **Implement (GREEN):**
   - [ ] `taskRecord`, the new `Save`/`Load`/`Path`.

3. **More tests (RED → GREEN):**
   - [ ] `TestStoreLoad/round trips a saved task` (existing): a task with `approved`, `phase`, `phase_label`, a 2-item `order` and a multi-line body round-trips. *Boundary:* every field set.
   - [ ] `TestStoreList/ignores a stray body` (subtest): a `tasks/BIT-9.md` with no `.json`, plus one saved task → `List` returns 1, and `NextID("BIT")` ignores 9. *Boundary:* an orphan `.md` reserves nothing (written `.md` first, so a crash leaves exactly this).
   - [ ] Extend `TestStoreRelocate/moves file out of list`: after `Relocate`, both `archive/tasks/BIT-1.{json,md}` exist and neither remains in `tasks/`. Add `TestStoreComplete/moves both files` (subtest of a new `TestStoreComplete`): the same for `Complete` → `completed/`. *Boundary:* the pair moves as one; a lone `.md` left behind would break `Load` after revive.
   - [ ] the reserves cases (`TestStoreNextID/reserves archived ids`, `TestStoreNextID/reserves completed ids`, `TestStoreNextChildID/reserves archived children`, `TestStoreNextChildID/reserves completed children`) pass with `.json` seeds.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `go test ./task -run 'Parse|Bytes'` passes unchanged (v1 frontmatter kept for migrate)
- [ ] delete any Verse 1 sandbox stores (their markdown records are no longer readable)

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): store tasks as JSON metadata plus a markdown body`