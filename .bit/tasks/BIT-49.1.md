---
id: BIT-49.1
title: Feedback notes land in the shared top-level feedback/ folder
status: todo
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

BIT-46.17 left notes in the project's own `<data>/<CODE>/feedback/`. A CLI test that expects the note at `<data>/feedback/` can't pass until the task store knows the data root. `trackExists`/`resolveTrack` stay on the project root, because research shares them (`task/research.go`).

## Scope
- `task/store.go`: `Store` gains an unexported `data string`. New `func (s *Store) WithDataRoot(dir string) *Store` sets it and returns `s`. **Seam:** the store's way to reach the shared `feedback/` and `retro/` folders. BIT-46 left adding it to this track.
- `task/feedback.go`: `feedbackDir()` becomes `filepath.Join(s.data, feedbackSubdir)`. `AddNote` returns an error when `s.data == ""` (`store has no data root`), so a misbuilt store can't write `feedback/` into the working directory.
- `project/store.go` `StoreFor` (BIT-46.19 extracted it, and `OpenStore` is `Find` + `StoreFor`): `data, err := store.Dir()`, then `task.NewProject(root, p.Code).WithDataRoot(data)`. Every CLI, MCP and TUI store, and `bp remove`'s, gets it.
- Test files touched: `cmd/feedback_add_test.go` (converted by BIT-46.8, `TestFeedbackAddCmd`), `cmd/serve_mcp_write_test.go` (converted by BIT-46.9, `TestFeedbackAddHandler`), `task/feedback_test.go` (created in shape by BIT-46.17, `TestStoreAddNote`). All three are already converted, so there's no conversion step.
  - `task/feedback_test.go`'s existing `TestStoreAddNote` cases build `NewProject(filepath.Join(d, "BIT"), "BIT").WithDataRoot(d)`, since `AddNote` now refuses a store with no data root.
  - Feedback path assertions change from `<data>/bit/<CODE>/feedback/...` to `<data>/bit/feedback/...`, including the base that `serve_mcp_write_test.go` joins its `testFeedbackDir` (`"feedback"`) onto: it becomes the data root.

## TDD cycle

0. **Restructure:** none. `cmd/feedback_add_test.go` was converted by BIT-46.8 (`TestFeedbackAddCmd`), `cmd/serve_mcp_write_test.go` by BIT-46.9 (`TestFeedbackAddHandler`), and `task/feedback_test.go` was created in shape by BIT-46.17 (`TestStoreAddNote`).

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd/writes to the shared feedback folder` (`cmd/feedback_add_test.go`)
     - **Behavior:** a note recorded in one project lands in the folder that every project shares, still keyed to its own track.
     - **Setup:** `dir := initProject(t, "BIT")` (sandboxes `HOME`); `createTask(t, "Track", "...")`; `bp feedback add BIT-1 -d "## What the plan said\n\n..."`.
     - **Assertions:** with `d, _ := store.Dir()` (`initProject` sets `XDG_DATA_HOME=""`, so `d` is `$HOME/.local/share/bit`): printed path == `d/feedback/BIT-1-001.md`; that file's bytes equal the body; `d/feedback/BIT-1-001.json` has `"project": "BIT"`; `d/BIT/feedback` does not exist.
     - **Boundary:** where the note lands: shared folder, not the project dir.
   - [ ] Confirm fails: the note is written under `bit/BIT/feedback/`.

2. **Implement (GREEN):**
   - [ ] `data` field, `WithDataRoot`, `feedbackDir`, `StoreFor`. Update the moved path assertions.

3. **More tests (RED → GREEN):**
   - [ ] `TestStoreAddNote/two projects share the folder` (`task/feedback_test.go`): stores `BIT` and `EX` on one data root, each with track 1. Notes on `BIT-1` and `EX-1` → `feedback/BIT-1-001.{md,json}` and `feedback/EX-1-001.{md,json}`, each `.json` with its own `project`. *Boundary:* two codes in one folder, each with seq 1.
   - [ ] `TestStoreAddNote/refuses without a data root`: `NewProject(dir, "BIT")` with no `WithDataRoot` → error, and nothing is created under the cwd (`t.Chdir(t.TempDir())`, then `os.ReadDir(".")` is empty). *Boundary:* the data root is unset.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): feedback notes land in the shared top-level feedback folder`