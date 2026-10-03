# Claim-by-claim audit, 2026-10-02 (v2 @ 6a1d345, == main)

**Checked:** every factual claim in topics bit-path-inventory, coexistence, config, daemon-removal, delivery-order, project-resolution, sqlite-registry, testing, xdg, and in the BIT-45 track body. I read the code, checked the machine, and ran a trial removal. The trial applied the whole removal in throwaway copies of HEAD under the scratchpad, as five bars. Each bar got a fresh copy with no `db/orm`, then `sqlc generate`, `go build ./...`, `go vet ./...`, `go test ./...` and `golangci-lint run ./...`. Baseline on HEAD: all green, `0 issues`.

**Totals:** 131 claims: 115 CONFIRMED, 13 WRONG, 3 STALE.

## Headline findings
- **"Green" must include lint.** `.pre-commit-config.yaml` runs `just fmt`, `just lint` and `just test`. `.claude/hooks/post-edit-go.sh` runs golangci-lint after every Go edit, and `stop-go.sh` blocks the stop on lint or test failure. The `standard` linters include `unused`, and the removal leaves three unexported things unused, which fails lint while build and test pass:
  - `task/store.go:442` `(*Store).listCompleted`. Its only caller is `task/counts.go:38`.
  - `cmd/cmd_test.go:58` `runWithContext`. Only the serve-daemon tests use it.
  - In tui: `idSet`, `barChildrenOf`, `trackTitle`, `allApproved`, `enqueueableBarIDs`, `targetBar`, `playPromptView` (`board.go:271`) and the `pendingApprovalID` plumbing (`model.go:99,282-292,414`, `board.go:193`).
- **`sqlc generate` (v1.31.1) does not delete `db/orm/queue.sql.go`** when `db/queries/queue.sql` is removed (tested). The stale file still compiles, because `models.go` keeps `Queue` while the queue migration stays. It is harmless, but it only goes away if someone deletes it by hand. A fresh checkout or worktree has no `db/orm/` at all, so `go build` fails until `sqlc generate` runs. Bars should run `just test` and `just lint`, or `sqlc generate` first, rather than bare `go test`, or a `.sql` edit isn't reflected.
- **Lost by this removal, and needed by later tracks:**
  - `claude.DirRunner`/`ExecDirRunner` (`claude/dispatch.go:14-33`): a stubbable func type that runs a command in a dir and returns stdout and the exit code. This is exactly the seam BIT-49's "small git helper that shells out to git… tests stub it" needs, and BIT-47/50 extend it. The surviving `claude.Runner` (`claude/sync.go:10`) returns only an error, so it can't capture output.
  - `daemon.ExecRunner` (`daemon/daemon.go:19-34`) has the same shape without a dir.
  - `(*Store).listCompleted` (`task/store.go:442-465`) is the only completed-dir lister, which BIT-49's migrate (it carries `completed/`) will want.
  - `task.ParentID` (`task/store.go:548`) loses its only caller.
  - Recommend deleting them anyway (lint) and pointing BIT-49 at `git show 6a1d345:claude/dispatch.go` and `6a1d345:task/store.go` to copy from.

## Verified bar order (each step green on build, vet, test and lint in a fresh copy)
1. **cmd surface**
   - Delete `cmd/{start,stop,status}.go` and their tests **together**. `start_test` uses `bootoutCall` (stop_test), and `launchctlDict`, `disabledStore`, `printDisabled` and `listSubcmd` (status_test).
   - `cmd/serve.go` keeps only `serveCmdUse` and `newServeCmd` with `serve mcp`. Drop `serveDaemonCmdUse`, `claudeBinEnv`, `claudeBinFallback`, `serveTick`, `serveRunner`, `claudePath`, `newHandler` and `newServeDaemonCmd`.
   - `cmd/serve_test.go` keeps only `TestServeMCPCmd_IsListedInServeHelp` (:158). The other 11 tests go.
   - `cmd/root.go`: drop the daemon import (:15), `daemon.ExecRunner` (:125), the `lc` param (:128) and the wiring at **:147-149**.
   - `cmd/cmd_test.go`: drop the daemon import (:11), `nothingLoaded` (:24), `runWithDaemon` (:28) and `runWithContext` (:58). Change `newRootCmd(…, nothingLoaded)` → one arg at `cmd_test.go:46,128`, `task_test.go:18` and `root_test.go:223,243`, and drop the `root_test.go:216` row.
