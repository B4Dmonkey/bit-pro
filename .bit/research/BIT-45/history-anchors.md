> **Superseded in part (banner added 2026-10-01).** The track bodies (BIT-46, BIT-47, BIT-49, BIT-50) win over this note:
> - **No project path or session dir in records.** Paths live only in the db (operator, 2026-09-30). "Record the session dir as well as the project path" below is out. Git facts still come from the session dir (now a BIT-49 git-helper decision). They just aren't stored as a path.
> - **`commit`, not `head`, main head or base.** Tracks and bars carry `branch` and a single `commit` (track = landing commit, the anchor that matters most; bar = best-effort, repointed to the squash commit when one is detected). Research topics, feedback notes and retro proposals carry a `commits` list of `{sha, branch, at}`: the first entry is HEAD when the record is created, and a later write appends one when HEAD has moved (operator, 2026-10-01).
> - **HEAD isn't captured on every write** for tracks and bars. A bar is committed first, by Claude through a dedicated commit skill that always asks the operator's permission (bot-dev included), and its hash is stored afterwards (BIT-47). The "off by one" problem below goes away.
> - **The squash commit** is found by completion after the work has landed (pushed), through a ladder: bars' commits on main, then a squash whose body lists their subjects, then ask for the PR (BIT-50). The open design question below is settled.
> - **Migrate** writes HEAD at migration and the migration time, not `git log --follow` or empty anchors (BIT-49).
> - **Order:** this is BIT-47 (commit side) and then BIT-50 (completion), which come last (45 → 46 → 48 → 49 → 47 → 50).

# Git history anchors for before/after state (Q3)

**Checked:** with squash merges, which git facts recover a track's before and after state, when bit would capture each, how bit gets git info today, and what worktrees change.

## Today
**bp never calls git.** The only `exec`s are `claude` (`claude/sync.go`, `plugin.go`, `dispatch.go`) and `launchctl` (`daemon/daemon.go`). No record carries a timestamp or a commit. Everything in this note is new code.

## How bit-pro actually merges (observed)
- PRs are squash-merged on GitHub. The squash subject is the PR title plus `(#N)`, e.g. `84886bb Worktree bit 39 (#15)`. The **squash body lists every branch commit subject** (`* feat(daemon): …`), so the branch commit *messages* survive on main. Their hashes and per-commit diffs do not.
- Branches are named after the worktree (`worktree-bit-39`, `worktree-agent-<id>`). Several `origin/worktree-*` branches are still on the remote, so branches are not auto-deleted here.
- GitHub also keeps `refs/pull/N/head` after a branch is deleted. That is a way back to the branch head via `gh`/`git fetch origin pull/N/head`. This is from general GitHub knowledge; verify it before relying on it.
- `origin/HEAD -> origin/main` is set, so `git symbolic-ref refs/remotes/origin/HEAD` gives the default branch name without hardcoding `main`.

## Anchors and when to capture them
| Fact | Command (run with `Dir` = the session's actual dir) | Captured at |
|---|---|---|
| branch | `git rev-parse --abbrev-ref HEAD` | every write (create, update, status change) |
| head | `git rev-parse HEAD` | every write → "last commit hash" |
| main head | `git rev-parse <default>`, local ref, no fetch | every write, "possibly last main hash" |
| base | `git merge-base HEAD <default>` | track create, and again at done |
| squash commit on main | `gh pr view <branch> --json mergeCommit`, or `git log <default> --grep '(#N)'` | **after** merge. Not available at done (see below) |

- **Before state:** `merge-base(branch head, main)` recorded at done is more accurate than main's head at create, because the branch may have been rebased. Keep both. They are cheap.
- **After state:** the squash commit. **`task_complete` runs on the user's sign-off, which is usually before the PR is merged.** So bit can't know the squash commit at done. Options:
  1. Record the branch name and head at done, and resolve the squash commit lazily later (a `bp` command or retro-time lookup via `gh` or the subject grep).
  2. A later "merged" status.

  This is an **open design question**.
- **Bar done ≠ bar commit:**
  - bit:do moves a bar to `done` and then the user commits, so `HEAD` at the status change is the commit *before* the bar's work.
  - bot-dev commits after the status change too.
  - So per-bar hashes captured at status time are off by one. Either accept `updated_at` plus branch as the anchor, or capture HEAD on the *next* write.
- Local `main` can be stale (no fetch). Recording `origin/<default>` as well is cheap. Never fetch implicitly.

## Worktrees
- The project comes from the registered path ([project-resolution](project-resolution.md)), but **git facts must come from the session dir**. A Claude worktree (`<repo>/.claude/worktrees/x`) has its own branch and HEAD.
  - CLI: `os.Getwd()`.
  - MCP: `$CLAUDE_PROJECT_DIR`. v1's `worktreeCut` exists because this env var is the worktree path in worktree sessions (pinned by `bitdir_test.go`).
- Record the session dir as well as the project path. It shows which worktree wrote the record.
- Non-git projects (e.g. the `acme/` folder) have empty git fields. Capture must never fail a write. Treat a git error as blank.

## Implementation notes
- Shell out to `git` (`exec.CommandContext`, `Dir` set) behind an injectable func, like `claude.Runner`, so tests don't need a repo. Alternatively build a real temp repo in tests. go-git would be a heavy dependency for four rev-parses.
- Cost: 3-4 git execs per write, a few ms each. Fine for MCP.
- Migrate can't reconstruct historical anchors. For git-tracked `.bit/` (bit-pro), `git log --follow --format=%H,%aI -- .bit/tasks/X.md` recovers created/updated times and the commits that touched the file. Otherwise leave anchors empty and mark the record `migrated`.