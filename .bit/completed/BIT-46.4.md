---
id: BIT-46.4
title: The longest registered path containing a folder resolves its project
status: done
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

The resolver that replaces `bitdir`. Built and unit-tested on its own; the CLI switches to it in a later bar. A single hardcoded match can't satisfy subfolder, nested-registration, symlink and case tests together, so they force the real matcher.

## Scope
- new `project/resolve.go`:
  - `type Project struct { ID int64; Code string; Path string }`
  - `var ErrNotRegistered = errors.New("not a bit project; run `bp add`")` (text from the track).
  - `var ErrNeedsMigrate = errors.New("found a v1 .bit/ directory; run `bp migrate`")`.
  - `func CanonicalPath(path string) (string, error)` — `filepath.Abs`, then `filepath.EvalSymlinks`. If `EvalSymlinks` fails because the path doesn't exist, walk up the cleaned absolute path to its deepest existing ancestor, `EvalSymlinks` that, and join the missing tail back on. Returning the cleaned absolute path unresolved would not match: the stored path is `/private/var/...` while a missing worktree path under `t.TempDir()` is `/var/...`, so `TestResolve/missing dir` and the MCP worktree test (`cmd/serve_mcp_test.go:104`) would both fail. **Seam:** `bp add` stores this; BIT-49 migrate registers with it.
  - `func Resolve(projects []Project, dir string) (Project, error)` — canonicalizes `dir` and each `p.Path`; a project matches when the canonical dir equals its path or starts with path + `/`, compared with `strings.EqualFold` (whole segments, case-insensitive); the longest matching path wins. No match: walk up from the canonical dir looking for a `.bit` directory → `fmt.Errorf("%s: %w", dir, ErrNeedsMigrate)`, else `fmt.Errorf("%s: %w", dir, ErrNotRegistered)`.
  - `func Load(ctx context.Context, q *orm.Queries) ([]Project, error)` — maps `ListProjects` rows. **Seam:** BIT-49 checks path/code collisions on this list.
  - `func Find(ctx context.Context, dir string) (Project, error)` — `db.Open`, `Load`, `Resolve`.
- `project/resolve_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestResolve/finds the project from a subfolder` (first row of the `TestResolve` table; each row builds its dirs under `root := t.TempDir()`, then gives its projects, the dir to resolve, and either the wanted code or the wanted error and message)
     - **Behavior:** bp run anywhere under a registered folder finds that project.
     - **Setup:** `root := t.TempDir()`; `mkdir root/src/pkg`; projects `[{Code:"BIT", Path: root}]`; `Resolve(projects, root+"/src/pkg")`.
     - **Assertions:** returned `Code == "BIT"`, nil error.
     - **Boundary:** depth 2 below the registered path (v1 failed at depth ≥ 1).
   - [ ] Confirm fails: `Resolve` undefined.

2. **Implement (GREEN):**
   - [ ] Hardcoded return of `projects[0]` is acceptable here.

3. **More tests (RED → GREEN):** (rows added to the `TestResolve` table)
   - [ ] `TestResolve/sibling prefix is not a match`: registered `root/app`, dir `root/application` → `ErrNotRegistered`. *Boundary:* segment boundary; a plain string prefix would match. (This contradicts the hardcoded return.)
   - [ ] `TestResolve/nested registration longest wins` and `TestResolve/nested registration outer folder`: `[{ACME, root}, {API, root/api}]`, dir `root/api/x` → `API`; dir `root/docs` → `ACME`. *Boundary:* two matches, either list order.
   - [ ] `TestResolve/claude worktree`: dir `root/.claude/worktrees/hazy/sub` (created) → the `root` project. *Boundary:* the worktree case needs no special code.
   - [ ] `TestResolve/missing dir`: dir `root/.claude/worktrees/wt` not created → the `root` project. *Boundary:* `EvalSymlinks` fails on the dir while the registered `root` is given as `/var/...` and resolves to `/private/var/...`, so only the ancestor-resolving fallback matches (the MCP worktree test passes a missing path, `cmd/serve_mcp_test.go:104`).
   - [ ] `TestResolve/symlinked temp dir as registered` and `TestResolve/symlinked temp dir as typed`: registered path is `t.TempDir()` as given (`/var/...`), dir is its `EvalSymlinks` form (`/private/var/...`), and the reverse → match. *Boundary:* the macOS `/var` symlink, both directions.
   - [ ] `TestResolve/case differs in the typed path`: registered `root/Repo`, dir `root/repo/x` (created as `Repo`) → match. *Boundary:* case differs only in the typed path (APFS case-insensitive).
   - [ ] `TestResolve/unregistered with bit above`: `mkdir root/.bit root/sub`, projects empty, dir `root/sub` → `errors.Is(err, ErrNeedsMigrate)` and `err.Error()` contains ``run `bp migrate` ``.
   - [ ] `TestResolve/unregistered without bit`: → `errors.Is(err, ErrNotRegistered)`, contains ``not a bit project; run `bp add` ``.
   - [ ] `TestCanonicalPath` (table): `TestCanonicalPath/existing dir resolves symlinks` (existing temp dir → its `EvalSymlinks` form); `TestCanonicalPath/missing relative path resolves its existing parent` (`root := t.TempDir()`, `t.Chdir(root)`, `CanonicalPath("a/../b/c")` → `EvalSymlinks(root)` + `/b/c`, nil error).
   - [ ] `TestFind/resolves through the registry` (subtest): sandboxed `HOME`, `XDG_DATA_HOME=""`; `CreateProject{Path: root, Code: "BIT"}`; `Find(ctx, root+"/sub")` → `BIT`. *Boundary:* the db round trip, one row.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): resolve the project by longest registered path`