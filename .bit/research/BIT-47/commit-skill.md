# Commit skill: packaging, and what do / bot-dev / plan / check / complete say today

Checked against branch `v2`, 2026-10-01. All paths relative to repo root.

## How skills ship (Q2)
- Skills are auto-discovered: one dir per skill under `bit/skills/<name>/SKILL.md`. `bit/.claude-plugin/plugin.json:1-10` lists no skills, only name/description/version (1.3.0). The marketplace (`.claude-plugin/marketplace.json:8-13`) points plugin `bit` at `./bit`. So a new skill = `bit/skills/commit/SKILL.md`, exposed as `bit:commit`. No registration step.
- Shipping: `bp init` runs `claude plugin marketplace update bit-pro` then `claude plugin update|install bit@bit-pro --scope project` (`claude/sync.go:33-40`). The marketplace installs from GitHub, so an unpushed skill never reaches a target project (memory `bit-plugin-installs-from-github`). Version bump via `just release` (`Justfile`, writes plugin.json). On v2, dev sessions use `claude --plugin-dir <checkout>/bit` (BIT-46 decision), so the skill is testable unreleased.
- `plugin.json` description (line 5) says "execute them one commit at a time" — neutral, no change needed.
- Evals: only `bit/skills/{analyze,complete,feedback}/evals/evals.json` exist (skill-creator format: `skill_name`, `evals[] {id, prompt, expected_output, files}`). do, plan, check, scope, retro, learn have none. No in-repo runner; evals run through skill-creator. A new `commit` skill can follow the `complete` evals shape. Note `complete/evals/evals.json` eval 1 expects "commit suggested, not run" — would need revisiting if complete starts using the commit skill.

## What bit:do says today (Q1)
- Description `bit/skills/do/SKILL.md:3`: "...hands off to the user for verification and commit between bars."
- `:17`: "Verification is the user's call, and so is the commit."
- `:73` (User-verifies path): state suggested commit message, leave bar `doing`, stop.
- `:75` (no-User-verifies path): run Verified good inline — "mark the bar done, roll the track up, state the commit message". Ends: "marking it now keeps the `.bit/tasks/*.md` status change in the tree for the *same* commit as the code". This is the done-before-commit rationale.
- `:77`: "do **not** commit, and do **not** start the next bar."
- `:79`: every follow-up reply ends with the commit message.
- **Verified good close-out `:89-100`** — this is where done + rollup happen:
  - `:91` step 1 mark bar `done` (`task_update status=done`).
  - `:92-98` step 2 roll up: verse checkoff `:94`, track status `:95`, one `task_update` on the track `:96`.
  - `:99` step 3 "Suggest the commit... The user commits — you never run the commit yourself. The `.bit/tasks/*.md` changes from steps 1–2 ... go into the same commit".
  - `:100` step 4 compaction point ("bar is done, verified, and committed").
- `:122` unwind: set back to `doing`, reverse verse checkoff (relevant: with commit-first, unwind after a commit means the commit exists but bar isn't done).
- `:129` "Commit — always the user's action; you suggest the message."
- Track sign-off `:104-108` untouched by this track (BIT-50 territory).

## What bot-dev says today (Q1)
- Description `bit/agents/bot-dev.md:3`: commits "and pushes if the repo has a remote — but only on a bar that has no `## User verifies` items"; "without an operator sitting in front of it".
- `:12` "operator who is **not watching**"; `:16-22` the one delta: commit when no User verifies, after close-out "bar `done`, track rolled up — and then commit" (`:20`).
- `:26` use bar's suggested message; `:27` `.bit/tasks/*.md` status changes join the commit "that is why bit:do makes them before this point"; `:28` stage bar's files plus `.bit/`.
- `:30-38` push: `git remote` check, `git push -u origin HEAD`, no force, no PR.
- `:44-46` operator keeps approval, track sign-off, next bar.

## What must change for "commit first via commit skill (asks permission), then record hash, then done"
- do Verified good: reorder to (a) invoke `bit:commit` — which asks permission, commits, returns the sha; (b) `task_update` the bar with `commit`+`branch` and `status=done`; (c) rollup. Rewrite `:75` rationale (no more "same commit" argument), `:77`, `:79` (commit message may now be the skill's prompt), `:99`, `:100`, `:129`, description `:3`, `:17`.
- If the operator declines the commit: bar stays `doing` (nothing recorded). The skill text needs a decline path; today there is none.
- The `.bit/` staging lines (`do:75` tail, `do:99` tail, `bot-dev:27-28`) are BIT-49 Verse 3's sweep, not this track — but the reorder edits the same sentences, so the two tracks touch the same lines. Ordering: BIT-49 lands first (track order 49 → 47).
- bot-dev: rewrite description `:3` and `:12-22` ("not watching", commit-without-asking). It now calls `bit:commit` and stops for permission on every bar, User verifies or not. Push section `:30-38`: stub doesn't say whether the commit skill pushes, or whether bot-dev still pushes — **open** (BIT-50 treats "landed" as pushed).
- The commit skill uses `git commit` via Bash; the operator's permission settings reject it so the harness prompts every time (stub decision). The skill should also ask explicitly in prose so the ask exists even where settings allow.

## Other skills that assume the user commits (Q5)
- `bit/skills/plan/SKILL.md:304` "**Claude never commits.** ... committing is always the user's action." — direct contradiction.
- `plan/SKILL.md:361` and `:398` bar template headings `## Commit (user)` (TDD bar and spike bar). bit:do and bot-dev read the message from this heading; renaming it (e.g. `## Commit`) changes every existing bar body's convention; old bars still say `(user)`. Decide whether the skill matches either.
- `bit/skills/complete/SKILL.md:30` "Then suggest a commit... The user runs the commit; you don't." Contradiction in spirit; in v2 the store leaves the repo (BIT-46), so filing produces no working-tree renames and this commit suggestion may simply disappear.
- `bit/skills/check/SKILL.md:34` fuzzy-matches `git log --oneline` against each bar's suggested message; `:38` marks stale bars done. Not a contradiction, but with bar `commit` hashes check could verify by sha; and `:38` would mark done without a hash.
- `bit/skills/retro/SKILL.md:57` mentions "Every reply ... should end with the commit message" as an example proposal — example text only.
- `bit/agents/bot.md`, `ruler.md`: no commit mentions.
