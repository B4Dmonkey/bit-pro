# BIT-45 research index

**Start at [decisions](decisions.md):** the operator decisions of 2026-09-30 (the split of BIT-46 into BIT-46/48/49) and 2026-10-01 (the split of BIT-47 into BIT-47/50, and the order 45 → 46 → 48 → 49 → 47 → 50). Each of BIT-46, 47, 48, 49 and 50 has its own `decisions` topic.

- [claim-audit-2026-10-02](claim-audit-2026-10-02.md): claim-by-claim audit of 9 topics + the BIT-45 body (115 confirmed, 13 wrong, 3 stale), a lint-green bar order verified in throwaway copies, and what the removal deletes that BIT-49/47/50 will want.

## Superseded by later decisions — read before the second-pass topics
These are operator decisions. The track bodies and v2-sketch.md win over the notes below.
- No project path or session dir in records. Paths live only in the db.
- Tracks and bars carry `branch` and a single `commit`, not `head`/main hash. Research, feedback and retro carry a `commits` list of `{sha, branch, at}`: the first entry is HEAD when the record is created, and a later write appends one when HEAD has moved.
- HEAD isn't captured on every write for tracks and bars. A bar is committed first, through a commit skill that always asks the operator's permission, and its hash is stored afterwards (BIT-47).
- The db is `main.db` (not `bit.db`), project dirs are uppercase, migrate timestamps are the migration time and commit fields get HEAD at migration (not git-log/mtime or empty anchors), and migrate stops and lists unknown files.
- Delivery order is BIT-45 → BIT-46 → BIT-48 → BIT-49 → BIT-47 → BIT-50, so anchors (BIT-47) and merge-aware completion (BIT-50) come last, not before feedback and migrate.
- Affected topics: [json-schema](json-schema.md), [history-anchors](history-anchors.md), [delivery-order](delivery-order.md) (each now carries a banner) and [migrate](migrate.md).

## Soundness pass 2 (revised bodies)
- [soundness-2](soundness-2.md): BIT-45's claims are confirmed. Touches misses `db/queries/projects.sql` (the count columns and `UpdateProjectCounts`) and every non-`cmd` test (`db/queue_test.go`, `claude/dispatch_test.go`, `task/counts_test.go`, `tui/*_test.go`).
- BIT-46 topic `soundness-2`: the global Claude wiring. The "Blocking: cutover never creates the user-scope wiring" finding is **resolved by BIT-49**: a first `bp migrate` ensures the global wiring. Also covers `marketplace add` being required, the scope-blind `mcp get`, the per-project notice, and the dev-test hazard.
- BIT-47 topic `soundness-2`: the no-fetch/origin gap, the timing of on-write capture, and v2-sketch's stale squash-only wording (since fixed). The completion parts now belong to BIT-50.

## Soundness pass (v2 @ 6a1d345)
A third pass re-checked BIT-45/46/47 against the code. Several second-pass notes below are now **stale against the operator's later decisions**: `migrate` and `json-schema` still say `bit.db` (now `main.db`), lowercase `<code>` dir (now uppercase), timestamps from `git log`/mtime (now migration time), and copy-or-refuse unknown files (now stop and list). See the superseded list above for the rest. Trust the track bodies and v2-sketch.md over those lines.
- [soundness](soundness.md): the BIT-45 delete list was re-verified and is sound. It missed the test caller `cmd/task_test.go:18`, the `add.go` Short text, and the gitignored `db/orm/`.
- BIT-46 issues (config.toml prefix blocker, `bp init`, `code UNIQUE`, Touches gaps) are in BIT-46 topic `soundness`. BIT-47 feasibility against `git log main` is in BIT-47 topic `soundness`.

---

This is the second pass. It answers 8 questions under the operator decisions:
- hard cutover on `v2`
- fresh `~/.local/share/bit/bit.db`
- daemon removed
- each record is a JSON metadata file plus a markdown content file, and bodies are not split
- feedback and retro live in top-level dirs with a `project` field
- `.bit/` is removed from bit-pro after migration
- no linking in the minimum scope

Topics from the first pass that these decisions superseded are marked below.

