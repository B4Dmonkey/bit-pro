---
id: BIT-45.5
title: Daemon scripts and notes are removed, and the whole v2 build runs without a daemon
status: done
phase: 1
phase_label: No daemon in a v2 build
---
## **Verse 1**

Removes the non-Go leftovers that describe or operate the daemon and queue, which no longer exist after BIT-45.1-.4. This is a docs and scripts deletion, so there's no test. It closes the verse, so it carries the whole-slice check.

## Scope
- `clear-queue.sh`: delete. It runs `DELETE FROM queue` against the v1 db, and nothing enqueues any more.
- `abort-run.md`: delete. It's a playbook for aborting a daemon-dispatched run (`abort-run.md:1-4`).
- `automation-notes.md`: delete the whole file. Every section (Checklist, Next steps, Decisions that still bind, launchd mechanics, Measured facts, Open gaps, Escape hatch, Position, Docs) is about the daemon/dispatch phase. Git history keeps it.
- `mcp-notes.md`: remove or reword the daemon content, keeping the MCP content intact:
  - the daemon mentions at :78, :125, :219, :224-231 (the "`bp serve` is a parent with two children" and pending-rename bullets: `serve` now has one child, `mcp`), :252-253 and :265-273 (the "MCP does not dispatch; the daemon does" decision);
  - the Command inventory row at :284: drop `start`, `stop`, `status`, `serve daemon` from the operator-only list;
  - the section "Relationship to the automation phase" (:327-345): delete it;
  - :373 and :442-447;
  - every pointer to `automation-notes.md` (:12, :30, :79, :129, :229, :253, :342, :418): drop it, since that file is gone.
- Untouched: `README.md`, `hierarchy.md`, `bit/`, `scripts/`, `Justfile`, `update/` (no daemon references, per research topic `claim-audit-2026-10-02`). Also untouched: the untracked `v2-sketch.md`, `known-issues.md` and `rfc-plan-orchestrator.md`. The launchd plist is removed by hand at cutover, not here.

## TDD cycle

1. **No new test.** Docs and a shell script only.
   - [ ] Delete the three files and edit `mcp-notes.md` per Scope.

## Claude verifies
- [ ] `git ls-files | xargs grep -n -i -l 'daemon\|clear-queue\|abort-run\|automation-notes' -- 2>/dev/null | grep -v '^\.bit/'` lists no file. Records under `.bit/` (completed tracks, BIT-42) keep their history and are excluded.
- [ ] `just test` passes. It includes the MCP harness tests (`cmd/mcp_harness_test.go`, `cmd/serve_mcp*_test.go`), which cover `bp serve mcp`.
- [ ] `just lint` reports `0 issues`.

## User verifies
- [ ] Whole slice, with a fresh temp build and never `just install`. First record the live v1 data's timestamps: `stat -f '%m %N' ~/.local/share/bit-pro/*`. Then `just db-gen-queries && go build -o /tmp/bp-v2 . && SBX=$(mktemp -d) && mkdir $SBX/proj && cp -R .bit $SBX/proj/`. Then, with `HOME=$SBX XDG_DATA_HOME=$SBX/share` on each call:
  - `/tmp/bp-v2 --help` lists no `start`/`stop`/`status`, and `/tmp/bp-v2 serve --help` lists only `mcp`;
  - `printf '\n' | … /tmp/bp-v2 add $SBX/proj` and then `… /tmp/bp-v2 list` print `BIT<tab><path>` with no counts;
  - `cd $SBX/proj && … /tmp/bp-v2 tui` browses, reloads and approves with no Play prompt and no cyan rows.
  - `stat -f '%m %N' ~/.local/share/bit-pro/*` prints the same timestamps as before, so the live v1 db wasn't touched.

## On track completion
- When the operator signs BIT-45 off, archive BIT-42 with `task_delete` (scope Decision: BIT-42 is superseded by this track). This is a record step at completion, not part of this bar's commit, so it doesn't run before sign-off.

## Commit (user)
`docs: remove daemon scripts and notes`