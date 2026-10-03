# Claim audit + build map, 2026-10-02 (v2 @ 6a1d345, pre BIT-45/46/48/49)

Every factual claim in the BIT-47 track body and in topics capture, commit-skill, mcp-git-inputs, verse-order, soundness, soundness-2, decisions, index (and review-2026-10-02, which the index leans on) was re-read against code, skills and settings. Read-only. Line numbers are today's; BIT-46/49 shift most of them (STALE section and section 3).

## 1. Verdicts

Tally: ~74 claims checked. **CONFIRMED 64, WRONG 2, STALE 8.**

### WRONG
- **W1. "The operator's Claude Code permissions reject the commit tool"** (track body Decision "always asks"; `decisions` 2026-10-01 Permission; `commit-skill` last bullet). They **ask**, not reject: `~/.claude/settings.json:15` `"ask": [`, `:18` `Bash(git commit *)`, `:19` `git push`, `:20` `git merge`, `:21` `git add` (plus rebase/reset/checkout/stash etc.). No git rule in `deny` and none in `.claude/settings{,.local}.json`. Corrected fact: every `git commit`/`git add`/`git push` raises a permission prompt in default mode. Unverified: whether auto mode or `--bg`/`-p` honours or auto-resolves `ask` (the prose ask covers the former; headless can't answer either).
- **W2. `capture`: "`claude.Runner` ... (`claude/sync.go:12`...)".** `type Runner` is `claude/sync.go:10`; `:12` is `ExecRunner`. `DirRunner` is `claude/dispatch.go:14`, `ExecDirRunner` `:16` (correct).
- Note (not a wrong claim): `CLAUDE_PROJECT_DIR` is read once at `cmd/serve_mcp.go:231` and the store is rebuilt per call (`:294,320,360,384,407,423,439,455,472,489`). `runMCPServer(ctx, root, transport)` (`:238`) has no git parameter, so the git seam has to be threaded through it.

### STALE (true today, wrong by the time BIT-47 runs, or incomplete)
- **S1. Verse 1 Touches** list do `:3,:17,:75-79,:89-100,:122,:129` but miss do `:73` (User-verifies path: "state the bar's suggested commit message ... then stop") and do `:106` ("verified, commit suggested"). Both must change or do contradicts bit:commit.
- **S2. Verse 4 Touches** `task/feedback.go`, `cmd/feedback_add.go`, `task/research.go`, `cmd/serve_mcp.go:449-481`: BIT-49 V1 moves feedback to a top-level `feedback/` store that needs the data root separately (BIT-49 topic `review-2026-10-02` §4) and adds `retro_write`; BIT-46 V2 rewrites research/feedback as JSON+md and replaces `bitdir.ForRoot` with a per-call resolver (and likely `task.New(root, code)`, BIT-46 review §2). Anchor by function names (`AddNote`, `WriteResearch`, BIT-49's retro write), not paths.
- **S3. All do/bot-dev line refs** (`do:3,17,73,75,77,79,89-100,106,122,129`; `bot-dev:3,12-28,30-38`): confirmed today, shift after BIT-49 V3 deletes do `:75` last sentence, do `:99` staging sentence and bot-dev `:27-28`.
- **S4. `cmd/serve_mcp.go:151/160/169/178`, handlers `:378-481`, `task/store.go:260-310`**: confirmed today; BIT-46 rewrites store + handlers.
- **S5. `capture` cites `.bit/tasks/BIT-49.md:33-34`** for the git helper: now `:40` (decision) and `:52` (V2 Touches). **`.bit/tasks/BIT-46.md:59-63`** for the empty fields: now `:63-67`.
- **S6. `capture` lists `daemon/daemon.go` among exec users**: true today (`daemon/daemon.go:22`), gone after BIT-45.
- **S7. Verse 3 Touches "the do description"**: duplicated with Verse 1 (`:3`); give it to Verse 1 only (already noted in review).
- **S8. track Decision "`task_update` and `task_complete` accept `commit` and `branch`"**: nothing in BIT-47 calls `task_complete` with them. BIT-50 is the only consumer (BIT-50 topic `fields-and-writes`: repoint bars via `task_update` before `task_complete`, track commit via `task_complete{id,commit,branch}`). YAGNI says move it to BIT-50.

### CONFIRMED (highlights; each re-read)
- No Go git exec: `exec.Command*` only in `claude/{sync.go:13,dispatch.go:17,plugin.go:78}`, `daemon/daemon.go:22`; daemon calls `launchctl` only. No `"git"` string in Go.
- `Store.Update` `task/store.go:272-310`; `contentChanged` `:298` = title/body/phase/phase_label sent (on field *sent*, not value change, `serve_mcp.go:57-60`); `sentBack` `:299`; revoke `:301-303`.
- `Patch` `:260-266`; `taskUpdateSchema` `:277-286`; handler `:378-399`; `taskCompleteInput` `:160-162`, handler `:417-431` → `Complete` `:108-110` → `relocateTree` `:112-147`; `feedbackAddInput` `:169-172`, handler `:449-464` (`AddNote` `:457`); `researchWriteInput` `:178-182`, handler `:466-481` (`WriteResearch` `:474`).
- `AddNote` `task/feedback.go:58-79` create-only raw body; `WriteResearch` `task/research.go:37-57` overwrite, raw body, no metadata.
- `bitdir.ForRoot` `bitdir/bitdir.go:49-55`, `worktreeCut` `:57-68` (cuts worktree back to main checkout `.bit`).
- CLI: `cmd/task/update.go:9-55` (flags `:23-41`, store `:19`); `cmd/task/complete.go:15`; `cmd/feedback_add.go:19,29`; no research CLI.
- do `SKILL.md`: `:3,:17,:73,:75,:77,:79,:91,:92-98,:99,:100,:104-108,:122,:129` say exactly what `commit-skill` quotes.
- bot-dev `:3` ("without an operator sitting in front of it", `--bg` example), `:12` "not watching", `:16-22`, `:20` "and then commit", `:22`, `:26-28`, `:30-38` push, `:44-46`.
- plan `:304` "**Claude never commits.**", `:361` and `:398` `## Commit (user)`. (`:131` "Commit." in the contradiction recipe is generic, no change needed.)
- check `:34` fuzzy log match, `:38` marks stale bars done; complete `:30` "The user runs the commit; you don't."; retro `:57` example only; bot.md / ruler.md: no commit text.
- Skills auto-discovered (`bit/.claude-plugin/plugin.json` lists none; `.claude-plugin/marketplace.json` → `./bit`); `name: bit_<x>` convention in all 9 skills; evals only analyze/complete/feedback; complete evals shape `{skill_name, evals[{id,prompt,expected_output,files}]}`, eval 1 "commit suggested, not run". `SyncPlugin` `claude/sync.go:33-45`; `just release` → `scripts/release.sh` writes plugin.json.
- `.pre-commit-config.yaml` runs trailing-whitespace, check-yaml, `just fmt`, `just lint`, `just test`; **the hook is installed** (`.git/hooks/pre-commit`).
- History: `main` = 320 commits, 304 `josiah <lowgen0@…>`, 16 GitHub-committed squashes, 0 merges; `54abfeb` body lists `* subject` bullets; BIT-44 `26f0435..8decf35` and `af07620`/`a7d9ea7` direct on main; `.bit/completed/BIT-43.1.md:65` `## Commit (user)`. `main` == `v2` == `6a1d345` today.
- `v2-sketch.md` no longer says "squash-merged, so branch commits don't survive" (soundness-2 "fixed since" holds).

## 2. Build map: what changes when the write surfaces gain git fields

### `taskUpdateInput` + `Store.Update` (Verse 1)
- Struct `cmd/serve_mcp.go:151-158`: add `Commit *string json:"commit,omitempty"`, `Branch *string json:"branch,omitempty"`. Schema `:277-286` picks them up automatically.
- Handler `:386-392`: map into `task.Patch`.
- `Patch` `task/store.go:260-266` + `Update` `:272-310`: apply; keep out of `contentChanged` `:298`. Update the doc comment `:268-271`.
- Description `:54-67`: say commit/branch don't revoke. Test `cmd/serve_mcp_test.go:40-42,202-206` asserts substrings; keep `"title, body, phase or phase_label revokes it"` intact, add a constant for the new sentence.
- Other `Update` callers: only `cmd/task/update.go:43` (CLI, Patch is additive, no change; no flags per decision).
- Tests to extend: `task/store_test.go:863-943` `TestStoreUpdate_AppliesOnlySetFields` (add a commit+branch case; uses `reflect.DeepEqual` on `Task`, so it needs BIT-46's `Commit`/`Branch` fields on `Task`), `:945-995` `TestStoreUpdate_ApprovalRevocation` (add "commit/branch keep approval", "commit+done keeps approval"); `cmd/serve_mcp_write_test.go:95-126` (revocation), `:128-192` `LeavesOmittedFieldsAlone` (add "commit only leaves the rest alone"; DeepEqual on `task.Task`). CLI approval tests `cmd/task/update_test.go:115-220` unaffected.
- **Gap: nothing reads the fields back.** `taskReadOutput` `:213-222`, `taskSummary` `:124-132` and `taskReadDescription` `:36` carry no commit/branch, and BIT-46 V2 says MCP "behaves as before". Verse 1's observable ("bar ends up with commit and branch filled in") has no MCP read path. Default: add `commit`,`branch` to `task_read` output in the Verse 1 Go bar.

### `taskCompleteInput` (if kept here; recommend moving to BIT-50)
- Struct `:160-162`, handler `:417-431` must `Update` before `Complete` (Complete moves files; Load reads `tasks/` only, `store.go:37-39,164`). Description `:76-82`; tests `cmd/serve_mcp_write_test.go:556-625`, `cmd/serve_mcp_test.go:201`; CLI `cmd/task/complete_test.go` unaffected.

### `researchWriteInput` / `feedbackAddInput` / retro (Verse 4)
- No input field change (capture is automatic). The change is a git fact flowing from entry point to store write:
  - MCP: `runMCPServer(ctx, root, transport)` `:238` and `newServeMCPCmd` `:231-233` need a git reader (BIT-49 helper) injected; harness `cmd/mcp_harness_test.go:13-21` calls `runMCPServer` directly, so the harness signature changes (one place) unless BIT-49 already threads the helper through. Handlers `:457`, `:474`, BIT-49's `retro_write` pass `{sha, branch, at}` to the store.
  - CLI: `cmd/feedback_add.go:19` reads from `os.Getwd()` (decision).
  - Store: `AddNote(track, body)` `task/feedback.go:58`, `WriteResearch(track, topic, body)` `task/research.go:37` (post-46/49 versions) gain a git-fact param or options struct. **No test calls `AddNote`/`WriteResearch` directly** (grep), so signature churn is confined to `cmd/serve_mcp.go`, `cmd/feedback_add.go`, BIT-49's retro path.
  - Append rule (review default): compare to last entry's sha; append if different or list empty; empty sha never creates an entry. Needs BIT-46's research rewrite to read-modify-write the JSON, not replace it (hazard noted in review §4).
  - Tests: `cmd/serve_mcp_research_test.go:35-168,292-372` (write cases), `cmd/serve_mcp_write_test.go:474-555` (feedback_add), `cmd/feedback_add_test.go:23-215` (8 CLI tests, already rewritten by BIT-46/49 for paths), plus new store-level tests for the append rule (none exist today for research/feedback in `task/`).

### Approval-revocation logic and its tests (complete list)
`task/store.go:268-303`; MCP description `cmd/serve_mcp.go:57-63`; tests `task/store_test.go:945-995`, `cmd/serve_mcp_write_test.go:95-126`, `cmd/serve_mcp_test.go:40-42,202-206`, `cmd/task/update_test.go:115-220`; skill text do `:37` and `:75` ("comes back `true`"). `SetApproved` `:312-321` (approve CLI/TUI) unaffected.

## 3. Skill/agent lines BIT-47 rewrites, and who else edits them

| Line (today) | BIT-47 change | BIT-49 V3 sweep | BIT-50 |
|---|---|---|---|
| do `:3` description | "hands off ... for verification and commit" | `.bit/` mentions (twice) | sign-off sentence ("user's sign-off flips the track to done and files it") becomes false; not in BIT-50 Touches — **gap** |
| do `:17` | "so is the commit" | — | — |
| do `:73` | User-verifies path: run bit:commit after confirm | — | — |
| do `:75` | optimistic path rewrite, unwind wording, "user reads the diff when they commit" | deletes last sentence | — |
| do `:77`, `:79` | "do not commit" / message-at-end | — | — |
| do `:91-100` Verified good | reorder: bit:commit → task_update(commit,branch,done) → rollup → compaction | `:95` `.bit/completed/`, `:99` staging sentence | `:95` (track status), `:102-110` |
| do `:106` | "commit suggested" → "committed" | `:104`, `:108` `.bit/completed/` | owns `:102-110` |
| do `:122` unwind | commit can't be unwound | — | — |
| do `:129` | "Commit — always the user's action" | — | — |
| do `:131-132` | — | `.bit/` | `:131` |
| bot-dev `:3,:8,:12,:16-22` | rewrite (asks every commit, no `--bg` unattended) | — | — |
| bot-dev `:26-28` | message source → bit:commit | deletes `:27`, half `:28` | — |
| bot-dev `:30-38` push | keep, gate on permitted commit | — | — |
| plan `:304`, `:361`, `:398` | rewrite / `## Commit` | plan `:3,8,10` only (no overlap) | — |
| bot.md `:41` | — | `.bit` lines `:8,16,22` | sign-off routing |

## 4. Bar order (green build/tests and consistent skills at each bar)

Pre-req: `go build ./...` fails on a fresh tree until `sqlc generate` (`db/orm/` gitignored, `.gitignore:3`); verify with `just test` / `just lint`, never `just install`.

1. **V1 bar 1 (Go):** `task_update` takes `commit`/`branch`, kept out of revocation; `task_read` returns them; description updated. RED: MCP test "commit+branch+done on an approved bar keeps approval and reads back". Additive, so build/tests stay green; no skill changes.
2. **V1 bar 2 (skill):** new `bit/skills/commit/SKILL.md` (`name: bit_commit`) + `evals/evals.json`: prose ask with diff stat + message (from `## Commit` or `## Commit (user)`), one `git add … && git commit`, `git rev-parse HEAD` / `--abbrev-ref HEAD`, one `task_update{commit,branch,status:done}`; decline/nothing-to-commit/hook-failure/no-git/detached paths. Standalone-invocable; do still says the user commits (tolerable: nothing calls it yet). quick_validate + `claude plugin validate ./bit`.
3. **V1 bar 3 + V2 together (skills):** do close-out reorder and lines above **and** bot-dev rewrite in the same bar. Reason: once do commits via bit:commit, bot-dev `:18,:20` ("bit:do leaves the commit to the user … then commit") is false and would try a second commit. If bars must stay one-verse, merge Verse 2 into Verse 1 at scope time (bot-dev is a thin wrapper on do), or accept a one-bar window with V2 immediately next.
4. **V3 (skills):** plan `:304`, `:361`, `:398`. Contradiction window ("Claude never commits") exists from bar 3 until here; low risk, or fold into bar 3.
5. **V4 bars (Go), independent; can run before 1:** (a) git seam into `runMCPServer`/harness + research append rule; (b) feedback_add MCP + CLI capture; (c) retro_write capture. Each additive with stubbed helper.
6. `task_complete` commit/branch: drop to BIT-50 (S8).

## 5. Would stall planning or implementation
- **Headless bot-dev** (`claude --bg`, bot-dev `:3`) cannot answer an `ask` prompt (W1). Settle before V2: drop the `--bg` claim, or spike how `--bg` surfaces prompts.
- **`CLAUDE_PROJECT_DIR` in a Claude worktree session** (worktree or main checkout?) is unverified (BIT-46 review §3). If it's the main checkout, MCP-side capture records the wrong HEAD/branch for worktree sessions. bit:commit itself is safe (it runs git in the session cwd). Spike in BIT-49's helper work or here.
- **No read path for commit/branch** (section 2) unless decided.
- **Pre-commit hook runs `just fmt/lint/test`** on every bit:commit in this repo: slow, and `just fmt` can modify files and fail the commit; bit:commit needs the "re-stage and ask again, never `--no-verify`" path (review §3 default).
- **Undecided edge defaults** in review §3 (nothing to commit, no git, detached HEAD, unwind after commit, "HEAD moved" rule) must become Decisions before bit:plan or it will guess.
- **Store signatures after BIT-46/49 are unknown** (task.New(root, code)? separate feedback store type?). Plan V4 bars only after BIT-49 lands; re-anchor every line ref then.
