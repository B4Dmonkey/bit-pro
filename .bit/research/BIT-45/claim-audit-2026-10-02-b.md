# Claim-by-claim audit (2026-10-02, v2 = main = origin/main @ 6a1d345)

**Checked:** every factual claim about code, repo, machine, skills or git history in topics `history-anchors`, `json-schema`, `migrate`, `skills`, `soundness`, `soundness-2`, `decisions`, `index`, and in `v2-sketch.md`. Each one was re-read against the code (grep + reading files), the machine (`ls`, `launchctl list`, `sqlite3 -readonly`), and read-only git (`git log/branch/ls-remote/rev-parse`). Superseding sources are the BIT-46..BIT-50 track bodies. Verdicts: **C** = CONFIRMED, **W** = WRONG, **S** = STALE (true when written, no longer), **X** = SUPERSEDED (by a later decision), **U** = UNVERIFIED (can't be settled read-only). Design proposals and opinions aren't audited. Topic `review-2026-10-02` (written concurrently by another pass) overlaps on BIT-45, and agrees where both checked the same thing.

## Tally
- CONFIRMED 119, WRONG 14, STALE 3, SUPERSEDED 13, UNVERIFIED 1. A row with a split verdict ("C / X") is counted once, by its first verdict.
- Every WRONG and STALE row carries its correction inline. The cross-source contradictions and the stall risks are collected after the sketch table.

---

## history-anchors
| # | Claim | V | Evidence / correction |
|---|---|---|---|
| H1 | bp never calls git | C | no `"git"` exec anywhere. `exec` sites are only `claude/sync.go:13`, `claude/dispatch.go:17`, `claude/plugin.go:78`, `daemon/daemon.go:22`, plus `cmd/start.go:54` (`LookPath("claude")`) |
| H2 | Only `exec`s are claude (sync/plugin/dispatch) and launchctl (daemon.go) | C | same as above. `start.go:54` is a LookPath, not an exec |
| H3 | No record carries a timestamp or commit | C | `task/task.go:19-28`: no time or commit fields |
| H4 | PRs squash-merged, subject `<title> (#N)`, e.g. `84886bb Worktree bit 39 (#15)` | C | `git show -s 84886bb` |
| H5 | Squash body lists every branch commit subject (`* feat(daemon): …`) | C | `84886bb` body lists `* refactor(daemon)…`, `* feat(daemon)…` |
| H6 | Branches are named after the worktree; `origin/worktree-*` branches remain | C | `git branch -r`: `worktree-bit-31/32-1/35/35-merge/36/39/40`, `worktree-agent-a6848…`, `worktree-zesty-strolling-raccoon` |
| H7 | GitHub keeps `refs/pull/N/head` (marked "verify") | C | `git ls-remote origin 'refs/pull/15/*'` → `97a5279… refs/pull/15/head` |
| H8 | `origin/HEAD -> origin/main` set; `symbolic-ref` gives the default | C | `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/main` |
| H9 | Capture table (branch/head/main head/base on every write) | X | BIT-46 Decisions (tracks/bars carry `branch` + one `commit`; research/feedback/retro carry `commits`), BIT-47 (no per-write HEAD for tracks/bars) |
| H10 | Squash commit not knowable at `task_complete`; open design question | X | BIT-50: completion runs after landing, through the ladder |
| H11 | bit:do sets `done`, then the user commits (off by one) | C (now) / X (v2) | `bit/skills/do/SKILL.md:75,99`. v2: commit first, hash after (BIT-47) |
| H12 | bot-dev commits after the status change | C | `bit/agents/bot-dev.md:20` |
| H13 | Local `main` can be stale; never fetch | C (design) | matches BIT-50 "bp never fetches" |
| H14 | `worktreeCut` exists because CLAUDE_PROJECT_DIR is the worktree path; pinned by `bitdir_test.go` | C | `bitdir/bitdir.go:57-68`, `bitdir/bitdir_test.go:19-23` |
| H15 | "Record the session dir as well as the project path" | X | BIT-46: paths live only in the db. BIT-49: git facts come from the session dir but aren't stored as a path |
| H16 | Migrate recovers anchors via `git log --follow`, else empty + `migrated` | X | BIT-49: HEAD + branch at migration, migration time |
| H17 | Non-git projects get empty git fields; a git error never fails a write | C (design) | matches BIT-47/49 |

## json-schema
| # | Claim | V | Evidence / correction |
|---|---|---|---|
| J1 | Sample: tasks 11 | S | `.bit/tasks` now has 16 files (10 tracked; BIT-45..50 untracked) |
| J2 | completed 321, archive/tasks 8, feedback 27 | C | `ls` counts, all tracked |
| J3 | research 12 topics | S | 54 `.md` under `.bit/research/` now (4 tracked) |
| J4 | No `retro/` dir | C (bit-pro) | but `<client>/.bit/retro/ACME-1-ACME-4-proposals.md` exists (see the migrate gap) |
| J5 | Task struct `task/task.go:19-28`, keys id,title,status,approved?,phase?,phase_label?,order? | C | — |
| J6 | `Parse` :30, `Bytes` :58, hand-rolled `---` split | C | — |
| J7 | Feedback path `feedback/<TRACK>-NNN.md` (`feedback.go:58`); no metadata | C | `AddNote` :58, format at :20 |
| J8 | Research `research/<TRACK>/<topic>.md` (`research.go:37`); no metadata | C | — |
| J9 | 330 completed/archived files: id/title/status 330, phase/phase_label 285, approved 107, order 2 | W | 329 files: id/title/status **329**, phase 285, phase_label 285, approved 107, order 2 (`grep -l '^key:'` over completed + archive/tasks) |
| J10 | No timestamps anywhere today | C | — |
| J11 | `.md` hardcoded at store.go :38,:46,:54,:62; globs :443,:467,:614,:626 | C | — |
| J12 | `feedback.go:20,24`, `research.go:21,95-96` | C | — |
| J13 | `relocateTree` :112; Complete :108 refuses unfinished; Relocate :104 allows force | C | `task/store.go:104-147` |
| J14 | ID minting `store.go:637-650` regexes across tasks/completed/archive | C | :640 loops over all three dirs |
| J15 | `trackExists` `feedback.go:36` | C | func at :35, loop at :36 |
| J16 | MCP JSON structs `serve_mcp.go:116-220`; `body` is a string | C | :116-222 |
| J17 | Descriptions naming `.bit/` at :76, :84, plus research | C | research at :102, :109 |
| J18 | `taskReadOutput` omits `order` | C | `serve_mcp.go:213-222` |
| J19 | 45 test files | C | `git ls-files '*_test.go' | wc -l` = 45 |
| J20 | 18 write `.md` paths or fixtures | C | 18 files match WriteFile/`.md`/tasks-path |
| J21 | 9 assert on frontmatter text, incl. `cmd/feedback_add_test.go`, `cmd/serve_mcp_write_test.go` | W | **7**: `task/{task,store,counts}_test.go`, `cmd/task/{create,update,list,complete}_test.go`. The feedback and write tests read raw `.md` bytes or paths (e.g. `feedback_add_test.go:33`). They break on path or extension changes, not on frontmatter |
| J22 | `completed/` is read only for `Counts` | W | its contents are listed only by `listCompleted` (:442, caller `counts.go:38`), but `completed/` is also scanned by `highestReserved` (:640) and `trackExists` (`feedback.go:36`). So it reserves IDs exactly like `archive/` |
| J23 | `List()` reads `tasks/` only | C | :466-467 |
| J24 | Common fields include project path, `head`, main hash | X | BIT-46 Decisions |
| J25 | do :63-89 / bot-dev :20-22 branch on "User verifies"; do :94 flips verse boxes | C | — |

## migrate
| # | Claim | V | Evidence / correction |
|---|---|---|---|
| M1 | Target `~/.local/share/bit/bit.db` | X | `main.db` (BIT-46, sketch :42) |
| M2 | `<code>` dir casing open; suggests lowercase | X | uppercase (BIT-46, sketch :57) |
| M3 | IDs are uppercase via `task.NormalizeID` | C | `task/store.go:69-71` (`strings.ToUpper`) |
| M4 | `config.toml` holds `prefix`, one key | C | `task/config.go:11-13`. bit-pro `prefix = "BIT"`, example `EX`, a client project `ACME` |
| M5 | feedback has no frontmatter, track+seq in filename; research track/topic from path | C | — |
| M6 | `retro/*-proposals.md` if present; bit-pro has none | C | but a client project has one (gap below) |
| M7 | Unknown files: copy raw or refuse, undecided | X | stop and list (BIT-49). FYI: none of the 3 v1 projects (bit-pro, example, a client project) has a file outside the known shapes today |
| M8 | v1 registry rows don't carry over; a client project was never registered | C | `sqlite3 -readonly ~/.local/share/bit-pro/bit.db`: only `BIT` (bit-pro) and `EX` (example) |
| M9 | In a Claude worktree, v1 kept state in the main checkout's `.bit/` | C | `bitdir.worktreeCut` :57-68, `ForRoot` :49-55 |
| M10 | Timestamps `git log --follow` / mtime; anchors empty + `migrated: true` | X | BIT-49: migration time; HEAD + branch at migration |
| M11 | ID reservation scans all three dirs (`store.go:637-650`) | C | :640 |
| M12 | `cmd/root.go execute()` prints a post-command stderr notice and skips `bit.quiet` commands (`tui`, `serve mcp`) | C | `root.go:50-68,70-76`; annotations at `tui.go:22`, `serve_mcp.go:229` |
| M13 | MCP option `mcp.ServerOptions.Instructions` exists | C | go-sdk v1.7.0 (`go.mod:13`), `mcp/server.go:71` |
| M14 | Step 6: bit-pro `.bit/` tracked; "other projects, `.bit/` is untracked" | W | `tools/example` tracks 9 `.bit/` files and `<client>` tracks **139** (`git -C … ls-files .bit`). Every known v1 project tracks `.bit/`. (Memory `bit-store-tracked-only-here` is also wrong on this.) |
| M15 | Only v2 warns; v1 untouched | C | `main` == `v2` == 6a1d345; nothing diverged yet |

**Gap (not a claim):** a client project's `retro/ACME-1-ACME-4-proposals.md` collides with BIT-49's naming rule `<CODE>-<track-or-album>-proposals`. Neither BIT-49 nor `migrate` says whether migrate renames existing proposals (→ `ACME-ACME-1-ACME-4-proposals`?) or keeps the v1 name. It's a client project, so the operator has to OK any trial against it.

## skills
| # | Claim | V | Evidence / correction |
|---|---|---|---|
| K1 | retro lists/reads `.bit/feedback/*.md` directly (:18, :29 "no tool of their own") | C | `bit/skills/retro/SKILL.md:18,29` |
| K2 | retro writes `.bit/retro/<track-or-album>-proposals.md` (:27, :76) | W (:27) | the write path is at :76 only. :27 is about reading via tools |
| K3 | retro globs `.bit/retro/*-proposals.md` (:110) | C | — |
| K4 | learn takes a `.bit/retro/` path or pasted content (:3, :16) | C | :3 names the path, :16 "path or its contents" |
| K5 | check writes `<track-id>-check.md` in the repo root (:101), re-reads it (:154-155) | C | — |
| K6 | check says retro consumes it (:8, :101); retro never reads it | C | `retro/SKILL.md` has no `check.md` reference |
| K7 | bot-dev stages `.bit/` (:27-28) | C | — |
| K8 | do :99 says `.bit/tasks/*.md` changes join the commit | C | :75 says the same thing (not listed) |
| K9 | Cosmetic `.bit/` mentions: feedback :3,:12,:113; do :8,:104,:108,:131-132; complete :3,:8,:12,:23; scope :10,:110,:205,:227; analyze :3,:17,:64; plan :8; ruler.md:12; bot.md:8,16,22; README; plugin.json:5 | W (incomplete) | the listed lines are right, but these are missing: do :3, :15, :43, :75, :95; plan :3, :10; scope :3, :20; check :12; complete :30; analyze `evals/evals.json:14`; ruler.md:10; bot.md:3 (`grep -n '\.bit' bit/`) |
| K10 | README `.bit/` lines 7, 34, 42, 59, 65-67, 103, 110, 137, 141 | C | also README:82 "`bp init` … ships seven skills" (there are 9 skills now) |
| K11 | Verse checklists/User-verifies branching are body text (do :94, :63-89, bot-dev :20-22) | C | — |
| K12 | `.opencode/bit-list.tsx` shells `bp task list/read` with cwd, no `.bit` paths | C | `.opencode/bit-list.tsx:6-7,17,50,91` |
| K13 | `opencode.json` registers `bp serve mcp` with no `CLAUDE_PROJECT_DIR` | C | `opencode.json` |
| K14 | Skills ship from GitHub `main` | C | memory `bit-plugin-installs-from-github`; marketplace `B4Dmonkey/bit-pro` |
| K15 | New tools `feedback_list/read`, `retro_write/list/read` | C | adopted in BIT-49 Decisions |

## soundness
| # | Claim | V | Evidence / correction |
|---|---|---|---|
| S1 | Only `cmd/` imports `daemon` (serve, start, stop, status, root + tests) | C | root.go:15, serve.go:12, start.go:13, stop.go:6, status.go:8; cmd_test.go:11, start/stop/status_test |
| S2 | `claude/dispatch.go` exports used only by `daemon/loop.go` and `cmd/serve.go:27` | C | — |
| S3 | `Store.Counts()` one non-test caller `daemon/loop.go:66` | C | — |
| S4 | `store` survives through `db/open.go:20`; other caller `cmd/start.go:59` | C | — |
| S5 | `claude.Runner`/`SyncPlugin`/`RegisterMCP` survive via `writeClaudeWiring` (`init.go:51`), used by add and init | C | `add.go:67`, `init.go:43,51` |
| S6 | TUI enqueue/Play lines: model.go :89-91, :174-181, :198-208, :319-374; board.go :78-84, :227, :272; delegate.go:19-34 | C | — |
| S7 | "No drop migrations" compiles: ListProjects keeps counts, `UpdateProjectCounts` stays generated but dead | X | BIT-45 Decision: trim `ListProjects` and delete `UpdateProjectCounts` now |
| S8 | No Justfile/scripts/update/plugin.json/skill references the daemon | C | `git grep -il daemon` |
| S9 | `grep -il daemon` hits "cmd/, tui/, daemon/, db/queue*, claude/dispatch_test.go, automation-notes.md, mcp-notes.md, v2-sketch.md" | W | no file in `tui/` or `db/` contains "daemon". Tracked hits: `cmd/`, `daemon/`, `claude/dispatch_test.go`, `automation-notes.md`, `mcp-notes.md`, plus `.bit/` records (incl. active `.bit/tasks/BIT-42.md`). `v2-sketch.md` is untracked |
| S10 | `newRootCmd(run, lc daemon.Runner)` at `root.go:128`; callers cmd_test.go (:24-61, :128), root_test.go:223,243, task_test.go:18 | C | cmd_test.go:31,46,61,128 |
| S11 | `cmd/list_test.go:67-100` counts test must go | W (range) | `TestListCmd_ShowsProjectCounts` spans 67-115, and `TestListCmd_PrintsProjectsByCode` (:17) also expects count text (:30-32) |
| S12 | `seedProject` (`list_test.go:117`) shared with serve_test/status_test | C | 4/4/2 call sites |
| S13 | `cmd/add.go:24` Short mentions the daemon | C | — |
| S14 | `.gitignore` has `db/orm/`; stale `queue.sql.go` can linger | C | `.gitignore:3`; `db/orm/queue.sql.go` exists locally |
| S15 | "Verify sqlc removes stale output (unverified)" | C (settled) | topic `review-2026-10-02`: sqlc does not prune |
| S16 | Fresh checkout needs `just db-gen-queries` first | C | `Justfile:5-6`; `run`/`test`/`install`/`lint` depend on it |
| S17 | Interim db path `store/store.go:22`, `db/open.go:25` = v1 live db | C | — |
| S18 | Nothing removed is needed by BIT-46/47 | C | — |

## soundness-2
| # | Claim | V | Evidence / correction |
|---|---|---|---|
| T1 | Plist present, not loaded | C | `ls ~/Library/LaunchAgents` (Aug 28 file); `launchctl list` has no bit-pro label |
| T2 | `add.go:24` Short | C | — |
| T3 | `task_test.go:18` passes `nothingLoaded` | C | — |
| T4 | `TestListCmd_ShowsProjectCounts` ~:67 | C | — |
| T5 | `db/orm/` gitignored (`.gitignore:3`) | C | — |
| T6 | `serve.go:74,89` daemon subcommand | C | — |
| T7 | `tui/board.go:227,272`, `model.go:89,174,319` | C | — |
| T8 | `cmd/tui.go:32-65` (`queueFuncs`, `listQueue`) | W (range) | vars at :31-34, wiring :36-46, `queueFuncs` is **:51-94** |
| T9 | `claude.Runner` still needed by add/init (`init.go:51`) | C | — |
| T10 | `ListProjects` selects `backlog, todo, done, completed`; `UpdateProjectCounts` exists; `list.go:33` changes | C | `db/queries/projects.sql:8,11`, `cmd/list.go:33-34` |
| T11 | Tests outside cmd: `db/queue_test.go` breaks, **`db/queries_test.go`**, `claude/dispatch_test.go`, `task/counts_test.go`, `tui/*_test.go` (`board_test.go:877`), `cmd/{start,stop,status}_test.go` | W (queries_test) | `db/queries_test.go:9-48` uses only Path/Code/ID and survives the trim. The rest is C. **The error was carried into the BIT-45 body's Tests line** (`db/{queue,queries}_test.go`) |
| T12 | `WithListQueue` `cmd/tui.go:33,46` | C | — |
| T13 | Dev `bp add/list` without XDG sandbox hits v1 db | C | — |

## decisions
| # | Claim | V | Evidence |
|---|---|---|---|
| D1 | Cutover is the operator's; no track gates it | C | every track body says so |
| D2 | Split of BIT-46 into 46/48/49; `bp init` deleted in BIT-46 | C | BIT-46 Decision "bp init is deleted in Verse 1", BIT-48 Summary |
| D3 | Order 45→46→48→49→47→50; 49 reuses 48's wiring; 47 extends 49's git helper | C | BIT-47/49/50 Decisions; sketch :31-37, :96-101 |
| D4 | BIT-47 split; BIT-50 relies on BIT-47's bar hashes | C | BIT-50 Decisions |
| D5 | No-`just install` + sandbox rule stated in every track body | C | BIT-45, 46, 47, 48, 49, 50 all state it |
| D6 | Callers move before packages are deleted (daemon/, dispatch.go; bitdir, task.Config) | C | BIT-45, BIT-46 Decisions |

## index
| # | Claim | V | Evidence / correction |
|---|---|---|---|
| I1 | Superseded list (no paths in records, branch/commit, no per-write HEAD, main.db, uppercase, migration-time, stop-and-list, order) | C | matches BIT-46/47/49/50 and sketch :42-65 |
| I2 | json-schema, history-anchors, delivery-order carry banners; migrate is affected | C | migrate has **no banner**, which matches "(each now carries a banner) and migrate". BIT-49 states it in its body |
| I3 | "fresh `~/.local/share/bit/bit.db`" (second-pass decision list) | X | main.db |
| I4 | bp never calls git; squash not available at task_complete; per-bar off by one | C / X | H1; then BIT-50, BIT-47 |
| I5 | Claude worktrees live inside the repo at `.claude/worktrees/` | C | `.gitignore:2`, `bitdir.go:10-11` |
| I6 | opencode runs `bp serve mcp` with no `CLAUDE_PROJECT_DIR` | C | K13 |
| I7 | Only `task/` knows the format; MCP body is a string; no Go parses checkboxes | C | — |
| I8 | "9 tests assert on frontmatter text and 18 write `.md` fixtures" | W (9) | 7, see J21 |
| I9 | No timestamps; migrate recovers them from git log/mtime | C / X | J10; then BIT-49 |
| I10 | `tools/example` reset restores task state through git-tracked `.bit/` | C | `tools/example/reset.sh:8,88` (`git reset --hard $TARGET`); 9 tracked `.bit/` files |
| I11 | Only retro and learn do direct file I/O; check writes to repo root | C | K1-K5 (check's own repo-root file is direct I/O but not `.bit/`) |
| I12 | Launchd plist on disk, not loaded | C | T1 |
| I13 | Open: whether `git rev-parse --path-format=absolute` exists | S (now answerable) | git 2.54.0 (Apple Git-157); `git rev-parse --path-format=absolute --git-common-dir` works |
| I14 | Open: whether opencode sets the MCP cwd | U | needs an opencode run |
| I15 | Open: `bp add <path>` wiring the cwd (`claude/sync.go:10-13`) | C | `claude.Runner` has no dir; `ExecRunner` :12-19 |

---

## v2-sketch.md
| # | Line | Claim | V | Evidence / correction |
|---|---|---|---|---|
| V1 | :12-13 | `.bit/` may stay as config; storage may move to sqlite + json, TBD | X | decided: JSON + `.md` files, db is registry only, `config.toml` dropped (BIT-46). The section is labelled "scratch", but it now contradicts :42-47 |
| V2 | :19 | Feature flags are one idea | X | :41 hard cutover |
| V3 | :31-38 | six tracks, order, contents | C | track bodies |
| V4 | :42 | `main.db`, `projects` table, sqlc + dbmate | C | BIT-46 |
| V5 | :43 | `~/.local/share/bit/BIT/` | C | BIT-46 |
| V6 | :44 | "Until a project is migrated, bit **warns**" | C in substance | BIT-46: it's a resolver **error** ("run `bp migrate`"), not a warning |
| V7 | :46 | `main` stays as it is; `~/.local/share/bit-pro/` deleted at cutover | C | `main == v2`; :115 |
| V8 | :47-52, :54-58, :63-66 | record format, git fields, commit skill, remove, feedback/retro, wiring, completion, migrate defaults, dev testing | C | match BIT-46/47/48/49/50 decisions |
| V9 | :53 | learn reads every project's proposals; "revisited after cutover" | C | BIT-49 says "revisited later" (no timing). Compatible |
| V10 | :88 | registry at `~/.local/share/bit-pro/bit.db`, BIT and EX, queue empty | C | `sqlite3 -readonly`: rows 1 BIT, 2 EX; `count(queue)=0` |
| V11 | :88 | only add, list, status, serve daemon and the TUI enqueue use it; task CLI/MCP don't | C | `db.Open` callers: `cmd/{add,list,serve,tui,status}.go`. The TUI also reads the queue for display (`cmd/tui.go:39,46`) |
| V12 | :88 | daemon not loaded; log last ran 2026-08-29 | C | `daemon.log` last line `2026-08-29T09:33:33 "stopped"` |
| V13 | :88 | migrations embedded, dbmate applies on open | C | `db/open.go:16-17,27-35` |
| V14 | :88 | `store.Dir()` honours `XDG_DATA_HOME` | C | `store/store.go:12` |
| V15 | :89 | `config.toml` has one key, `prefix` | C | M4 |
| V16 | :90 | no search-up; resolved in four places (CLI, MCP, daemon, TUI) | C | CLI `bitdir.Resolve` (`root.go:136`); MCP `bitdir.ForRoot(CLAUDE_PROJECT_DIR)` (`serve_mcp.go:231,294…`); daemon `p.Path` from the registry; TUI `bitdir.Current()` + `GetProjectByPath(Getwd)` (`cmd/tui.go:24,60`) |
| V17 | :91 | retro **and learn** read **and write** `.bit/feedback` and `.bit/retro` directly | W | retro reads feedback (:18,29) and writes retro (:76). learn only **reads** a proposals path or pasted text (:3,16). It never writes and never touches feedback |
| V18 | :92 | dev build and v1 share the `bp` binary (`just install` overwrites), the launchd label, and GitHub skills | C | `Justfile:24-25` → `scripts/install.sh:8-9` (GOBIN or GOPATH/bin); label const in `daemon/plist.go`. The launchd item goes stale once BIT-45 lands |
| V19 | :96-103 | delivery order; cutover not gated | C | — |
| V20 | :105 | candidate separate tracks: fixture rework, linking, opencode | C | BIT-46 out-of-scope list (it also lists distribution and syncing skills to opencode) |
| V21 | :107 | old topics still say bit.db, lowercase, git-log, copy unknown, project path/session dir/head; last three carry banners | C | M1, M2, M7, M10, H9, H15, J24; banners on json-schema, history-anchors, delivery-order |
| V22 | :112 | cutover: `just release` | W (minor) | the recipe takes a level: `just release <level>` (`Justfile:28`). `just release-push` pushes the tag only (:31-33), so pushing `main` is a separate step, as the sketch says |
| V23 | :112 | plain merge keeps v2 bar hashes valid | C (git semantics) | BIT-50 "v2 itself reaching main through a merge" |
| V24 | :113 | `bp migrate` in bit-pro, example, a client project; the first sets up global wiring | C | BIT-48/49. Client project `.bit/` at `<client>/.bit` |
| V25 | :114 | `git rm -r .bit` **in bit-pro** | W (incomplete) | example and the client project track `.bit/` too (M14). They need `git rm -r .bit` as well, which breaks the example fixture's reset (accepted, out of scope) |
| V26 | :115 | delete `~/.local/share/bit-pro/` and the plist | C | both exist |

## Contradictions between sketch, tracks and notes
1. **BIT-45 body vs code:** Tests line lists `db/queries_test.go` (needs no change) and `cmd/list_test.go:67-100` (actually 67-115, plus `TestListCmd_PrintsProjectsByCode`). Both are inherited from `soundness-2` and `soundness`.
2. **`soundness` S7 vs BIT-45 Decision:** the note keeps `UpdateProjectCounts` "generated but dead", while the track trims it now. The track wins.
3. **`migrate` step 6 and sketch :114 vs disk:** "other projects' `.bit/` untracked" and "`git rm -r .bit` in bit-pro" only. example and the client project both track `.bit/`. BIT-49's per-project printed cleanup is right. The checklist is incomplete.
4. **Sketch :91 vs `skills`:** the sketch says learn reads and writes `.bit/feedback` and `.bit/retro`. `skills` says correctly that learn only takes a path or pasted text.
5. **Sketch :12-13 vs :42-47** (and BIT-46): "`.bit/` may be configuration only / sqlite+json TBD" vs the decided format and the removed `.bit/`.
6. **Sketch :44 "warns" vs BIT-46** "resolver error".
7. **BIT-46 body vs disk (outside the audited topics, found in passing):** "`tools/example/reset.sh:4` … runs `bp init`". Line 4 is a comment describing the `blank` checkpoint as "bp init'd". The script never runs `bp init`. It restores state with `git reset --hard` (:88) and only runs `bp task list` (:134). The conclusion (the fixture breaks) still holds.
8. **Index second-pass block vs top banner:** it still lists "fresh `~/.local/share/bit/bit.db`" and the git-log/off-by-one findings. They're covered by the banner, but they read as current.

## Things that could stall planning or implementation
- **Client project retro proposal naming at migrate** (BIT-49 V2): there's no rule for v1 names like `ACME-1-ACME-4-proposals.md` under the new `<CODE>-…` scheme. It's also a client project, so a trial needs the operator's explicit OK.
- **Cutover step 3 misses example/client-project `.bit/`**, so it would leave tracked `.bit/` in two repos and keep tripping the resolver's "run bp migrate" check.
- **BIT-45 Tests list** has to be corrected before planning, or a bar will "fix" an untouched `queries_test.go` and miss `PrintsProjectsByCode`. `review-2026-10-02` adds more (listCompleted, runWithContext, GetProjectByPath, abort-run.md, the sqlc prune trap).
- **Active track BIT-42** (on hold; "unapproved bars must never read as queued", `.bit/tasks/BIT-42.md`) is about the queued colour that BIT-45 deletes. No note or track mentions it, so it should be closed or archived once BIT-45 lands.
- Nothing else blocks. The settled unknown: git 2.54 supports `--path-format=absolute`.