2. **packages**
   - Delete `daemon/` (8 files) and `claude/dispatch.go` and `claude/dispatch_test.go`.
   - Delete **`claude/testdata/agents.json`**. Its only reader is `dispatch_test.go:11`.
   - Delete `task/counts.go` and `task/counts_test.go`, plus `listCompleted` (`task/store.go:442-465`).
   - Steps 1 and 2 can merge.
3. **TUI**
   - `cmd/tui.go`: drop :31-40, :45-46 and `queueFuncs` (:51-94), and the imports `context`, `os`, `db` and `orm`.
   - `tui/model.go`:
     - `reloadedMsg.queued` (:36) and the fields :89-91 and :96-99.
     - `WithEnqueue`/`WithListQueue` (:174-182) and the listQueue part of `reloadCmd` (:197-209).
     - `applyQueued`, `handlePlayPrompt`, `enqueueSelected` and the "e" key (:438).
     - The view branches (:636, :648), and the dead helpers listed above (:721-753).
   - `tui/board.go`: the `queued` param (:78,84), the "e" key (:226), `pendingApprovalID` (:192-194) and `playPromptView`.
   - `tui/delegate.go`: `queuedColor` (:19), the `queued` field (:24) and :33-35.
   - Tests:
     - `model_test.go`: **:38-169** (8 tests) and **:1456-1699** (9 tests), plus the now-unused `slices` import.
     - `board_test.go`: :877-905, and the `nil` 2nd arg at **:237,246-248,259**.
     - `delegate_test.go`: :278-302.
4. **db + list**
   - Delete `db/queries/queue.sql` and `db/queue_test.go`. `openTestDB` lives there, and `queries_test.go` doesn't use it.
   - `projects.sql`: `ListProjects` → `SELECT id, path, code`, and delete `UpdateProjectCounts` (:10-11) and `GetProjectByPath` (:13-14). After step 3 it has no callers, and the trial deleted it with everything green. The track doesn't decide this.
   - `cmd/list.go:33-34` → `"%s\t%s\n"`.
   - `cmd/list_test.go`: the expected strings at **:30-32**, and delete `TestListCmd_ShowsProjectCounts` (:67-115).
   - `cmd/add.go:24` Short text.
   - `db/queries_test.go` and `db/open_test.go` (4 migrations) need **no** change.
   - Locally, delete the stale `db/orm/queue.sql.go` by hand.
   - The order of steps 3 and 4 is forced: `cmd/tui.go:60-80` uses `GetProjectByPath`, `EnqueueTask` and `ListQueueByProject`, and `daemon/loop.go:73` and `status_test.go:184-194` use `UpdateProjectCounts`.
5. **non-Go:** `clear-queue.sh`, `automation-notes.md` and the daemon parts of `mcp-notes.md` (see D6). README, hierarchy.md, `bit/`, `scripts/`, Justfile and `update/` have no daemon references.

