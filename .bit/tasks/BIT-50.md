---
id: BIT-50
title: 'v2: merge-aware completion'
status: todo
---
## Why
Today a track is "completed" when the operator signs it off, before its work has necessarily been pushed or merged. Completion only moves files (`task/store.go:108-147`), never looks at git, and the complete skill forces every unfinished bar to `done` without asking (BIT-50 topic `completion-today`). So "completed" doesn't mean "on main", and nothing records where the work landed. v2 wants each completed track anchored to its landing commit, so the state of the world before and after a track can be worked out later. A track whose work never landed should be noticed, not quietly filed.

## Summary
Completion moves from sign-off to after the work has landed (pushed). `/bit:complete` asks git where the track's bar commits are: on trunk directly or through a merge, inside a squash, only on a branch or local, or untraceable. It then records the track's landing commit and files the track. If the track is partly done or not done, it checks in with the operator instead. When it can't tell, or the work isn't on trunk, it asks for the PR. bp never fetches, so every "not landed" answer first tells the operator to fetch, pull or push and try again.

## Visual aid
```
/bit:complete BIT-N  (after push/merge)
  bars' commit hashes ──► trunk = origin/main (else main)
     all on trunk ───────────────► landing(h) = oldest first-parent trunk commit containing h;
                                   track landing = newest of those ─► record, file
     some landed, or a bar not done ─► partly done: check in with operator
     none landed (branch/local) ─► "fetch/pull/push and retry", or ask for the PR or a commit,
                                   then archive or keep open (operator)
     no usable hashes ───────────► ask for the PR or a commit ─► found on trunk ─► record, file
     each bar's subject in a squash on trunk ─► landing = newest of the bars' squashes;
                                                each bar repointed to its own squash ─► record, file
```

