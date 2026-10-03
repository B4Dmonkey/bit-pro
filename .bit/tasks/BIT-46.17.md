---
id: BIT-46.17
title: A feedback note is stored as a <TRACK>-NNN.json record beside its .md
status: todo
phase: 2
phase_label: records are JSON metadata plus markdown
---
## **Verse 2**

Feedback notes get the pair format and the same metadata, plus their creating commit from the caller. Notes stay in the project's own `feedback/` until BIT-49 moves them to the shared folder. A test reading the `.json` forces it. This bar completes the verse.

## Scope
- `task/feedback.go`:
  - unexported `noteRecord`: `project, id ("<TRACK>-NNN"), track, seq, created_at, updated_at, commits, content`.
  - `AddNote(track, body string, head Commit) (string, error)`: write `<TRACK>-NNN.md`, then `.json`, with `commits = appendCommit(nil, head)`. Return the `.md` path, as today.
  - `nextNoteSeq` globs and matches `*.json` (`:24`), so a stray `.md` reserves no number.
  - The overwrite race in a shared folder is BIT-49's (it switches to `O_EXCL`); not this bar.
- callers pass `task.Commit{}`: `cmd/feedback_add.go:19`, `cmd/serve_mcp.go:457`.
- **Seam (exported):** `(*Store).AddNote(track, body string, head task.Commit) (string, error)`.
- Tests: `cmd/feedback_add_test.go` (raw-byte reads of the `.md` stay valid, since the body is unchanged; the directory listing at `:103` must ignore `.json` or assert both), `cmd/serve_mcp_write_test.go:474-555` (`*.md` globs at `:523,546`). Both files were converted in BIT-46.8 and BIT-46.9.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreAddNote/writes a json record beside the body` (subtest of a new `TestStoreAddNote`, `task/feedback_test.go`, new file)
     - **Behavior:** a feedback note is metadata plus the verbatim note body, numbered per track.
     - **Setup:** `NewProject(dir, "BIT")`, fixed clock, saved track `BIT-46`; `AddNote("BIT-46", "## What the plan said\n\n...", Commit{SHA: "aaa111", Branch: "v2", At: t0})`.
     - **Assertions:** returned path `feedback/BIT-46-001.md` with exactly that body. `BIT-46-001.json` has `"project":"BIT","id":"BIT-46-001","track":"BIT-46","seq":1`, `commits` with one entry, `"content":"BIT-46-001.md"`, and both timestamps.
     - **Boundary:** first note, seq 1.
   - [ ] Confirm fails: `AddNote` signature (compile), then no `.json`.

2. **Implement (GREEN):**
   - [ ] `noteRecord`, the new `AddNote`, the `nextNoteSeq` glob.

3. **More tests (RED → GREEN):** (subtests of `TestStoreAddNote`)
   - [ ] `TestStoreAddNote/second note counts records only`: second note → seq 2, `BIT-46-002`; a stray `feedback/BIT-46-003.md` with no `.json` doesn't push it to 4. *Boundary:* counting from records only.
   - [ ] `TestStoreAddNote/no commit writes an empty list`: `Commit{}` → `"commits": []`.
   - [ ] `TestStoreAddNote/accepts an archived track`: a note on an archived track (track in `archive/tasks/`) still works (`trackExists` with `.json` paths).

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
Whole verse, in a fresh sandbox (rebuild as in BIT-46.11's steps; throw away any Verse 1 sandbox):
- [ ] `printf 'demo\n' | bp add .`, then `bp task create "T"`, `bp task create "B" -p DEMO-1`, `bp approve DEMO-1.1`, `bp feedback add DEMO-1 -d "note"`. Then `ls $SB/data/bit/DEMO/tasks` shows `DEMO-1.json DEMO-1.md DEMO-1.1.json DEMO-1.1.md`, and `cat $SB/data/bit/DEMO/tasks/DEMO-1.1.json` shows `"approved": true`, `"project": "DEMO"`, `"branch": ""`, `"commit": ""`, `"content": "DEMO-1.1.md"`.
- [ ] `ls $SB/data/bit/DEMO/feedback` shows `DEMO-1-001.json DEMO-1-001.md`.
- [ ] `bp task list`, `bp task read DEMO-1.1` and `bp tui` show the same things they did before this verse.
- [ ] Whole slice: every record is a readable `.json` beside its untouched `.md`, and the CLI, MCP and TUI behave as before.

## Commit (user)
`feat(bit): store feedback notes as JSON metadata plus markdown`