---
id: BIT-47
title: 'v2: commit skill and git capture'
status: doing
approved: true
---
## Why
In v1, Claude marks a bar `done` and then hands the commit to the operator. So bit never learns which commit a bar became, and records carry no git facts at all (`bp never calls git`, BIT-45 topic `history-anchors`). v2 wants traceability: the state of the world when a record was written, and the commits a track's work became. BIT-46 adds empty git fields to every record. This track fills them as work happens. Claude makes each commit through one focused skill, always with the operator's permission, and records the hash, and research, feedback and retro writes note the HEAD they were written at. BIT-50's merge-aware completion depends on these bar hashes.

## Summary
- A new `bit:commit` skill asks the operator's permission, commits, and records the commit's hash and branch on the bar.
- bit:do's close-out becomes: commit through `bit:commit`, record the hash, mark the bar `done`, roll up the track.
- bot-dev uses the same skill and stops for permission before every commit.
- `task_update` accepts a bar's `commit` and `branch`, and `task_read`/`task_list` return them. Research, feedback and retro writes record the HEAD they were written at, and append a `commits` entry when HEAD has moved.
- The skills stop saying "the user commits".

## Visual aid
```
v1 close-out:  bar done ─► roll up ─► suggest message ─► operator commits (bit never sees the sha)
v2 close-out:  bit:commit ─(permission prompt)─► git commit ─► sha+branch ─► task_update(bar: commit, branch, done) ─► roll up
research/feedback/retro write ─► HEAD+branch from session dir ─► append to commits if HEAD moved
```