## Decisions
- **Depends on BIT-47** (bar hashes, and `task_update` accepting `commit` and `branch`) **and BIT-49** (the git helper). It comes last: 45 → 46 → 48 → 49 → 47 → 50.
- **Split from BIT-47 (operator, 2026-10-01).** This track is merge-aware completion.
- **Cutover belongs to the operator.** This track doesn't gate it.
- **The track's `commit` is its landing or merge commit, the anchor that matters most (operator).** Completion records it.
- **Completion runs after the work has landed, not at sign-off (operator).**
- **Sign-off no longer completes the track (Claude default).** When the last bar is done and the operator signs off, bit:do tells them to push (or merge) and then run `/bit:complete`, and the track stays `doing` until then. Completion marks the bars and the track `done` and files them, which matches the sketch: "mark everything done, find the landing commit and record it, then move the track to completed". bit:bot's routing line for sign-off (`bit/agents/bot.md:41`) and its routing-table row (`bot.md:54`) change to match.
- **"Landed" means pushed (operator).** Trunk is `origin/main` when it exists, otherwise `main`. A commit only on local `main` is not done.
- **When `origin/main` doesn't exist (no `origin` remote, or one whose `main` was never fetched or pushed), local `main` is trunk and the "not landed" message says fetch or pull, not push (operator, 2026-10-02).**
- **Trunk is always named `main` (Claude default).** The decisions name `origin/main` and `main`, and every project in use has `main`. Other default-branch names aren't supported yet.
- **bp never fetches (operator).** Every "not landed" result first tells the operator to fetch or pull (and push, if commits are only local) and try again. No hardcoded author identity: detection rests on commits and messages.
- **The ladder, in order (operator):** the bars' commits already on trunk, then a squash on trunk whose body lists the bars' commit subjects, then ask the operator for the PR.
- **A bar's landing commit is the oldest commit on trunk's first-parent chain that has the bar's commit as an ancestor. The track's commit is the newest of its bars' landing commits (Claude default, revised 2026-10-02).** For a direct commit that is the bar's own commit. For a bar that reached trunk through a true merge, it is the merge commit that brought it in, which also covers v2 itself reaching main through a merge. The earlier recipe in topic `merge-commit` (`rev-list --first-parent --ancestry-path h..trunk | tail -1`) is wrong. It returns nothing when the bar's branch merged trunk back in before its own merge, and `git log -1` on that empty result silently gives HEAD. This was reproduced in a scratch repo, and `97a5279` on `origin/worktree-bit-39` has that shape (topics `review-2026-10-02`, `claim-audit-2026-10-02`).
- **Squash matching is per bar (Claude default, revised 2026-10-02).** A bar matches a squash on trunk's first-parent chain whose body lists the bar's commit subject as a `* subject` line, or whose subject line minus ` (#N)` equals it, since a single-commit squash has no list (topic `landing-ladder`). The earliest such squash after the bar's commit is the bar's landing, and the track's commit is the newest of those. The earlier rule, every bar's subject in one squash, can't complete real history: BIT-31 landed bars 1–3 in `4cf6137` (#5) and bar 4 in `e59da6e` (#6). A shared generic subject still can't match, because each bar needs its own matching line.
- **When completion matches a squash, each bar is repointed to its own squash before the track is filed (operator; order is a Claude default).** The store can't update a task once it's filed (topic `fields-and-writes`).
- **The write order at completion is: mark unfinished bars `done` (only with the operator's OK), repoint bars, set the track's `commit`, `branch` and `done`, then `task_complete` (Claude default, 2026-10-02).** `task_update` fails on a filed task (topic `claim-audit-2026-10-02`).
- **`task_complete` accepts `commit` and `branch`, and writes them on the track before it moves any file (Claude default, 2026-10-02; moved from BIT-47).** Completion is its only caller, so the input belongs here. BIT-47 adds `commit` and `branch` to `task_update` only.
- **Bar hashes are best-effort (operator).** A hash that doesn't resolve isn't proof of anything. Hashes are recorded and compared as full SHAs (Claude default, 2026-10-02).
- **Each bar is classed as landed, landed via squash, pushed but not merged, local only, unresolvable, or no hash (Claude default, from topic `done-classification`).** The track verdict follows:
  - every bar with a hash landed: done;
  - some landed and the rest pushed, local or unresolvable: partly done, so check in with the operator;
  - none landed, at least one pushed or local: not done;
  - every hash unresolvable or empty: can't tell, so ask for the PR. This isn't reported as not done.
- **When the operator says to go ahead with a partly done track, its commit is the newest landing commit among the landed bars (Claude default, 2026-10-02).**
- **A bar whose status isn't `done` makes the track partly done at best (Claude default, 2026-10-01).** Hashes alone can't see unfinished work: a bar that was never committed has no hash, and a bar migrated by BIT-49 carries HEAD at migration whatever its status. So the check-in names the unfinished bars, and the skill marks them `done` only when the operator says so. `task_complete` still refuses a track with an unfinished bar (`task/store.go:112-130`).
- **Migrated hashes count like any other (Claude default, 2026-10-01).** BIT-49 stamps migrated tracks and bars with HEAD at migration (operator). A done bar of a track migrated in flight then reads as landed at that HEAD. That's accepted: it only touches tracks that straddle cutover.
- **Only tracks with recorded hashes can land on their own (fact, 2026-10-02).** v1 records carry no hash (`task/task.go:19-28`). Verse 1 reaches tracks whose bars were committed through BIT-47's `bit:commit` or stamped by BIT-49's migrate. Everything else falls to asking for the PR. "Most of bit-pro's history is direct commits" (304 of 320, 0 merges) is still true, and it is why the squash rung comes last.
- **Not done (operator):** warn that the work probably should be archived, and let the operator decide. Work only on a local branch counts as not done, and the operator either keeps the track open as a reminder or archives it. A missing PR is a signal, not a rule.
- **Archiving goes through `task_delete`, the existing archive path (Claude default, 2026-10-02).** It force-archives unfinished bars only after the operator confirms.
- **Asking for the PR takes a PR number or a commit (Claude default).** A PR number is found on trunk by its `(#N)` squash subject, locally, without `gh`. If it isn't there, the operator is told to fetch and retry.
- **The `(#N)` lookup matches subjects ending in ` (#N)` on trunk's first-parent chain only. If more than one matches, the skill asks which (Claude default, 2026-10-02).** On this repo's history every PR number tested maps to one commit.
- **An answer of a PR or a commit repoints every bar that hasn't landed: no hash, unresolvable, pushed or local (operator, 2026-10-02).** Bars that landed keep their own commit.
- **A shallow clone is reported as such, with "run `git fetch --unshallow`", instead of classing its bars unresolvable (Claude default, 2026-10-02).**
- **Rebase-merges aren't detected (Claude default).** Their hashes don't survive on trunk and they leave no squash body (topic `landing-ladder`). Their bars read as pushed or local, so the track is not done, and the not-done check-in also offers "or tell me the PR number or a commit that landed it". An answer is checked exactly as for can't tell. A rebased `v2` takes the same path (operator, 2026-10-02).
- **The track's `branch` is the trunk it landed on, e.g. `main` (Claude default).** That's what "landed" means.
- **The landing logic lives in Go, behind a read-only MCP tool, `task_landing`, that the complete skill calls (Claude default; the name is 2026-10-01).** It's testable with BIT-49's stubbed git helper, and it keeps git commands out of skill prose. The skill runs the conversation: the check-ins, the archive-or-keep choice, the PR question.
- **`task_landing`'s shape (Claude default, 2026-10-02).** Input `{id, pr?, commit?}`. Output: `trunk`, `branch`, `verdict` (done, partly, not_done, cant_tell, no_git), `landing`, `shallow`, the unfinished bars, and per bar `{id, status, commit, class, landing}`. It never writes. The field names are bit:plan's to finalise, but these are the facts it returns.
- **`bp task complete` on the command line stays a plain filing command with no landing check, a documented manual override (Claude default, 2026-10-02).** Two tests use it as a fixture (`cmd/feedback_add_test.go:136`, `cmd/serve_mcp_research_test.go:124`). `Store.Complete` itself doesn't change.
- **The complete skill stops forcing unfinished bars to `done` and stops suggesting a filing commit (Claude default).** The forcing clashes with the "partly done" check-in. And in v2 the store is outside the repo, so filing changes nothing there to commit. Its evals, which expect both (`bit/skills/complete/evals/evals.json`, evals 1 and 2), are rewritten to match.
- **BIT-49's sweep changes only the wording of lines this track rewrites, and this track changes the behaviour (Claude default, 2026-10-02).** do's description and sign-off and the complete skill (`SKILL.md:3,8,12,23,30`) are edited by BIT-49, BIT-47 and this track in turn. Each later track anchors on section names, not line numbers.
- **Every bar that edits a skill or agent validates it (Claude default, 2026-10-01).** It runs skill-creator's `quick_validate.py` on each changed skill and `claude plugin validate ./bit` (memory `bit-skill-edits-run-skill-creator`; the kebab-case `name:` failure is known noise).
- **Dev runs follow BIT-46's rule (Claude default, 2026-10-01):** never `just install` on `v2`, and sandbox `XDG_DATA_HOME` and `HOME`. The ladder is exercised in scratch git repos with a local bare `origin`. Nothing on `v2` reaches `origin/main` before cutover, so scratch repos are the only test path.
- **Each bar leaves `just lint` and `just test` green (Claude default, 2026-10-02).** That is what the pre-commit hook runs (`.pre-commit-config.yaml`).
- **Git facts come from the session dir (BIT-49).** Git fields may be empty.
- **In an `acme/`-style folder with no git, completion files the track with no landing commit, after the operator confirms.** This is Verse 2's no_git verdict (Claude default; verse assigned 2026-10-02).
- **Verse order (Claude default, revised 2026-10-02).** The PR answer (now Verse 3) comes before the squash rung (now Verse 4). Verse 2 ends at "ask for the PR", so the answer path has to exist next. Squashes are historical here: none in the last 20 commits, and v2's bars are direct commits through `bit:commit`.

## Verses
- [ ] Verse 1 — The operator completes a track after pushing it, and the track records where it landed. `/bit:complete` finds every bar commit on trunk (directly or through a merge), records the track's landing commit and branch, marks everything `done` and files it. Anything else gets "not landed yet: fetch, pull or push and retry", and nothing is filed. bit:do's sign-off now points to this step. Direct commits are most of bit-pro's history (304 of 320 commits).
  Touches: `bit/skills/complete/SKILL.md` (description, intro, steps, report) and its evals; `bit/skills/do/SKILL.md` (description, the track sign-off section, and the lines that say sign-off completes the track; today `:3`, `:15`, `:95`, `:100`, `:102-110`, `:131`); `bit/agents/bot.md` (`:41`, `:54`); `bit/agents/bot-dev.md:45`; `cmd/serve_mcp.go` (`task_complete`'s description, input and handler, and the new `task_landing` tool; today `:76-82`, `:160-162`, `:265-268`, `:417-430`) and `cmd/serve_mcp_test.go` (the tool-description test, today `:201`); `cmd/task/complete.go:12` and `README.md` (`:65`, `:115`, the "signed off" wording); BIT-49's git helper. Line numbers are as of `6a1d345` and drift after BIT-46, BIT-49 and BIT-47. See topics `completion-today`, `landing-ladder`, `merge-commit` and `claim-audit-2026-10-02`.
- [ ] Verse 2 — The operator is told when a track is only partly landed, not landed, or can't be traced, and decides what happens. Partly done (including unfinished bars) gets a check-in, not done gets the archive-or-keep choice, a folder with no git gets a confirm-and-file, and can't tell moves on to asking for the PR.
  Touches: `task_landing`'s classification, `bit/skills/complete/SKILL.md` and its evals. See topic `done-classification`.
- [ ] Verse 3 — When git can't place the work, or the work isn't on trunk, the operator answers with a PR number or a commit, and completion checks it on trunk and records it.
  Touches: `task_landing` (`(#N)` lookup), `bit/skills/complete/SKILL.md`. See topic `landing-ladder`.
- [ ] Verse 4 — A track that landed as GitHub squashes completes on its own. Completion finds the squash that lists each bar's subject, records the newest as the track's commit and repoints each bar to its own squash.
  Touches: `task_landing` (the squash rung), `bit/skills/complete/SKILL.md`. See topic `landing-ladder`.

## References
- `.bit/research/BIT-50/`: start at `index`. Topics `completion-today`, `landing-ladder`, `done-classification`, `merge-commit`, `fields-and-writes`, `verse-order` and `decisions`. `review-2026-10-02` and `claim-audit-2026-10-02` (2026-10-02) verify every claim and correct the landing formula and the one-squash rule. They win where the older topics differ.
- `.bit/research/BIT-47/`: topics `soundness` (real `git log main` evidence) and `soundness-2` (the fetch/origin gap).
- `.bit/research/BIT-45/`: topic `history-anchors` (read its banner first).
- `v2-sketch.md` (repo root).
