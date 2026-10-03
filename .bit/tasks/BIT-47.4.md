---
id: BIT-47.4
title: bit:do closes a bar through bit:commit, and bot-dev asks before every commit and pushes only after one
status: todo
approved: true
phase: 1
phase_label: a bar lands as a real commit with its hash recorded
---
## **Verse 1**

Once do's close-out commits through `bit:commit`, bot-dev's "bit:do leaves the commit to the user … then commit" is false, and bot-dev would try a second commit. So both change in this one bar (scope Decision "Verse shape"). Text only, so there's no Go test. It's checked with skill-creator, `claude plugin validate` and a sandboxed run. This bar completes the verse.

## Scope
Anchor on section names, not lines. BIT-49.27 and BIT-49.29 have already edited these files (the `.bit/` wording, the staging sentences). Re-read both files first.
- `bit/skills/do/SKILL.md`:
  - `description:`: "and hands off to the user for verification and commit between bars" → "and commits each verified bar through bit_commit, which asks the operator first, before handing off between bars". Leave the sign-off sentence alone (BIT-50 rewrites it).
  - Intro paragraph ("carry out **one bar, then stop**"): "Verification is the user's call, and so is the commit." → "Verification is the user's call, and no commit happens without their yes, through bit_commit."
  - `### 5. Close out the bar`, the User-verifies paragraph: instead of stating the suggested commit message, present the checklist and show the message bit_commit will use. Then stop and leave the bar `doing`. When they confirm, run **Verified good**, which commits.
  - The no-User-verifies paragraph: run **Verified good** inline (commit through bit_commit, roll up, compaction point). Replace "the user still reads the diff when they commit" with "bit_commit shows the files and the message, and nothing is committed until the operator says yes". Unwind: the commit stays. Set the bar back to `doing`, reverse any verse checkoff and treat it as **Not as expected**. The fix lands as a follow-up commit through bit_commit, which overwrites the bar's `commit` and reuses its subject. Keep the sentence that the call returns `approved: true`. BIT-49.29 already deleted the paragraph's last sentence ("The done state is cheap to undo, and marking it now keeps the `.bit/tasks/*.md` status change …"), so there's nothing more to drop.
  - "Either way, two lines hold firm": → never commit except through bit_commit, and do **not** start the next bar.
  - The small-cleanup paragraph: if the bar is already committed, a cleanup reply ends by offering a follow-up commit through bit_commit. If it isn't committed yet, it ends with the message bit_commit will use.
  - `### Verified good` steps become: 1. **Commit through bit_commit.** It asks, commits, and records `commit`, `branch` and `done` in one `task_update`. If the commit is declined, or a hook failure is unresolved, the bar stays `doing`: stop, with no rollup. With nothing to commit or no repo, bit_commit marks the bar `done` after the operator's OK. 2. **Roll the track up** (unchanged). 3. **Compaction point** (unchanged). Remove the "Mark the bar done" and "Suggest the commit" steps, since bit_commit does the first and the second is gone.
  - `### Track sign-off`: "(verified, commit suggested)" → "(verified, committed)". Nothing else in that section changes (BIT-50's).
  - `### Not as expected`, last paragraph: add that a committed bar keeps its commit, and its fix is a follow-up commit through bit_commit.
  - `## What this skill does not do`: the "Commit" bullet → "**Commit on its own.** Every commit goes through bit_commit, which asks first. A declined commit leaves the bar `doing`." Add "**Push.** bit_do never pushes. bot-dev does, after a permitted commit."
- `bit/agents/bot-dev.md`:
  - `description:`: "Executes one bar of a bit plan and lands it. Runs the bit:do skill exactly as written, whose close-out commits through bit:commit once the operator says yes, then pushes that commit if the repo has a remote. Every commit waits for the operator: in a headless run (`-p`, `--bg`) it stops at the ask and reports the files and the message, and the operator answers by resuming the session. Use when a bar should be implemented, committed and pushed in one go." Drop "the one thing that skill leaves to a human", "without an operator sitting in front of it" and the `--bg` dispatch example.
  - Intro: "land it as a commit" → "land it as a pushed commit". Replace "You are written for an operator who is **not watching**…" with: bit:do already commits through bit:commit, and what you add is the push and how to ask when nobody may be watching.
  - `## The one delta: you commit, and push` → `## The one delta: you push`. The body says bit:do's Verified good commits through bit:commit on every bar, with or without User verifies, and that dispatching you isn't permission. Never commit a second time or outside bit:commit. A bar with User verifies: present them, leave the bar `doing` and stop. On confirmation, Verified good commits.
  - New `### When nobody can answer`: in a `-p` or `--bg` run, bit:commit's ask ends your turn. The bar stays `doing` and nothing is staged. The report lists the files and the commit message, and the operator answers by resuming the session.
  - Delete `### How to commit`. Its message and staging rules now live in bit:commit.
  - `### Then push…`: push only after bit:commit reports a commit. A declined commit, or a bar with nothing to commit, isn't pushed. Steps 1–3 and "Do not open a PR" stay.
  - `## What stays the operator's`: add "**Every commit.** bit:commit asks each time, and dispatching you isn't a yes." Leave the "Track sign-off" bullet for BIT-50.
- Text only, with no test files.

## Change checklist
- [ ] Edit both files as above.
- [ ] `grep -n -i "user commits\|never run the commit\|suggest the commit\|commit suggested\|always the user's action\|not watching\|sitting in front" bit/skills/do/SKILL.md bit/agents/bot-dev.md` finds nothing.
- [ ] `grep -n "bit_commit\|bit:commit" bit/skills/do/SKILL.md bit/agents/bot-dev.md` shows the Verified good step, the close-out paragraphs and bot-dev's delta.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/do/`. Only the known kebab-case `name:` failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `just lint` and `just test` pass

## User verifies
Reuse BIT-47.3's sandbox and `$SB/dev.json`. In the sandboxed shell, add two approved bars:
`cd $SB/proj && for n in world bang; do bp task create "Add $n" -p DEMO-1 -d "## **Verse 1**"$'\n\n'"Create \`$n.txt\` containing \`$n\`."$'\n\n## User verifies\n- none\n\n## Commit\n'"\`feat(demo): add $n\`"; done && bp approve DEMO-1.2 && bp approve DEMO-1.3`.
Then, from the shell without the `HOME` export, in `$SB/proj`:
- [ ] `claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config`, then `/bit:do DEMO-1.2`. It creates `world.txt`, and its close-out asks through bit:commit with `world.txt` and `feat(demo): add world`. Say yes and allow the prompts. `git log -1 --format=%H` equals `"commit"` in `$SB/data/bit/DEMO/tasks/DEMO-1.2.json`, which also has `"branch": "main"` and `"status": "done"`. The session never offers a second commit.
- [ ] `claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config --agent bit:bot-dev --permission-mode acceptEdits --allowedTools 'mcp__bit__*' -p "/bit:do DEMO-1.3"`. The two permission flags only let the headless run write `bang.txt` and call the bit tools. They don't cover `git add` or `git commit`. The output ends with the ask, naming `bang.txt` and `feat(demo): add bang`. `git status --porcelain` shows `?? bang.txt` (nothing staged), `git log -1 --format=%s` is still `feat(demo): add world`, and `DEMO-1.3.json` has `"status": "doing"`.
- [ ] Resume it (`claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config --agent bit:bot-dev --continue`) and answer yes. It commits, `DEMO-1.3.json` has the new HEAD as `"commit"`, and bot-dev reports the commit as local-only, because `git remote` is empty.
- [ ] Whole slice: in bit:do and in bot-dev, a bar becomes `done` only as a real commit, made after the operator's yes, with its full hash and branch on the bar.

## Commit
`feat(bit): bit:do and bot-dev commit through bit:commit`