## Track body (BIT-45)
| # | Claim | Verdict | Evidence / correction |
|---|---|---|---|
| T1 | launchd service isn't loaded | CONFIRMED | `launchctl list \| grep bit-pro` → no match |
| T2 | log last ran 2026-08-29 | CONFIRMED | `~/.local/share/bit-pro/daemon.log` last line 2026-08-29T09:33:33 "stopped" |
| T3 | the daemon is "the only thing that uses v1's registry db" | **WRONG** | `bp add` (`cmd/add.go:32-73`), `bp list` (`cmd/list.go:19-25`) and the TUI (`cmd/tui.go:36`) use it too. The daemon is the only user of the **queue** and the **count columns** |
| T4 | queued display `WithListQueue` at `cmd/tui.go:33,46` | CONFIRMED | :33 decl, :46 call |
| T5 | `bp status` = launchd status + counts | CONFIRMED | `cmd/status.go:22-35,40-63` |
| T6 | `claude.Runner` used by init/add for plugin sync + MCP | CONFIRMED | `cmd/init.go:51-68`, `cmd/add.go:67` |
| T7 | `cmd/add.go:24` says "the daemon watches" | CONFIRMED | |
| T8 | counts in `ListProjects`, `UpdateProjectCounts` in `projects.sql` | CONFIRMED | `db/queries/projects.sql:8,10-11` |
| T9 | a v2 build opens v1's live db (`store/store.go:22`, `db/open.go:25`) | CONFIRMED | both lines; `~/.local/share/bit` doesn't exist |
| T10 | `bp add` without `.bit/` runs the real claude wiring (`cmd/add.go:66`) | CONFIRMED | add.go:66-69 → `writeClaudeWiring` init.go:51 (settings.json, `claude plugin`, `claude mcp add`) |
| T11 | daemon/ imported by `cmd/{root,serve,start,stop,status}.go` + `cmd/{cmd,start,stop,status}_test.go` | CONFIRMED | grep: exactly those 9 files |
| T12 | `claude/dispatch.go` used only by daemon/ and `cmd/serve.go` | CONFIRMED | `cmd/serve.go:27`; `daemon/loop.go` |
| T13 | plist present, not loaded | CONFIRMED | `~/Library/LaunchAgents/com.github.b4dmonkey.bit-pro.plist` (Aug 28) |
| T14 | plist deletion is on the v2-sketch cutover checklist | CONFIRMED | `v2-sketch.md:115` |
| T15 | Tests: `cmd/task_test.go:18` (the newRootCmd signature change) | **WRONG** (incomplete) | Also `cmd_test.go:31,46,61,128` and `root_test.go:223,243` |
| T16 | Tests: `cmd/list_test.go:67-100` (the counts test) | **WRONG** (incomplete) | `TestListCmd_PrintsProjectsByCode` expects counts too (:30-32), and the counts test runs :67-115 |
| T17 | Tests: `db/{queue,queries}_test.go` | **WRONG** | `db/queries_test.go` needs no change (only CreateProject/ListProjects Code+Path; green in the trial) |
| T18 | Touches code list | **WRONG** (incomplete) | missing `task/store.go:442` listCompleted (lint), `claude/testdata/agents.json`, `cmd/cmd_test.go:58` runWithContext, the tui dead helpers, `board.go:226` "e" key and `board.go:271` playPromptView |
| T19 | stale `queue.sql.go` can linger; "regenerate it" | **WRONG** remedy | `sqlc generate` leaves it in place (tested). Delete it by hand. It is harmless if left (compiles) |
| T20 | callers go before the packages they import keeps every bar green | CONFIRMED | trial: bars 1-4 above all green incl. lint |

## daemon-removal
| # | Claim | Verdict | Evidence / correction |
|---|---|---|---|
| D1 | daemon/ is 8 files; only cmd/ imports it | CONFIRMED | `ls daemon`; grep |
| D2 | DirRunner, ExecDirRunner, WorktreeName/slug, Agent/Under, Agents, Spawn used only by daemon/loop.go + `cmd/serve.go:27` | CONFIRMED | grep |
| D3 | delete `cmd/{start,stop,status}.go` + tests | CONFIRMED | They must go in one bar (cross-file test helpers) |
| D4 | only caller of `Store.Counts()` is `daemon/loop.go:66` | CONFIRMED | |
| D5 | delete queue.sql, orm/queue.sql.go, queue_test.go | CONFIRMED | the orm file needs manual deletion (T19) |
| D6 | mcp-notes daemon sections ≈219-232 and 334-447 | **WRONG** | mentions at :78, :125, :219, :252-284 (incl. the command inventory row :284), section "Relationship to the automation phase" :327-345, :373, :442-447 |
| D7 | serve.go removal list | **WRONG** (incomplete) | also `serveDaemonCmdUse`, `claudeBinFallback`, and `cmd/cmd_test.go:58 runWithContext` (lint `unused`) |
| D8 | keep `TestServeMCPCmd_IsListedInServeHelp` ~:158 | CONFIRMED | :158 |
| D9 | root.go :125, :128, start/status/stop wiring :143-145 | **WRONG** (lines) | :125, :128 right; wiring is **:147-149** |
| D10 | `claude.Runner` stays via `writeClaudeWiring` `cmd/init.go:51` | CONFIRMED | |
| D11 | remove runWithDaemon, nothingLoaded; `root_test.go:216` row | CONFIRMED | |
| D12 | `cmd/tui.go` remove queueFuncs + db.Open block 31-94, then no db/orm/os; no test file | CONFIRMED | `context` import goes too |
| D13 | tui model removals (fields :89-91, With* :174-181, reload branch, applyQueued, enqueueSelected, "e") | **WRONG** (incomplete) | also `board.go:226` "e", the playPrompt fields :96-98, `pendingApprovalID` and the 7 dead helpers (lint) |
| D14 | Play prompt "y" only enqueues, `model.go:319-336` | CONFIRMED | |
| D15 | `queued` param board.go, `queuedColor` delegate.go | CONFIRMED | `board.go:78`, `delegate.go:19,24` |
| D16 | tests to cut: model_test 38-160 & 1456-1697, board_test:877, delegate_test:278-300 | **WRONG** (ranges) | model_test **38-169**, **1456-1699**; board_test also :237,246-248,259 (`nil` arg); delegate_test ends :302 |
| D17 | `db/open_test.go:42` asserts 4 migrations | CONFIRMED | :41; unaffected by BIT-45 (no drop migrations) |
| D18 | UpdateProjectCounts callers: daemon, list_test 90-100, status_test | CONFIRMED | `loop.go:73`, `list_test.go:90-100`, `status_test.go:184-194` |
| D19 | GetProjectByPath exact match, TUI only | CONFIRMED | non-test: `cmd/tui.go:60`; tests: `daemon/loop_test.go`, `db/queue_test.go` |
| D20 | `store.Dir()` stays (`db/open.go:20`) | CONFIRMED | |
| D21 | `bp add` flow (ProjectExists, config prefix default, prompt, wiring if no .bit, CreateProject) | CONFIRMED | `cmd/add.go:40-77` |
| D22 | list counts at `cmd/list.go:32-33` | **STALE** | :33-34 |
| D23 | `list_test.go:117 seedProject` shared with serve/status | CONFIRMED | |
| D24 | plist exists, not loaded | CONFIRMED | |
| D25 | no Justfile/scripts target references the daemon | CONFIRMED | grep |
| D26 | keep `task.ParentID`/`barParent` (shared) | **WRONG** | `barParent` is shared; `task.ParentID` (`task/store.go:548`) has no caller except `daemon/loop.go:211`, and no test |
| D27 | removal deletes the only exact-path project lookup | CONFIRMED | GetProjectByPath |

