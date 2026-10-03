---
id: BIT-45.4
title: bp list prints code and path only, and the queue and count queries are gone
status: todo
phase: 1
phase_label: No daemon in a v2 build
---
## **Verse 1**

`bp list` stops printing the count buckets, which only the daemon wrote. This is a removal, so it adds no new test: `TestListCmd`'s expected output (today `TestListCmd_PrintsProjectsByCode`) loses the count columns because the output really changes, and `TestListCmd_ShowsProjectCounts` goes with the counts. With BIT-45.1-.3 done, the queue queries, `UpdateProjectCounts` and `GetProjectByPath` have no callers left, so they go here too. The schema and migrations stay as they are (no drop migrations, per the scope).

## Scope
- `cmd/list_test.go`: change the `"three projects"` want (:30-32) to `"ACE /tmp/ace MID /tmp/mid ZED /tmp/zed"`, and delete `TestListCmd_ShowsProjectCounts` (:67-115). `seedProject` (:117) stays, because the surviving list test still uses it.
- Test conversion (`.claude/rules/go-tests.md`). This bar touches `cmd/list_test.go`, so it converts what survives in it, as a pure restructure: same cases, same assertions, still green. `TestListCmd_PrintsProjectsByCode` is a table test and, once `ShowsProjectCounts` is deleted, the unit's only test, so it becomes `TestListCmd` with its rows unchanged: `TestListCmd/three projects` and `TestListCmd/no database yet`. `db/queue_test.go` is deleted outright, so it needs nothing.
- `cmd/list.go:33-34`: `fmt.Fprintf(out, "%s\t%s\n", p.Code, p.Path)`.
- `db/queries/projects.sql`: `ListProjects` → `SELECT id, path, code FROM projects ORDER BY code;`. Delete `UpdateProjectCounts` (:10-11) and `GetProjectByPath` (:13-14). `CreateProject` and `ProjectExists` stay.
- `db/queries/queue.sql`, `db/queue_test.go`: delete. `openTestDB` lives in `queue_test.go`, but `db/queries_test.go` doesn't use it.
- `db/orm/queue.sql.go`: delete it **by hand**. `db/orm/` is gitignored and `sqlc generate` (v1.31.1) doesn't prune output for a deleted query file. A stale copy still compiles and can hide a missed caller. Then run `just db-gen-queries`.
- `cmd/add.go:24`: `Short: "Enroll a project in the registry"`. Only the daemon clause is dropped.
- No change: `db/migrations/*` (the `queue` table and count columns stay until BIT-46's fresh db), `db/open_test.go` (still 4 migrations), `db/queries_test.go`.

## Steps
- [ ] Rename `TestListCmd_PrintsProjectsByCode` to `TestListCmd`, rows unchanged. `just test` stays green.
- [ ] Trim `ListProjects` in `projects.sql`, run `just db-gen-queries`, change the `Fprintf` in `cmd/list.go`, and update the `"three projects"` want in `cmd/list_test.go`.
- [ ] Delete `TestListCmd_ShowsProjectCounts`, `UpdateProjectCounts`, `GetProjectByPath`, `queue.sql`, `queue_test.go` and the stale `db/orm/queue.sql.go`, then regenerate. Reword `add.go`'s `Short`.

## Claude verifies
- [ ] `rm -f db/orm/queue.sql.go && just test` passes.
- [ ] `just lint` reports `0 issues`.
- [ ] `grep -rn 'UpdateProjectCounts\|GetProjectByPath\|EnqueueTask\|ListQueueByProject\|Backlog' --include='*.go' . ` matches nothing outside `db/orm/models.go`. `models.go` keeps the `Queue` struct and the count fields while the migrations keep the columns.
- [ ] `grep -rn -i daemon cmd/` matches nothing.

## User verifies
- [ ] Rebuild the temp binary (`just db-gen-queries && go build -o /tmp/bp-v2 .`). Then, in a sandbox with a copy of the records (so `bp add` skips the real `claude` wiring, which only runs for a folder without `.bit/`): `SBX=$(mktemp -d); mkdir $SBX/proj; cp -R .bit $SBX/proj/; printf '\n' | HOME=$SBX XDG_DATA_HOME=$SBX/share /tmp/bp-v2 add $SBX/proj`. It prints `added BIT <path>`. Then `HOME=$SBX XDG_DATA_HOME=$SBX/share /tmp/bp-v2 list` prints `BIT<tab><path>` with no `backlog:`/`todo:` columns, and `HOME=$SBX XDG_DATA_HOME=$SBX/share /tmp/bp-v2 add --help` doesn't mention the daemon. Keep the sandbox on every call: even `--help` runs the plugin refresh against `$HOME`.

## Commit (user)
`feat(list)!: drop project counts and the queue queries`