## Decisions
- **Depends on BIT-46** (the empty git fields), **BIT-49** (the git helper and the `.bit/` staging sweep, which edits the same do and bot-dev lines first) and **BIT-48**, by order. BIT-50 comes after and depends on this track.
- **Split from the old BIT-47 (operator, 2026-10-01).** This track keeps the commit side. Merge-aware completion is BIT-50.
- **Cutover belongs to the operator.** The operator alone decides when v2 is ready and merges the branch. This track doesn't gate cutover.
- **Claude makes the commits, through a dedicated, focused commit skill (operator, 2026-10-01).** The skill commits, then records the hash. do and bot-dev use it. do's "the user commits — you never run the commit yourself" text (`bit/skills/do/SKILL.md:99`, also `:17`, `:129`) is retired.
- **The commit skill always asks the operator's permission before committing, bot-dev included (operator, 2026-10-01).** The operator's Claude Code permissions put the commit tool on the `ask` list, so every commit prompts (corrected 2026-10-02: the rules are `ask`, not `reject`. `~/.claude/settings.json:15-21` lists `git commit`, `git add` and `git push` there. The effect the decision relies on, a prompt on every commit, holds). Dispatching bot-dev isn't permission. bot-dev stops and asks before every commit, and its "without an operator sitting in front of it" and "not watching" text (`bit/agents/bot-dev.md:3,12`) is rewritten.
- **One bar can mean up to three prompts, for `git add`, `git commit` and, in bot-dev, `git push`. That's accepted (Claude default, 2026-10-02).** Each is on the `ask` list, and the decision is to ask.
- **In a headless run (`-p` or `--bg`), nobody can answer a prompt, so bot-dev asks in its own words and ends its turn (Claude default, 2026-10-02).** The bar stays `doing` and nothing is staged. The report lists the files and the commit message, and the operator answers by resuming the session. bot-dev's description, which names a `--bg` dispatch, is reworded to match.
- **A bar is committed first, and its hash is stored afterwards (operator).** The bar's `commit` holds its own commit, as far as it can. Bar hashes are best-effort and may break.
- **A bar whose commit is declined stays `doing`, and nothing is recorded (Claude default).** The skill says the commit was declined and stops. Done means committed, so a declined commit can't leave the bar `done`.
- **The skill also asks in its own prose before committing (Claude default).** The ask then exists even on a machine where the permission settings would allow the commit. When it asks, it shows the files it will stage and the message. That replaces do's "the user still reads the diff when they commit" (`do/SKILL.md:75`) (Claude default, 2026-10-02).
- **Edge cases the skill handles without stopping to re-scope (Claude default, 2026-10-02):**
  - Nothing to commit: say so, and on the operator's OK mark the bar `done` with empty git fields. Never `--allow-empty`. This is likely in v2, where `.bit/` changes no longer land in the repo.
  - A pre-commit hook fails or rewrites files (this repo's hook runs `just fmt`, `just lint` and `just test`, `.pre-commit-config.yaml`): show the output, re-stage, and ask again. Never `--no-verify`. The bar stays `doing` until a commit succeeds.
  - No git, or a bar whose work is outside the repo: skip the commit, leave the fields empty, and mark `done` after the operator confirms.
  - Detached HEAD: commit, and leave `branch` empty.
  - A fix after the bar was committed (an unwind or a follow-up): a new commit through `bit:commit`, which overwrites the bar's `commit`. It reuses the bar's subject, so BIT-50's squash matching still finds it.
- **The commit skill only commits. It doesn't push (Claude default).** bot-dev keeps its existing push step after a permitted commit, when the repo has a remote (`bit/agents/bot-dev.md:30-38`). bit:do doesn't push, the same as today. "Landed" means pushed, and BIT-50 tells the operator to push when commits are only local.
- **The commit message comes from the bar's commit section (Claude default).** bit:plan's template heading `## Commit (user)` becomes `## Commit` (`bit/skills/plan/SKILL.md:361,398`). The skill reads either heading, because existing bars still say `(user)`. bit:plan's "Claude never commits" (`:304`) is rewritten.
- **The MCP write tools take git info (operator, 2026-10-01).**
  - `task_update` accepts `commit` and `branch`. `task_complete`'s `commit` and `branch` moved to BIT-50, its only caller (Claude default, 2026-10-02; the operator decision named both tools, and BIT-50 still delivers the `task_complete` half).
  - `research_write`, `feedback_add` and BIT-49's `retro_write` record `commits` entries of `{sha, branch, at}`. The first entry is HEAD when the record is created, and a later write appends one when HEAD has moved.
  - For tracks and bars, HEAD isn't captured on every write. Only the commit skill records a bar's commit, and completion records a track's landing commit (BIT-50), not HEAD (corrected 2026-10-02).
- **`task_read` and `task_list` return `commit` and `branch` (Claude default, 2026-10-02).** Nothing returns them today (`taskReadOutput`, `cmd/serve_mcp.go:213-222`), so Verse 1's "the bar has its commit" can't be seen through MCP otherwise, and BIT-50's completion needs to read them.
- **"HEAD has moved" means HEAD's sha differs from the last entry's sha (Claude default, 2026-10-02).** An empty sha (no git, or a git error) never adds an entry. A branch change at the same sha doesn't add one.
- **Writing `commit` or `branch` doesn't revoke a bar's approval (Claude default).** It records work that was already approved, like a forward status move. Today any content change revokes approval (`task/store.go:298-303`), so these fields stay out of that check (topic `mcp-git-inputs`).
- **Git facts are read where the request comes in, from the session dir, and passed to the store (Claude default, from BIT-49's git-helper decision).** The store only sees the project root, so the worktree is gone by then (topic `capture`). `runMCPServer` gets BIT-49's git helper passed in, so tests stub it.
- **`CLAUDE_PROJECT_DIR` is the session's launch folder, which is the worktree for a worktree session (fact, from `mcp-notes.md:256-260`, `:429-431`).** So research and feedback writes record the worktree's HEAD. A session that moves into a worktree after it starts records its launch folder's HEAD, which is accepted. Bar hashes don't depend on this: `bit:commit` reads them from git in the folder it commits in.
- **The CLI's `bp feedback add` captures the same way, from the current folder (Claude default, 2026-10-01).** A note's `commits` then doesn't depend on which entry point wrote it, and it matches BIT-49's "the current folder for the CLI".
- **The CLI doesn't gain `--commit`/`--branch` flags (Claude default).** Only the skills write these fields, and they go through MCP.
- **check and complete are left alone here (Claude default).** check's message matching (`bit/skills/check/SKILL.md:34,38`) still works when a hash is missing, and the complete skill's commit suggestion (`bit/skills/complete/SKILL.md:30`) is rewritten with completion in BIT-50.
- **The commit skill gets evals in the same shape as `bit/skills/complete/evals/evals.json` (Claude default).** It's a new skill, and the pipeline's other user-facing skills have them.
- **Every bar that edits a skill or agent validates it (Claude default, 2026-10-01).** It runs skill-creator's `quick_validate.py` on each changed skill and `claude plugin validate ./bit` (memory `bit-skill-edits-run-skill-creator`; the kebab-case `name:` failure is known noise).
- **Each bar leaves `just lint` and `just test` green, and the skills consistent (Claude default, 2026-10-02).** That's what the pre-commit hook runs.
- **Line numbers here are as of `6a1d345`.** BIT-46 rewrites `cmd/serve_mcp.go` and `Store.Update`, and BIT-49 rewrites do `:75`/`:99` and bot-dev `:27-28`, so planning anchors on section and function names (topic `review-2026-10-02` maps each one).
- **Dev runs follow BIT-46's rule (Claude default, 2026-10-01):** never `just install` on `v2`, sandbox `XDG_DATA_HOME` and `HOME`, and test skills through `--plugin-dir` in a scratch git repo.
- **Git fields may be empty.** `acme/`-style folders have no git. A git error never fails a write.
- **Verse shape (Claude default, revised 2026-10-02).** The old Verse 2 (bot-dev) is folded into Verse 1, so the old Verses 3 and 4 are now 2 and 3. Once do commits through `bit:commit`, bot-dev's "do leaves the commit to the user" is false, and bot-dev would try a second commit. So the two change in one bar, and that bar can serve only one verse (topic `claim-audit-2026-10-02`).

## Verses
- [ ] Verse 1 — A bar lands as a real commit with its hash recorded, in bit:do and in a dispatched bot-dev. In a bit:do session, the close-out runs `bit:commit`, the operator approves the prompt, and the bar ends up `done` with its `commit` and `branch` filled in. A dispatched bot-dev implements a bar, then stops and asks before committing through `bit:commit`, and pushes only after a permitted commit. A declined commit leaves the bar `doing`.
  Touches: new `bit/skills/commit/` (SKILL.md and evals); `bit/skills/do/SKILL.md` (the description, and the close-out and its User-verifies path; today `:3`, `:17`, `:73`, `:75-79`, `:89-100`, `:106`, `:122`, `:129`); `bit/agents/bot-dev.md` (`:3`, `:12-28`, `:30-38`); `cmd/serve_mcp.go` (`taskUpdateInput`, today `:151`, `taskReadOutput`, today `:213-222`, and their handlers); `task/store.go` needs no change here: BIT-46.14 already adds `Patch.Commit`/`Branch` and keeps them out of the revocation check. See topics `commit-skill`, `mcp-git-inputs` and `claim-audit-2026-10-02`.
- [ ] Verse 2 — The pipeline's skills describe Claude committing. bit:plan's bars carry a `## Commit` section, and nothing tells Claude that the user commits.
  Touches: `bit/skills/plan/SKILL.md` (`:304`, `:361`, `:398`). See topic `commit-skill`.
- [ ] Verse 3 — Research, feedback and retro records show the commits they were written at. A new record's `commits` starts with the HEAD it was written at, and a later write after a new commit appends one.
  Touches: `cmd/serve_mcp.go` (the inputs don't change; the write handlers, today around `:449-481`, BIT-49's `retro_write`, and `runMCPServer`), `cmd/feedback_add.go`, the callers of the store's research, feedback and retro writes, which already take a `head` (BIT-46.16, BIT-46.17, BIT-49.5), and a new `cmd/session_head.go`, and BIT-49's git helper. It doesn't depend on Verses 1–2 and can be built first. See topic `capture`.

## References
- `.bit/research/BIT-47/`: start at `index`. Topics `commit-skill`, `mcp-git-inputs`, `capture` and `verse-order` are this track's research (they use the old verse numbers). `decisions` holds the operator decisions. `soundness` and `soundness-2` are earlier passes, now mostly BIT-50's. `review-2026-10-02` and `claim-audit-2026-10-02` (2026-10-02) verify each claim and win where older topics differ.
- `.bit/research/BIT-45/`: topic `history-anchors` (read its superseded banner first).
- `v2-sketch.md` (repo root).