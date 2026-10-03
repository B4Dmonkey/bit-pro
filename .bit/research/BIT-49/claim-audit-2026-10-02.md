# Claim audit + build depth, 2026-10-02 (v2 @ 6a1d345, read-only)

Scope: every factual claim in the BIT-49 body, topics `decisions`, `index`, `review-2026-10-02`, and BIT-46 `file-inventory`'s BIT-49 sections. Baseline: `go build ./... && go test ./...` green at HEAD. Client project not opened.

## 1. Verdicts

Counts: **58 CONFIRMED, 6 WRONG, 7 STALE** (line drift / count drift). Operator decisions and Claude defaults are not facts and were not graded.

### WRONG
1. **review §3: "bit-pro moved [flat `archive/<ID>.md`] to `archive/tasks/` by hand (95aec1f)".** 95aec1f ("file completed BIT-18 and migrate the flat archive") renamed all **126 flat-archive files to `completed/`** (`git show --name-status -M 95aec1f`: 126 × `R100 .bit/archive/X .bit/completed/X`). The flat archive was used by hand to file *finished* work ("chore(bit): archive completed BIT-11 and BIT-12", 1766d38, 2026-07-23) as well as by `delete` (6f01ee5, 2026-07-22 → 305d09f, 2026-07-29). So the review's default "treat `archive/<ID>.md` as `archive/tasks/<ID>.md`" contradicts bit-pro's own precedent. **Corrected:** a flat-archive file's meaning is ambiguous (completed or deleted). Both dirs reserve IDs, so either mapping is safe for ID reservation; it changes only whether the track shows as completed or archived. Needs a decision (suggest: same as bit-pro precedent → `completed/`, or ask per run).
2. **review §3: "6 files in `completed/` are todo/doing (BIT-18.2, BIT-2.4, BIT-25, BIT-31.3, BIT-36, BIT-43)".** BIT-2.4 is `done`; the count came from a `grep ^status:` that matched body text. Frontmatter-only count: **5** (doing: BIT-31.3, BIT-36, BIT-43; todo: BIT-18.2, BIT-25). The point (verify must not assume completed ⇒ done) stands.
3. **review §1 / feedback_add description (`cmd/serve_mcp.go:97`): "AddNote … create-only".** `AddNote` writes with `os.WriteFile` (`task/feedback.go:74`), which truncates. Uniqueness rests only on `nextNoteSeq` (glob then write, no `O_EXCL`). Harmless per-repo today; in a **shared** top-level `feedback/` two sessions (two worktrees of one project) can race to the same seq and one note overwrites the other. Verse 1 should write with `os.OpenFile(O_CREATE|O_EXCL)` and retry on `EEXIST`.
4. **track Why: "retro and learn read and write `.bit/feedback` and `.bit/retro` directly".** learn never touches `.bit/`: it asks for a path or pasted contents (`learn/SKILL.md:16`). Only retro does file I/O (`retro/SKILL.md:18,29,76,110`, and :27 says it writes "a plain file … directly").
5. **review §1: retro `:122` ("Move the proposals file anywhere …").** That text is `:123`. `:122` is "Edit a skill … that's bit_learn, and it runs in bit-pro itself, not here" (still true in v2).
6. **review §2 "Never hand-edit" list: "Seven places … do :15/:132, plan :10, scope :20, complete :12, bot :22, ruler :12".** That is 8 lines and misses **`check/SKILL.md:12`** ("Never hand-edit `.bit/tasks/*.md`"). Full set: do :15, :132; plan :10; scope :20; complete :12; check :12; bot :22; ruler :12; plus feedback :12, :113 (hand-write `.bit/feedback`).

