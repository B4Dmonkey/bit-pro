# Claim audit + build map, 2026-10-02 (v2 @ 6a1d345, git 2.54.0)

Every factual claim in the BIT-50 body and topics `completion-today`, `landing-ladder`, `done-classification`, `merge-commit`, `fields-and-writes`, `verse-order`, `decisions`, `index` checked against code and read-only git. Scratch repos under the session scratchpad. Baseline `go build ./... && go test ./...`: green.

**Tally: 52 CONFIRMED, 3 WRONG, 12 STALE.**

## WRONG
1. **`landing-ladder` (Rung a, "Landing commit for a bar") and `merge-commit`**: `git rev-list --first-parent --ancestry-path <h>..<trunk> | tail -1` returns **empty** for a bar committed before the branch back-merged trunk. Scratch: main m1,m2,m3,`Merge branch 'feat'`,m4; feat b1,`Merge branch 'main' into feat`,b2 → b1 = `[]`, and `git log -1 $(empty)` = m4 (HEAD, wrong anchor). Back-merges are real here: `97a5279 Merge remote-tracking branch 'origin/main' into worktree-bit-39`. Correct rule (verified: b1,b2 → `Merge branch 'feat'`, m3 → m3): landing(h) = oldest X in `git rev-list --first-parent --reverse <trunk>` with `merge-base --is-ancestor h X` = 0. (`review-2026-10-02` already says this.)
2. **Body Decision + Verse 3 "Squash matching needs every bar's subject in one squash"** as a way to complete squash-landed tracks: BIT-31 (`.bit/completed/BIT-31{,.1-.4}.md`) landed bars 1-3 in `4cf6137 (#5)` (all 3 `main..origin/worktree-bit-31` subjects grep to #5 only) and bar 4 as `e59da6e (#6)` ("Closes BIT-31.4"). One-squash rule can't complete it.
3. **Body "covers ... 304 of 320 commits" as Verse 1's reach** — count is true (`git log main` 320, 0 merges, 16 `(#N)`, shortlog -c: 304 josiah / 16 GitHub) but no v1 record has a hash (`task/task.go:19-28` has no `commit`), so Verse 1 completes only tracks whose bars carry BIT-47 hashes or BIT-49 migration HEAD. Historical commits are irrelevant to reach.

## STALE
- `completion-today`: "Unknown: whether sign-off still sets done" → settled (stays `doing`, body Decisions).
- `landing-ladder`: "unknown whether to support non-`main` trunks" → settled (main only). "earliest vs latest unknown" → settled (earliest).
- `done-classification`: "mixed landed + unresolvable → operator decision" → settled (partly done). Missing the "status ≠ done" input the body added.
- `fields-and-writes`: `.bit/tasks/BIT-46.md:60-62` → now `:63-66` (git fields `:64`); `.bit/tasks/BIT-47.md:26-28` → `:35` (task_update/task_complete accept commit/branch) and `:37`; `taskUpdateInput` `:150-157` → `:151-158`; "track branch unknown" → trunk; "Go vs skill open" → `task_landing`.
- `review-2026-10-02`: task_delete handler `:433-441` → `:433-447`.
- **Body Verse 1 Touches incomplete** (see edit map): misses do `:3,:15,:100,:106`, complete `:3,:8`, bot `:54`, bot-dev `:45`, `cmd/serve_mcp.go:76-82,265-268`, `cmd/task/complete.go:12`, `README.md:65,115`, `cmd/serve_mcp_test.go:201`.
- **Cross-track:** `.bit/tasks/BIT-47.md:37` says completion records *HEAD* for tracks; BIT-50 records the *landing* commit. Wording conflict for BIT-47's scope.

## CONFIRMED (evidence)
- `Store.Complete` `task/store.go:108-110` → `relocateTree` `:112-147`; guard `:118-130` (`UnfinishedBarsError` `:96-102`); renames bars then track `:132-140`; never sets status, never git. **New fact:** it doesn't check the *track's* own status — a `doing` track with done bars files fine (skill step 3 sets it).
- `Path`/`Load` read only `tasks/` (`:37-39`, `:164-171`); `Update` starts with `Load` (`:273`) → filed tasks can't be updated; repoint before `task_complete`. Approval revoke `:298-303`.
- complete SKILL: forces bars `:21`, sets track done `:22`, files `:23`, verifies `:24`, suggests commit, user commits `:30`. Evals 1 (commit suggested) & 2 (forced without asking) as stated; eval 3 survives.
- do `:95` (rollup never sets done), `:102-110` sign-off, hand-off `:108`, `:131`. bot `:41`. serve_mcp `:160-162`, handler `:417-431`, desc `:76-82`.
- Ladder facts: `main`=`origin/main`=`v2`=`6a1d345`; `origin/HEAD → origin/main`; not shallow; 9 squashed `origin/worktree-*` refs all is-ancestor exit 1; `cebf42d` resolves, not ancestor, `branch -r --contains` → `origin/worktree-agent-a6848e37a1646d392`; exit codes `cat-file -e` bad=128, `merge-base --is-ancestor` 0/1/128 (scratch); subject grep → `54abfeb (#17)`,`ad04ca4 (#16)`; both 11 `* ` lines, `git diff --stat ad04ca4 54abfeb` empty; `98597f7 (#14)` 9 lines ⊂ `84886bb (#15)` 16; `e59da6e (#6)` no `* ` list; `(#N)` unique for 1,2,3,6,16,17, none for 18; `chore(bit): file completed BIT-34` twice on main (`c6ca8ba`,`65c0228`) and twice on worktree-bit-35; `backup/main-prerebase`: 4 only-backup, 135 only-main, merge-base `fb11ade`; last squash `54abfeb` 2026-08-29, 20 direct since; 0 `Merge pull request`.
- `is-ancestor` true for a bar merged via `--no-ff` (scratch). `rev-list -n1 --topo-order` newest holds for related commits (superseded by first-parent index).
- No git helper in tree (`grep '"git"'` empty). Reusable pattern: `claude.DirRunner` (`claude/dispatch.go:14-30`) returns `(out, exitCode, err)` — needed for is-ancestor's 0/1/128.
- MCP root = `CLAUDE_PROJECT_DIR` (`cmd/serve_mcp.go:231`); `bitdir.ForRoot` cuts worktrees to main checkout `.bit` (`bitdir/bitdir.go:49-55`) → git must run in the handler with the session dir.
- Trunk `main`: both `.bit` projects under ~/Developer (bit-pro: origin/HEAD→origin/main; tools/example: local `main`) — "every project" unverifiable further.
- Order 45→46→48→49→47→50, BIT-49 HEAD stamping, BIT-47 commit/branch inputs (`BIT-47.md:35`), BIT-46 store outside repo (`BIT-46.md:13`), `bp remove` archives done-not-completed tracks (`BIT-46.md:70`), v2-sketch `:58` — all match.

## Build map (task 2)

### Call sites and tests
| Site | Today | BIT-50 change |
|---|---|---|
| `task/store.go:108-110` `Complete`, `:112-147` `relocateTree` (also `Relocate` `:104-106`) | file-move only | **None expected.** BIT-46 rewrites the store; BIT-47 adds commit/branch to the complete path. BIT-50 consumes them. |
| `cmd/serve_mcp.go:24` name, `:76-82` desc, `:160-162` input, `:265-268` register, `:417-431` handler | ID only | desc drops "signed-off" (BIT-49 owns `.bit/` wording); input/handler gain commit/branch in **BIT-47**. New `task_landing` const + desc + input/output + handler + `AddTool` beside it. |
| `cmd/task/complete.go:12` Short, `:15` RunE | files without git | Help text: manual escape hatch, no landing check. **Keep behaviour**: `cmd/feedback_add_test.go:136` and `cmd/serve_mcp_research_test.go:124` use complete as a fixture. |
| Tests: `task/store_test.go:394,420` (Complete), `:44-258` (Relocate/relocateTree); `cmd/serve_mcp_write_test.go:556,576`, `seedDoneTrack :601`; `cmd/serve_mcp_research_test.go:119-124`; `cmd/serve_mcp_test.go:182-201` (desc test: `task_complete` must keep `testTrackSentence`; add a `task_landing` row); `cmd/task/complete_test.go:13,40,65,104,142`; `cmd/task/delete_test.go:67`; `cmd/feedback_add_test.go:132`. | | No existing test asserts the exact tool set, so adding `task_landing` breaks nothing. |

### Prose edits, and who else edits the same lines
| File:line | BIT-50 edit | Also edited by |
|---|---|---|
| complete `SKILL.md:3` desc, `:8`, `:21`, `:22-24`, `:30` | triggers "pushed/merged, complete BIT-N"; call `task_landing`; no forcing; no commit suggestion; check-ins | **BIT-49 sweep** (`.bit/` on `:3,:8,:12,:23,:30`); BIT-47 leaves it (`BIT-47.md:42`) |
| complete `evals/evals.json` 1, 2 | rewrite; add not-done / partly / PR cases | none |
| do `:3`, `:15`, `:95`, `:100`, `:102-110` (esp. `:106,:108`), `:131` | sign-off → "push/merge, then `/bit:complete`"; `:15` lists `task_complete` as a do tool though do never calls it | BIT-49 sweep (`.bit/` on `:3,:8,:15,:43,:75,:95,:99,:104,:108,:131,:132`); BIT-47 (`:3,:17,:75-79,:89-100,:122,:129`) — `:3` and `:100` are triple-edited |
| bot `:41`, `:54` | sign-off ≠ completion | BIT-49 sweep (`.bit/` on `:3,:8,:16,:22`, not 41/54) |
| bot-dev `:45` | optional: "tell operator to push, then `/bit:complete`" | BIT-49 (`:27-28`), BIT-47 (`:3,:12-38`) |
| `README.md:65`, `:115` ("signed-off work"), `:171` (sign-off in CLI question) | wording | BIT-49 (`.bit/` on `:65`, command table) |
| `cmd/task/complete.go:12` | escape-hatch wording | BIT-49 (`.bit/` help text) |

All line numbers will drift before BIT-50 starts; anchor on section names and function names.

### Bar order (each green on build+test; Go bars additive)
Verse 1: (1) landing core in Go on BIT-49's helper: trunk resolve, per-bar is-ancestor, landing(h) by first-parent index — tests in `t.TempDir()` git repos with a bare origin (set `GIT_AUTHOR_*`/`GIT_COMMITTER_*`, `HOME` sandboxed); (2) `task_landing` MCP tool + desc-test row + harness test; (3) skill/agent/README/help prose + evals, validated with `quick_validate.py` and `claude plugin validate ./bit`.
Verse 2: (4) buckets pushed/local/unresolvable/no_hash, unfinished bars, verdicts incl. cant_tell/no_git, shallow flag; (5) skill check-ins, archive via `task_delete` (force when unfinished), no-git confirm.
Then Verse 4 before 3 (review's 1,2,4,3): (6) `pr`/`commit` input, `(#N)` subject-suffix lookup on first-parent; (7) skill PR question. Verse 3: (8) squash rung (per-bar if scope adopts it), `bars[].landing`; (9) skill repoints bars via `task_update` before `task_complete`.

## Would stall planning / implementation
- Scope still says one-squash rule and order 1,2,3,4; `task_landing` I/O, per-bar squash, no-git verse, and 1,2,4,3 are open in `review-2026-10-02`, not in the body. bit:plan would have to guess.
- Who makes `task_complete{commit,branch}` write *before* the move (Update can't touch filed tasks) — BIT-47 owns the input; confirm BIT-47 implements write-then-move, else BIT-50 inherits it.
- BIT-47.md:37 "completion records HEAD" vs landing commit.
- Go seams (`Store`, handler signature `taskCompleteHandler(root)`, `bitdir.ForRoot`) are all replaced by BIT-46 (per-call resolver, JSON store); plan against post-BIT-46 code, not these lines.
- Dev can't dogfood: nothing new on `v2` reaches `origin/main` before cutover; scratch repos are the only test path.
