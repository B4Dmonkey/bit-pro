> **Superseded in part (banner added 2026-10-01).** The track bodies (BIT-46 Verse 2, BIT-49) win over this note:
> - **No project path or session dir in records.** Paths live only in the db (operator, 2026-09-30). The "project path" in *Common fields* below is out.
> - **`commit`, not `head` or a main hash.** Tracks and bars carry `branch` and a single `commit` (track = landing commit, bar = best-effort). Research topics, feedback notes and retro proposals carry a `commits` list of `{sha, branch, at}` instead: the first entry is HEAD when the record is created, and a later write appends one when HEAD has moved (operator, 2026-10-01).
> - **HEAD isn't captured on every write** for tracks and bars (BIT-47).
> - **Timestamps at migrate are the migration time,** not `git log --follow` or mtime. Commit fields get HEAD at migration (BIT-49).
> - **Order:** this format work is BIT-46 Verse 2. Feedback and retro going top-level is BIT-49 Verse 1.

# Record format: JSON metadata + markdown content (Q2)

**Operator decision (constraint):** each record is a JSON metadata file plus a markdown content file. Frontmatter becomes JSON fields. The body stays markdown, unchanged. The JSON has a `content` field pointing to the `.md`. Body content is **not** split into fields.

**Checked:** every stored kind and its current metadata, the body shapes in the data, and how much code and how many tests assume the one-file frontmatter format. Sampled bit-pro `.bit/`: tasks 11, completed 321, archive/tasks 8, feedback 27, research 12 topics. There is no `retro/` dir yet.

## Kinds today → v2 metadata
| Kind | Today | v1 metadata | v2 JSON fields (beyond common) |
|---|---|---|---|
| Track/bar (one struct) | `tasks/`, `completed/`, `archive/tasks/` `<ID>.md` | yaml frontmatter `task/task.go:19-28`: `id,title,status(todo/doing/done),approved?,phase?,phase_label?,order?`. Hand-rolled `---` split in `Parse` (:30), `Bytes` (:58). | the same keys. `parent` is derivable from the ID, so optionally store it |
| Feedback note | `feedback/<TRACK>-NNN.md` (`task/feedback.go:58`) | **none**. Track and seq exist only in the filename | `id`, `track`, `seq`. Moves to the **top-level** `feedback/` with a `project` field (operator) |
| Research topic | `research/<TRACK>/<topic>.md` (`task/research.go:37`) | **none**. Track and topic exist only in the path | `track`, `topic` |
| Retro proposals | `retro/<x>-proposals.md`, written by the skill | none | `name`, plus `project`/`track` as applicable. Top-level `retro/` (operator) |

- **Common fields (operator):** `project` (code), project path, `branch`, `head`, `created_at`, `updated_at`, optional main hash, and `content` (relative path to the `.md`). See [history-anchors](history-anchors.md).
- Real frontmatter usage across 330 completed/archived files: id/title/status 330, phase/phase_label 285, approved 107, order 2. The migration maps these 1:1.
- **No timestamps exist anywhere today.** At migrate, the only sources are file mtime and, for git-tracked `.bit/` like bit-pro, `git log --follow` per file.

## Design points the pair format raises
- **Naming:** `<ID>.json` + `<ID>.md` side by side is simplest. `content` then holds a sibling filename. Relative, never absolute, so the store dir can move.
- **Write order / atomicity:** write the `.md` (temp + rename), then the `.json` (temp + rename). The JSON is the record, so a crash leaves at worst an orphan `.md`. `relocateTree` (`store.go:112`) must move both files per task. `Relocate`/`Complete` rename trees of tasks and bars.
- **ID minting** (`store.go:637-650`) regexes filenames across tasks/completed/archive. Switch the glob to `*.json` so orphan `.md` files don't reserve IDs. Feedback seq minting (`feedback.go`) is the same.
- **`updated_at` on body-only edits:** `task_update` with only a body still has to rewrite the JSON to bump `updated_at`/`head`.
- **Research `index` links** like `[store](store.md)` keep working, because the topic content is still a `.md` file named by topic.
- **Top-level feedback IDs:** `<TRACK>-NNN` is already globally unique once project codes are UNIQUE, since a track ID carries its prefix. So `feedback/<TRACK>-NNN.{json,md}` needs no per-project subdir. `trackExists` (`feedback.go:36`) must then resolve the track in the owning project's store.

## Body shapes (for reference; they stay as markdown)
- Track: `## Why`/`## Summary`/`## Decisions`/`## Verses` with `- [ ] Verse N` and `Touches:`, `## Risks & unknowns`, `## References`. Older tracks use `## Phases`.
- Bar: `## **Verse N**`, `## Scope`, TDD cycle, `## Claude verifies`/`## User verifies`/`## Commit (user)`.
- The skills keep parsing these as text. do :63-89 and bot-dev :20-22 branch on a "User verifies" section, and do :94 flips verse checkboxes. The format change affects none of that.

## What assumes the one-file markdown format in code
- **`task/store.go`**: `.md` hardcoded at :38, :46, :54, :62 and in the globs at :443, :467, :614, :626. Also `feedback.go:20,24` and `research.go:21,95-96`. Replace `Parse`/`Bytes` with JSON (un)marshal plus a separate body read/write.
  - The rest of the code (cmd, MCP, TUI) talks to `task.Task{…, Body}`. If the store keeps returning `Body`, **callers don't change**.
- **MCP** (`cmd/serve_mcp.go:116-220`): the JSON structs already exist, and `body` is a string. Only the descriptions that name `.bit/completed/` (:76), `.bit/archive/tasks/` (:84) and `.bit/research/<track>/` need rewording. `taskReadOutput` omits `order`.
- **TUI**: glamour renders `t.Body`. Unchanged.
- **CLI**: `bp task read` prints a header plus the body. Unchanged.
- **Tests**: 45 test files.
  - 18 write `.md` paths or fixtures (heaviest: `cmd/feedback_add_test.go`, `cmd/serve_mcp_write_test.go`, `cmd/task/delete_test.go`, `complete_test.go`).
  - 9 assert on frontmatter text: `task/task_test.go`, `task/store_test.go`, `task/counts_test.go` (deleted with the daemon), `cmd/task/{create,update,list,complete}_test.go`, `cmd/feedback_add_test.go` and `cmd/serve_mcp_write_test.go`.
  - These are the rewrites. Tests that go through `Store` methods survive.

## completed/ and archive/
- Both go through `relocateTree`.
  - `Complete` (:108 → `completed/`) refuses on unfinished bars.
  - `Relocate` (:104 → `archive/tasks/`) is the delete path, and `force` is allowed.
- `completed/` is read only for `Counts`, which is deleted with the daemon. `archive/` only reserves IDs and satisfies `trackExists`. `List()` reads `tasks/` only.
- **Minimal:** keep the three dirs and move pairs. **Alternative:** one dir plus a `state` field. Not needed for this track.