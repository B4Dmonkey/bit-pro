---
id: BIT-47.3
title: bit:commit asks, commits a bar's files with its message, and records the hash
status: todo
approved: true
phase: 1
phase_label: a bar lands as a real commit with its hash recorded
---
## **Verse 1**

The new focused skill that every commit goes through. It's skill text plus evals, so there's no Go test. It's checked with skill-creator, `claude plugin validate`, a JSON check and a sandboxed run. Nothing calls it yet, so bit:do still says the user commits until BIT-47.4 (one bar's window, accepted in the scope's bar order). It's usable on its own as `/bit:commit <bar>`.

## Scope
- new `bit/skills/commit/SKILL.md`. Frontmatter `name: bit_commit` (the `bit_<x>` convention; the kebab-case validator failure is known noise). The `description` says it commits one bar's work after asking the operator, records the commit's full hash and branch on the bar, and marks it `done`. Triggers: "commit this bar", "commit BIT-N.M", `/bit:commit`, and bit_do's or bot-dev's close-out. It only commits: it doesn't push, roll the track up or start the next bar. The body, in order:
  1. **Read the bar** with `mcp__bit__task_read`. The message is the backticked line under `## Commit`, or `## Commit (user)` on older bars. Refine it only if the work diverged. On a follow-up commit to a bar that already has a `commit` (an unwind fix or a small cleanup), reuse that commit's subject line (`git log -1 --format=%s <commit>`) so BIT-50's squash matching still finds the bar.
  2. **No repo:** `git rev-parse --is-inside-work-tree` fails (an `acme/`-style folder, or work outside the repo). Say there's nothing to commit here. On the operator's OK, `task_update {id, status: done}` with no git fields.
  3. **Pick the files:** `git status --porcelain`. Stage only what this bar touched (its Scope files and anything created for it). List any unrelated changes and leave them alone (moved from bot-dev's "How to commit").
  4. **Nothing to commit:** say so. On the operator's OK, `task_update {id, status: done}` with no git fields. Never `git commit --allow-empty`.
  5. **Ask in prose, before staging anything:** show the file list, `git diff --stat -- <files>` (and name new untracked files), and the full message. Then wait. Only an explicit yes goes on. This ask exists even where the permission settings would allow the commit, and it replaces "the user reads the diff when they commit".
  6. **Declined** (a no, or anything that isn't a yes): say the commit was declined. The bar stays `doing`, nothing is staged and nothing is recorded. Stop.
  7. **Commit:** `git add -- <files>` and then `git commit` with the message, as two separate Bash calls. Each can raise its own permission prompt, which is accepted (scope Decision). A denied prompt counts as declined: report what's staged, leave the bar `doing` and record nothing.
  8. **Hook fails or rewrites files:** show the hook output. If the hook changed files the bar touched, re-stage them and go back to step 5 with the new stat. If a lint or test hook failed, the bar's work isn't finished: stop, so the caller fixes it within the bar's scope and runs bit_commit again. Never `--no-verify`. The bar stays `doing` until a commit succeeds.
  9. **Read the result:** `git rev-parse HEAD` (the full 40-character SHA) and `git symbolic-ref --short -q HEAD`. That exits non-zero on a detached HEAD, which means the branch is `""`. These are the same commands as BIT-49's `git.ReadHead`.
  10. **Record:** one `mcp__bit__task_update {id, commit: <sha>, branch: <branch>, status: "done"}`. It returns `approved: true`, because git fields and a forward move keep approval (BIT-47.2). A follow-up commit overwrites `commit` in the same way.
  11. **Report:** the short SHA, the branch and the subject. Then hand back to the caller (bit_do rolls the track up, and bot-dev pushes).
  - A "What this skill does not do" list: push, roll up, mark a track done, `--no-verify`, `--allow-empty`, amend, or commit without the prose yes.
- new `bit/skills/commit/evals/evals.json` in the shape of `bit/skills/complete/evals/evals.json` (`skill_name: "bit_commit"`, `evals[{id, prompt, expected_output, files: []}]`):
  1. `"commit BIT-7.2"`: shows the files, the diff stat and the message taken from `## Commit (user)`, and runs no `git add`/`git commit` before a yes.
  2. `"no, hold off on committing BIT-7.2"` after the ask: the bar stays `doing`, nothing is staged, and no `task_update` carries `commit`.
  3. `"commit BIT-7.3"` with a clean tree: says there's nothing to commit, never `--allow-empty`, and marks the bar `done` with empty git fields only after the operator agrees.
  4. `"commit BIT-7.4"` where the pre-commit hook reformats a staged file: shows the hook output, re-stages, asks again, never `--no-verify`, and the bar stays `doing` until the commit succeeds.
  5. `"commit BIT-7.5"` on a detached HEAD: commits, then `task_update` sends the full SHA, `branch: ""` and `status: done`.
- Skills are auto-discovered from `bit/skills/<dir>/` (`plugin.json` lists none), so no registration is needed.
- `README.md`: add `commit` to the skill list under `## Agent skills` that BIT-49.30 wrote (no count to change), in the same form as its neighbours, with one line saying it commits a bar's files after asking and records the hash on the bar.
- **Seam for BIT-50:** every bar committed through `bit:commit` holds its own commit's full SHA in `commit` and the branch it was committed on in `branch` (`""` on a detached HEAD). A follow-up commit overwrites `commit` and keeps the subject.

## Change checklist
- [ ] Write `SKILL.md` and `evals/evals.json` as above.
- [ ] Add `commit` to the README's skill list.
- [ ] `grep -n 'no-verify\|allow-empty' bit/skills/commit/SKILL.md` shows each only as something the skill never does.
- [ ] `sed -n '/^## Agent skills/,/^## [^A]/p' README.md | grep -nE 'bit[_:]commit'` prints the new entry.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/commit/`. Only the known kebab-case `name:` failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `python3 -m json.tool bit/skills/commit/evals/evals.json > /dev/null`
- [ ] `just lint` and `just test` pass

## User verifies
Sandbox as in BIT-46.11's first bullet (never `just install`), then in that shell:
`cd $SB/proj && git init -b main && git -c user.name=t -c user.email=t@t commit --allow-empty -m init && printf 'demo\n' | bp add . && bp task create "T" && bp task create "Add hello" -p DEMO-1 -d $'## **Verse 1**\n\n## Commit (user)\n`feat(demo): add hello`' && bp approve DEMO-1.1 && bp task update DEMO-1.1 --status doing && printf 'hello\n' > hello.txt`.
From a new shell without the `HOME` export, write `$SB/dev.json` as in BIT-46.11, then `cd $SB/proj && claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config`.
- [ ] `/bit:commit DEMO-1.1`: the ask lists `hello.txt` and `feat(demo): add hello`. Answer "no". `git log --oneline` shows only `init`, and `$SB/data/bit/DEMO/tasks/DEMO-1.1.json` has `"status": "doing"` and `"commit": ""`.
- [ ] `/bit:commit DEMO-1.1` again and answer yes, then allow each git permission prompt. `git log -1 --format='%H %s'` prints `<sha> feat(demo): add hello`. `DEMO-1.1.json` has `"commit": "<that same full sha>"`, `"branch": "main"`, `"status": "done"` and `"approved": true`.

## Commit
`feat(bit): add the bit:commit skill`
