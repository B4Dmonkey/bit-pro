# The landing ladder against bit-pro's real history

**Checked:** rungs (a) bars on origin/main, (b) squash body lists bar subjects, (c) ask for the PR — feasibility, exact commands, no-fetch behaviour. git 2.54.0.

## Repo facts (2026-10-01)
- `git rev-parse main origin/main v2` → all `6a1d345`. v2 currently has no commits beyond main (`git log main..v2` empty).
- `git log --oneline main | wc -l` = 320; `git log --merges main` = **0 merge commits**. 16 subjects end `(#N)` (squash merges #1–#17, #11 absent); the other 304 are direct.
- Remote-tracking refs still exist for squashed branches, e.g. `origin/worktree-agent-a6848e37a1646d392`, `origin/worktree-bit-39`, `-bit-40`, `-bit-35`; `git merge-base --is-ancestor origin/<b> origin/main` → exit 1 for all (not ancestors, as expected after squash).

## Rung (a): bars' commits already on trunk
- Trunk ref: `git rev-parse --verify --quiet refs/remotes/origin/main` (exit 0 → use `origin/main`; exit 1 → no origin, use `main`). Assumes trunk is named `main`; `git symbolic-ref refs/remotes/origin/HEAD` (→ `origin/main` here) would avoid hardcoding — **unknown whether to support non-`main` trunks**.
- Per bar: `git cat-file -e <h>^{commit}` (exit 128 = unresolvable), then `git merge-base --is-ancestor <h> <trunk>` (0 = landed, 1 = not, 128 = bad object). Verified: `cebf42d` resolves, is-ancestor exit 1.
- "Last of them": `git rev-list -n1 --topo-order <h1> <h2> …` returns the newest tip (verified in a scratch repo); bar order in the track need not match commit order.
- Landing commit for a bar: if the bar is on the first-parent chain (`git rev-list --first-parent <trunk> | grep <h>`), it is its own landing; otherwise `git rev-list --first-parent --ancestry-path <h>..<trunk> | tail -1` gives the merge commit that brought it in (see `merge-commit`).

## Rung (b): squash body lists bar subjects
- Subject: `git log -1 --format=%s <h>` (needs the bar hash to resolve).
- Find candidates: `git log <trunk> -F --grep="<subject>" --format=%H` (`-F` = fixed strings; matches the whole message). Verified: `refactor(skills): drive bit:do through MCP tools` → `54abfeb (#17)` and `ad04ca4 (#16)`.
- GitHub multi-commit squash body is `* <subject>` per commit (`git show -s 54abfeb`, `1eec300`); merge commits inside the branch appear too (`worktree-bit-39` had `Merge remote-tracking branch 'origin/main'`). **Single-commit squash** (`e59da6e (#6)`) has no `*` list: the squash subject is the PR title + ` (#6)` and the body is the commit body — so matching must look at the subject line too, and fails if the PR title was edited.
- **Surprise: one branch squashed twice.** #16 and #17 both list all 11 subjects of the a6848e37 branch (and #14/#15 overlap for bit-39). Rule needed: pick the squash that contains *all* bars' subjects; if several, **unknown** whether earliest (first landing) or latest (final state) is the anchor — earliest-on-first-parent seems right but is an operator call.
- False positives: generic subjects recur (`chore(bit): file completed BIT-34` twice on `worktree-bit-35`). Require all bars' subjects in one commit, not any.

## Rung (c): ask for the PR
- A PR number maps to the squash locally with no network: `git log <trunk> -F --grep="(#17)" --format=%H` (GitHub appends `(#N)` to squash subjects; true merge commits say `Merge pull request #N`). So the operator's answer can be checked without `gh`.

## No fetch
All commands read local objects and remote-tracking refs only; none touch the network. Staleness is the risk: a squash merged on GitHub, or a branch deleted remotely, is invisible until fetch. Hence the decided "fetch/pull (and push) and retry" message before suggesting archive. Commands must run with `git -C <session dir>` (BIT-49 helper decision).

## Gaps the ladder doesn't cover
- **Rebase-merge or local rebase** (new hashes; note the `backup/main-prerebase` branch shows rebases happen here): rung (a) fails, (b) fails (no squash body), falls to (c). Exact-subject lookup on trunk or `git cherry <trunk> <h>` (patch-id) would catch it — not in the decided ladder.
- Bars with no commit (docs-only, or pre-BIT-47 bars) contribute nothing; a track with zero hashes goes straight to (c).

**Verdict:** feasible, all rungs verified on real history; corrections: single-commit squash shape, duplicate squashes, PR→squash via `(#N)`.