## bit-path-inventory
| # | Claim | Verdict | Evidence / correction |
|---|---|---|---|
| P1 | bitdir `defaultDir=".bit"`; Current/Canonical/ForRoot/Root | CONFIRMED | `bitdir/bitdir.go:12,17,28,41,49` |
| P2 | `cmd/root.go:136` Resolve, `:27` Root | CONFIRMED | |
| P3 | `cmd/init.go:39,75` | CONFIRMED | |
| P4 | approve, feedback_add, `tui.go:24` use `task.New(bitdir.Current())` | CONFIRMED | approve.go:15,26; feedback_add.go:19 |
| P5 | `cmd/task/*` use `taskstore.New(bitdir.Current())` | CONFIRMED | all 7 files |
| P6 | serve_mcp 10 handlers :294..489 use ForRoot | CONFIRMED | |
| P7 | descriptions mention .bit paths at :76,84,102,109 | CONFIRMED | |
| P8 | `cmd/add.go:51,66` hardcode `.bit` | CONFIRMED | |
| P9 | `daemon/loop.go:64` hardcodes `.bit` | CONFIRMED | deleted by BIT-45 |
| P10 | tui/ has no `.bit` reference | CONFIRMED | |
| P11 | retro SKILL :18,29 read and :76,110 write `.bit/` directly | CONFIRMED | |
| P12 | learn SKILL :3 expects a `.bit/retro` path | CONFIRMED | |
| P13 | bot-dev :27-28 stages `.bit/`; do SKILL :99 | CONFIRMED | |
| P14 | plugin.json says "tracked in .bit/" | CONFIRMED | `bit/.claude-plugin/plugin.json:5` |
| P15 | README has a `.bit/` tree at ~110 | CONFIRMED | README.md:110 |
| P16 | `update/normalize.sh` + `normalize_test.sh` | CONFIRMED | |
| P17 | hooks and Justfile have no `.bit` refs | CONFIRMED | grep counts 0 |
| P18 | `.bit/` tracked, ~370 files: 321 completed, 27 feedback, 10 tasks, 8 archive | CONFIRMED | `git ls-files .bit`: 371 (+4 research, +1 config); BIT-45..50 tasks/research are untracked |
| P19 | ~11 test files t.Chdir: init_test 11, add/serve/start 7 each, loop_test 5 | **WRONG** | t.Chdir in 7 files: init_test 11, root_test 2, add_test 2, cmd_test 1, task/helpers_test 1, task/create_test 1, bitdir_test 1. serve/start/loop_test have 0. The real seam is `initProject` (80 calls) |
| P20 | `bitdir_test.go` pins worktree cut | CONFIRMED | bitdir_test.go:20-23 |