### STALE (drifted line refs / counts)
- review: `AddNote` `:57-79` → **`:58-79`**.
- review: `task.Parse` uppercases at `task/task.go:47-51` → **`:48-51`**.
- review: `bitdir.worktreeCut` `bitdir/bitdir.go:61-72` → **`:57-68`**.
- review: `.bit/` "423 total, 371 tracked, 52 untracked" → now **427 / 371 / 56** (research grew; this topic adds one more).
- review: research "48 files across BIT-44..50" → **57**.
- `README.md:82` "seven skills" vs nine — confirmed, but the README also omits the agents.
- BIT-46 `file-inventory` "Regenerate it with" grep: misses by-meaning lines (bot.md:3 has `.bit/` and *is* caught; but README :3/:13-16/:121-135 and retro :10/:27/:102/:123, learn :8/:16/:22 aren't `.bit` strings).

### CONFIRMED (abridged; all checked at file:line)
- `retro/SKILL.md:76` proposals path; `learn/SKILL.md:3` "carried over by hand", "never inside the project"; `cmd/task/complete.go:12` Short text.
- `task/feedback.go`: `feedbackDir` :15-17 off `s.root`; `notePath` :19-21 via `pathologize.Join`, `%s-%03d.md`; `nextNoteSeq` :23-33; `trackExists` :35-43 (tasks/completed/archive); `resolveTrack` :45-56 (rejects `..`, `/`, `\`; accepts a bar ID because it stats `Path(id)`).
- `feedback_add`: input `{track, body}` :169-172, output `{path}` :174-176, registration :256, handler :449-463 (`task.New(bitdir.ForRoot(root))`). CLI `bp feedback add <track> -d` prints path (`cmd/feedback_add.go:11-31`).
- `cmd/serve_mcp.go` `.bit/` descriptions :76, :84, :102, :109; `CLAUDE_PROJECT_DIR` read once at :231.
- `update/normalize.sh` and its five carriers (`update/README.md` "What it rewrites"); it ignores flat `archive/*.md` and `research/`.
- `analyze/evals/evals.json:14` is the only eval naming `.bit/`; complete and feedback evals have none.
- BIT-46 reserves `FEEDBACK`/`RETRO`, code `^[A-Z][A-Z0-9]*$`.
- The do/bot-dev staging lines: do :75 (last sentence), :99; bot-dev :27-28.
- `hierarchy.md:15,23`; README :7,34,42,59,65,66,67,82,103,110,137,141.
- `cmd/init.go:23`, `cmd/add.go:51,66` name `.bit` (owned by BIT-46/48).
- `highestReserved` scans tasks/completed/archive/tasks (`task/store.go:637-650`).
- BIT-45 topic `migrate` still says `bit.db`, lowercase dirs, git-log timestamps, `migrated: true`, copy unknown files — superseded, as the body says.
- `.bit/` has no `retro/`, no dotfiles, no empty dirs, only `config.toml` (`prefix = "BIT"`) as non-`.md`; 27 feedback names all match `^BIT-\d+-\d{3}\.md$`; `order` only on completed/BIT-10, BIT-39; `approved` always `true` (115); id == filename stem on all 345.
- No git helper exists today (`grep exec.Command`: only `claude/` and `daemon/`).
- BIT-46 file-inventory "Later edits": BIT-47 do/plan/bot-dev/commit; BIT-50 complete+evals, do sign-off, `bot.md:41` — all lines exist as cited (plan :304, :361, :398; complete :30).

## 2. Verse 1 — feedback/retro: exact build surface

### Go call sites
- `task/feedback.go:58` `AddNote`; called from `cmd/feedback_add.go:19` and `cmd/serve_mcp.go:457`. Nothing else.
- `resolveTrack` is shared with research (`task/research.go:38,60,78`) — moving feedback off `s.root` must keep `resolveTrack`/`trackExists` on the project store. The feedback dir needs the data root (new field or separate type).
- MCP registration list `cmd/serve_mcp.go:238-275`; tool-name consts :16-29; descriptions :32-115; input/output types :145-222. New tools add consts, descriptions (written without `.bit/`), types, handlers, and `mcp.AddTool` lines here.
- No `task/feedback_test.go` exists. Tests that change:
  - `cmd/feedback_add_test.go` (8 tests, :23-215; `.bit/feedback/...` asserted at :11,33,50,54,63,80,84,93,122,144,161,189,207,211). BIT-46 rewrites these first; Verse 1 rewrites them again for top-level `feedback/`.
  - `cmd/serve_mcp_write_test.go:474-555` (`FeedbackAddWritesANote`, `…RefusesAnUnknownTrack`, `…RefusesAPathLikeTrackID`; `.bit` globbed at :523, :546; `testFeedbackDir` const :37).
  - `cmd/serve_mcp_test.go:182` `TestMCPToolDescriptions_CarryTheDomain` — add rows for the new tools.
- `claude.Runner` (`claude/sync.go:10`) returns only `error`; it can't carry git output, so the git helper (Verse 2) needs its own `func(ctx, dir, args...) (string, error)`-style seam.

### Skill lines that must change
- **retro/SKILL.md:** :3 (description: `.bit/feedback/*.md`; "runs somewhere else entirely"), :10 ("leaves this project … runs somewhere else"; also points at a *Handoff* section that **doesn't exist**), :18, :27 ("plain file you write directly … the surface carries no retro tool"), :29 ("no tool of their own … List `.bit/feedback/*.md` directly"), :76 (path + naming), :102 ("the handle they carry elsewhere"), :110 (glob `.bit/retro/*-proposals.md` → `retro_list`), :123 (move file manually).
- **learn/SKILL.md:** :3 (description), :8 ("runs elsewhere … hands you a file"), :16 ("Ask for the proposals file … path or contents" → `retro_list`/`retro_read`), :22 ("before they're carried over").
- **feedback/SKILL.md:** :3 ("into `.bit/feedback/`"), :12, :113. Tool list on :12 is unchanged (feedback only adds).
- **Outside Touches:** `bit/agents/bot.md:22` enumerates "the whole write surface" (`task_create`, `task_update`, `task_move`, `task_complete`, `feedback_add`) — `retro_write` joins it (and it already omits `research_write`, `task_delete`). `bot.md:57` routes "handing over a retro proposals file" to learn.

## 3. Verse 2 — migrate inputs

### This repo's `.bit/` (inventory)
- Dirs: `tasks/` 16, `completed/` 321, `archive/tasks/` 8, `feedback/` 27, `research/BIT-44..50/` 57, `config.toml`. Statuses: tasks 8 done/1 doing/7 todo; completed 316 done/3 doing/2 todo; archive 1 done/7 todo.
- Frontmatter keys (345 task files): id/title/status 345, phase+phase_label 293, approved 115, order 2. Feedback and research files have **no frontmatter**.
- **Split trees (new):** 5 bars sit in `archive/tasks/` while their track is in `completed/`: BIT-23.2, 23.3 (track BIT-23), BIT-39.13, 39.14 (BIT-39), BIT-41.6 (BIT-41). Verify must not assume a track and its bars share a dir.
- **Partial `order` (new):** `completed/BIT-10` orders 8 bars but 9 exist (BIT-10.9 absent); BIT-39's order omits its two archived bars. Carry `order` verbatim; don't "repair".
- **Feedback on an archived track (new):** `BIT-19-*.md` keys to `archive/tasks/BIT-19.md`. `trackExists` accepts it; migrate must too.
- **Round-trip (new):** a scratch program parsed all 345 task files with `task.Parse` (0 errors) and `t.Bytes()` reproduced each file **byte-for-byte**. So "Parse → Bytes == source" is a cheap verify that also catches unknown frontmatter keys (yaml.v3 non-strict `Unmarshal` drops them silently).

### What `task.Parse` accepts (`task/task.go:30-56`)
- Needs a leading `---\n` and a `\n---\n` closer; anything else errors (CRLF, BOM, a closer at EOF with no newline).
- Non-strict YAML: unknown keys dropped; no validation of status, empty id, or id/filename agreement.
- Uppercases `id` and `order` (:48-51); `Config()` uppercases the prefix (`task/config.go:25`). The uppercase refusal must read raw bytes.

### v1 layout variants in git history (`git log --all -- .bit`)
- daeed54 (2026-07-17): first import, `tasks/` + `config.toml`; keys id/title/status/phase/phase_label from day one.
- 2ea0768 (2026-07-22): `order` lists.
- 6f01ee5 (2026-07-22) → 305d09f (2026-07-29): **flat `archive/<ID>.md`** (delete + hand-filed completed work); 126 files ever.
- 57b9ce0 (2026-07-29): `completed/`; 305d09f: `archive/tasks/`; 95aec1f: flat archive → `completed/`.
- 6a89aa3 (2026-08-05): `feedback/`. f0dcf1b (2026-08-20): uppercase IDs; BIT-21 normalize (PR #1, 2026-08-14).
- 01b4853 (2026-08-14): `approved`. 26f0435 (2026-09-23): `research/`.
- Never seen: lowercase `id:` lines, bar-level feedback names, non-`.md` files besides `config.toml`, `retro/`.
- What other projects can still hold: flat `archive/*.md` (normalize ignores it), mixed-case research dirs (normalize ignores `research/`), a missing `config.toml`, a prefix that doesn't match the IDs' prefix, a prefix with `-` (fails the code format).

### Other open points for migrate
- Worktree detection: v1 cuts only at `.claude/worktrees/` (`bitdir.go:57-68`); a plain `git worktree add` checkout with a tracked `.bit/` snapshot would be migrated and registered as itself. `git rev-parse --git-common-dir` covers both.
- That `CLAUDE_PROJECT_DIR` is the worktree path in a worktree session is inferred from v1's `ForRoot` cut, not verified here.

## 4. Verse 3 — full line list (non-Go and Go)
`.bit` / `bp init` strings (✱ = BIT-47 re-edits, † = BIT-50 re-edits):
- `bit/.claude-plugin/plugin.json:5` (also says "through the bp CLI").
- `bit/agents/bot.md:3` (**missed by review**: "any project with a `.bit/` directory"), :8, :16, :22; `bot-dev.md:27-28` ✱; `ruler.md:10, :12`.
- `analyze/SKILL.md:3, 17, 64`; `analyze/evals/evals.json:14`.
- `check/SKILL.md:12`.
- `complete/SKILL.md:3, 8, 12, 23` †, **:30** † (**missed**: "out of `.bit/tasks/`").
- `do/SKILL.md:3` (×2) ✱, :8, :15, :43, :75 ✱, :95 †, :99 ✱, :104 †, :108 †, :131 †, :132.
- `feedback/SKILL.md:3, 12, 113`; `learn/SKILL.md:3`; `plan/SKILL.md:3, 8, 10`; `retro/SKILL.md:3, 18, 29, 76, 110`; `scope/SKILL.md:3, 10, 20 (×2), 110, 205, 227`.
- `README.md:7, 34, 42, 59, 65, 66, 67, 82, 103, 110, 111, 116, 137, 141`; by meaning also :3, :13-16 ("lives next to the code", "markdown files in git"), :121-135 (Storage: "No index — `bp task list` globs", frontmatter example). BIT-48 also edits README (setup commands).
- `hierarchy.md:15, 23`.
- Go: `cmd/serve_mcp.go:76` †, :84, :102, :109; `cmd/task/complete.go:12`. `cmd/init.go:23`, `scripts/install.sh:22`, `cmd/add.go:51,66` are BIT-46/48's.

## 5. Bar order (each bar green)
Pre-req: BIT-46 (resolver, data root, JSON+md records) and BIT-48 (ensure-wiring func) landed.
- V1.1 feedback store moves to top-level `feedback/` with `project`; `O_EXCL` write; update the two test files.
- V1.2 `feedback_list` + `feedback_read` (current project only) + description-test rows.
- V1.3 retro record store + `retro_write`/`retro_list`/`retro_read`.
- V1.4 retro/learn/feedback SKILL.md + bot.md:22,:57; validate.
- V2.1 git helper (stub seam; empty on error).
- V2.2 migrate source checks: locate `.bit/` (walk-up, worktree), raw uppercase, unknown/flat-archive files, Parse errors, code format/collision; refuse before writing.
- V2.3 stage → convert (git fields, timestamps) → verify (counts, round-trip, byte-equal bodies, split trees) → commit → register last; "already migrated".
- V2.4 first-registration wiring + printed cleanup.
- V3.1 Go descriptions + `complete.go:12` (tests check domain sentences, not paths).
- V3.2 skills/agents/plugin/evals sweep except the staging lines; validate.
- V3.3 staging lines (do :75, :99; bot-dev :27-28); validate.
- V3.4 README + hierarchy.md.

## 6. Stalls
- Bar-level planning of V1/V2 needs BIT-46's concrete API (data-root accessor, record type, registry calls) and BIT-48's ensure-wiring function — none exist yet.
- Flat-archive mapping (completed vs archive) needs an operator call (WRONG #1).
- Concurrent feedback writes in the shared dir (WRONG #3).
- `feedback_read` id shape, `retro_list` scope, retro naming for multi-track: still Claude defaults in review §4, not in the body.