## Findings that change the scope's premises
- **bp never calls git.** Every anchor is new code. The squash commit **doesn't exist yet at `task_complete`**, because sign-off comes before the PR merge. Per-bar HEAD at status change is off by one, because the commit comes after `done`. → [history-anchors](history-anchors.md)
- **Claude worktrees resolve for free** under longest-prefix, because they live inside the repo at `.claude/worktrees/`. But git facts must come from the session dir, not the registered path. → [project-resolution](project-resolution.md)
- **opencode runs `bp serve mcp` with no `CLAUDE_PROJECT_DIR`**, so the MCP resolver needs a cwd fallback. → [project-resolution](project-resolution.md)
- **The store swap is contained.** Only `task/` knows the file format. Callers use `Task.Body`, MCP `body` is already a string, and no Go code parses checkboxes. 9 tests assert on frontmatter text and 18 write `.md` fixtures. → [json-schema](json-schema.md)
- **No timestamps exist today.** Migrate can only recover them from `git log --follow` (bit-pro) or mtime. → [json-schema](json-schema.md), [migrate](migrate.md)
- **The rehearsal fixture `tools/example` breaks.** Its reset/checkpoint scripts restore task state through git-tracked `.bit/`. → [testing](testing.md)
- **Only retro and learn do direct file I/O.** The do and bot-dev commit-staging lines have to be deleted. check writes to the repo root and is unaffected. → [skills](skills.md)
- **A launchd plist is still on disk** (`~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist`, not loaded). Delete it by hand at cutover. → [daemon-removal](daemon-removal.md)

## Topics (this pass)
- [decisions](decisions.md): the operator decisions of 2026-09-30 and 2026-10-01 (the splits and the order).
- [daemon-removal](daemon-removal.md): Q1. The exact delete list, partial edits (serve, root, tui), dead queries and migrations, and what survives (`bp add` as registration, `bp list` without counts). Product call: the TUI "Play?" prompt.
- [json-schema](json-schema.md): Q2. Every stored kind, its v1 metadata → v2 JSON fields, the pair-file design points (write order, ID glob, relocate moving pairs), the code and tests that assume frontmatter, and completed/ vs archive/. **Partly superseded (see its banner).**
- [history-anchors](history-anchors.md): Q3. The git facts table, when to capture each, how bit-pro squash-merges (the body keeps branch commit subjects), and worktrees and non-git projects. **Partly superseded (see its banner).**
- [project-resolution](project-resolution.md): Q4, **rewritten**. v1 resolvers re-verified, the longest-prefix design (normalisation, segment boundary, per-call MCP resolve, not-found errors), and what replaces bitdir.
- [migrate](migrate.md): Q5, **rewritten**. Target layout, conversion, staged build + verify + atomic commit, idempotency, rollback (the source is never touched), and the warning, which is just the resolver's error text.
- [skills](skills.md): Q6. Per-skill breakage and the new MCP tools (`feedback_list/read`, `retro_write/list/read`).
- [testing](testing.md): Q7. `just run` + `XDG_DATA_HOME`, `./bin/bp`, per-project `claude mcp add bit -e … -- <abs>/bin/bp`, `--mcp-config --strict-mcp-config --plugin-dir ./bit`, and the cutover checklist.
- [delivery-order](delivery-order.md): Q8. Seven slices (daemon out → registry + resolver → JSON pairs → anchors → top-level feedback/retro → migrate → sweep + cutover) and the separate-track list. **Superseded by the track splits (see its banner).**

## First-pass topics (still on disk)
- [bit-path-inventory](bit-path-inventory.md): the `.bit` reference inventory is still accurate for Go code. The daemon rows are moot, and the skills section is superseded by [skills](skills.md).
- [sqlite-registry](sqlite-registry.md): accurate for v1's db. Its "extend the existing db" framing is superseded by the fresh db decision. The embedded-dbmate + modernc driver + sqlc setup carries over.
- [config](config.md): still valid. `prefix` → `projects.code`, and config.toml is dropped.
- [coexistence](coexistence.md): largely superseded by the hard cutover. Its binary, plugin-from-main and MCP-name facts are reused in [testing](testing.md).
- [xdg](xdg.md): still valid. Its launchd env gotcha is moot once the daemon is gone.

## Open unknowns
- The casing of the `<code>` dir (the operator's example is lowercase `bit/`, while IDs are uppercase). **Settled since: uppercase.**
- When and how to resolve the squash commit: a lazy `gh` lookup, or a "merged" status. **Settled since: completion runs after the work lands and walks a ladder (BIT-50).**
- Whether unknown files under `.bit/` get copied raw or refused at migrate. **Settled since: stop and list.**
- Whether `--plugin-dir ./bit` conflicts with a project that has `bit@bit-pro` enabled. **Settled: no conflict (BIT-46 `plugin-dir-testing`).**
- Whether opencode sets the MCP server's cwd to the project.
- Whether `git rev-parse --path-format=absolute` exists in the installed git version, needed for plain `git worktree` support, which is optional.
- Whether learn should read every project's retro proposals directly now that they are central. **Settled since: yes (BIT-49).**
- The TUI "Play?" prompt after the daemon is removed. **Settled since: dropped.**