## coexistence
| # | Claim | Verdict | Evidence |
|---|---|---|---|
| C1 | install.sh builds `$(GOBIN or GOPATH/bin)/bp`, then `marketplace add` | CONFIRMED | scripts/install.sh |
| C2 | here `/Users/appstack/go/bin/bp` | CONFIRMED | `which -a bp` |
| C3 | `claude mcp add bit -- bp serve mcp` at `claude/sync.go:26` | CONFIRMED | |
| C4 | registered for bit-pro and a client project | CONFIRMED | `~/.claude.json` projects; no user-scope `bit` |
| C5 | plist pins os.Executable, label, log in bit-pro/daemon.log, env only BP_CLAUDE | CONFIRMED | `cmd/start.go:49,69`, `daemon/daemon.go:13`, `daemon/plist.go:39-41` |
| C6 | marketplace source github, no ref | CONFIRMED | `~/.claude/plugins/known_marketplaces.json` |
| C7 | `v2 == main` at 6a1d345 | CONFIRMED | `git rev-parse main v2` |
| C8 | install stamps `cmd.version` from plugin.json | CONFIRMED | install.sh `-ldflags -X …cmd.version` |
| C9 | tool names `mcp__bit__*` come from the server name | CONFIRMED | sync.go:26 |
| C10 | `~/.local/share/bit/` doesn't exist yet | CONFIRMED | ls |

## config
| # | Claim | Verdict | Evidence |
|---|---|---|---|
| G1 | Config has one key, `prefix` | CONFIRMED | task/config.go:12 |
| G2 | SaveConfig rewrites only prefix | CONFIRMED | :31 |
| G3 | `Store.Create` reads Config for track IDs (`task/store.go:216`) | CONFIRMED | :213-221 |
| G4 | init writes/reads it | CONFIRMED | init.go:39,75 |
| G5 | add.go:51 uses it as the default code | CONFIRMED | |
| G6 | normalize.sh uppercases it | CONFIRMED | update/normalize.sh:73-78 |
| G7 | bit-pro's file is `prefix = "BIT"` | CONFIRMED | .bit/config.toml |
| G8 | the daemon logs by `p.Code` | CONFIRMED | daemon/loop.go |

## delivery-order
| # | Claim | Verdict | Evidence |
|---|---|---|---|
| O1 | the banner's order 45→46→48→49→47→50 | CONFIRMED | v2-sketch.md:30-36, 96-101 |
| O2 | Justfile `run` target exists | CONFIRMED | `run *ARGS: db-gen-queries` |
| O3 | `bp init` is deleted in BIT-46 Verse 1 | CONFIRMED | BIT-46 body |

## project-resolution
| # | Claim | Verdict | Evidence / correction |
|---|---|---|---|
| R1 | no walk-up; Resolve in PersistentPreRunE; relative `.bit` unless in a worktree | CONFIRMED | bitdir.go:21-47, root.go:135-139 |
| R2 | CLI callers list | CONFIRMED | (P4/P5) |
| R3 | `Root()` has one caller, root.go:27 | CONFIRMED | |
| R4 | `serve_mcp.go:226` reads CLAUDE_PROJECT_DIR once | **STALE** | it's **:231** (newServeMCPCmd RunE); still read once |
| R5 | handlers resolve per call | CONFIRMED | |
| R6 | `runMCPServer(ctx, root, transport)` seam | CONFIRMED | serve_mcp.go:237 |
| R7 | opencode.json registers `bp serve mcp` with no env | CONFIRMED | note: opencode.json and `.opencode/` are **untracked** on v2 |
| R8 | `.opencode/bit-list.tsx` runs `bp task list/read` with cwd | CONFIRMED | :6-7,17,50 |
| R9 | the TUI and daemon resolvers disappear with the removal | CONFIRMED | |
| R10 | git isn't used anywhere; the only execs are claude and launchctl | CONFIRMED | sync.go:13, plugin.go:78, dispatch.go:17, daemon.go:22 |
| R11 | ID allocation is glob max+1 with no lock (~625-675) | CONFIRMED | task/store.go:625-675 |

## sqlite-registry
| # | Claim | Verdict | Evidence |
|---|---|---|---|
| S1 | the 4 migrations as described | CONFIRMED | db/migrations/* |
| S2 | sqlc.yaml: sqlite, package orm, out db/orm, gitignored | CONFIRMED | sqlc.yaml; .gitignore:3 |
| S3 | install/test/lint/run depend on db-gen-queries | CONFIRMED | Justfile |
| S4 | query list | CONFIRMED | db/queries/*.sql |
| S5 | the db's users | CONFIRMED | grep |
| S6 | path `$XDG_DATA_HOME/bit-pro/bit.db` | CONFIRMED | store.go:22, open.go:25 |
| S7 | daemon.log beside it | CONFIRMED | |
| S8 | live rows BIT, EX; queue empty | CONFIRMED | sqlite3 -readonly: BIT(2/0/0/38), EX(1/1/0/0), queue 0 |
| S9 | Client project not registered | CONFIRMED | |
| S10 | db/bit.db is a throwaway dev db, gitignored | CONFIRMED | .gitignore:4 |
| S11 | embedded migrations; CreateAndMigrate; Strict false; no dump; log discarded | CONFIRMED | open.go:16-35; dbmate v2.35.0 `db.go:85` default Strict=false |
| S12 | modernc driver; DumpSchema errors | CONFIRMED | driver.go:16,20,72 |
| S13 | Justfile dbmate targets | CONFIRMED | |
| S14 | `NormalizeID` uppercases | CONFIRMED | task/store.go:69-71 |
| S15 | `store.Dir` MkdirAll's on every call | CONFIRMED | store.go:23 |

## testing
| # | Claim | Verdict | Evidence / correction |
|---|---|---|---|
| X1 | install.sh always builds `…/bp` | CONFIRMED | |
| X2 | `just run` exists | CONFIRMED | |
| X3 | "v1 and v2 data dirs already differ; a v2 dev build can't corrupt v1's registry" | **STALE** | On v2@6a1d345 `store.Dir` still uses `bit-pro` (store.go:22). Until BIT-46 a v2 build opens v1's live db, as the track body says. Sandbox HOME/XDG_DATA_HOME |
| X4 | store.Dir honours XDG_DATA_HOME | CONFIRMED | |
| X5 | `claude mcp add -e KEY=value`, `-s` local default | CONFIRMED | Claude Code 2.1.287 `--help` |
| X6 | `--mcp-config`, `--strict-mcp-config`, `--plugin-dir` exist | CONFIRMED | `claude --help` |
| X7 | RegisterMCP is a no-op when `mcp get bit` succeeds | CONFIRMED | sync.go:22 |
| X8 | runMCPServer + `cmd/mcp_harness_test.go` | CONFIRMED | |
| X9 | `tools/example` reset relies on git-tracked `.bit` (reset.sh:8,129) | CONFIRMED | |
| X10 | plist present, not loaded | CONFIRMED | |

## xdg
| # | Claim | Verdict | Evidence |
|---|---|---|---|
| Y1 | `store.Dir` implements XDG with a `~/.local/share` fallback | CONFIRMED | store.go:11-28 |
| Y2 | used by `db.Open` and `cmd/start.go` | CONFIRMED | open.go:20, start.go:59 (only db.Open after BIT-45) |
| Y3 | no xdg lib in go.mod | CONFIRMED | |
| Y4 | tests sandbox via XDG_DATA_HOME (store, db, cmd add/list/serve/start/status, daemon loop) | CONFIRMED | 10 files |
| Y5 | plist sets only BP_CLAUDE | CONFIRMED | plist.go:39-41 |
| Y6 | relative XDG_DATA_HOME accepted (Clean only) | CONFIRMED | store.go:12,22 |
| Y7 | `os.UserConfigDir` on macOS = `~/Library/Application Support` | CONFIRMED | stdlib behaviour |

## Would stall planning or implementation
- Lint failures from unused leftovers (listCompleted, runWithContext, 7 tui helpers): a bar that follows the Touches list exactly passes `go test` but fails the post-edit hook, the stop hook and pre-commit.
- The `db/orm` regeneration gap (T19), and fresh worktrees having no `db/orm`.
- Undecided: whether `GetProjectByPath` is deleted now (it's dead after the TUI bar). The trial deleted it, green.
- BIT-49's git helper loses its ready-made model (DirRunner) unless its plan says to copy it from 6a1d345.
