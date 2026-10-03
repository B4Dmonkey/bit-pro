---
id: BIT-50.1
title: task_landing reports a track whose bars are on trunk as done
status: todo
phase: 1
phase_label: complete after push
---
## **Verse 1**

The complete skill needs one read-only call that says whether a track landed and where. This bar stands `task_landing` up end to end with the simplest rule that passes a fully pushed track: verdict `done`, and the landing is the last bar's commit. BIT-50.2's unpushed bar contradicts it.

## Scope
- new package `landing` (`landing/landing.go`), the domain logic behind the tool:
  - `type Bar struct{ ID, Status, Commit string }`.
  - `type Query struct{ Dir string; Bars []Bar }`. It's a params struct, so Verse 3 can add the operator's answer without changing callers.
  - `type Verdict string` with `Done Verdict = "done"`. `type Class string` with `Landed Class = "landed"`.
  - `type BarResult struct` with `ID`, `Status`, `Commit`, `Class`, `Landing`, tagged `json:"id"`, `"status"`, `"commit"`, `"class"`, `"landing"`.
  - `type Report struct` with `Trunk`, `Branch`, `Verdict`, `Landing`, `Bars []BarResult`, tagged `json:"trunk"`, `"branch"`, `"verdict"`, `"landing"`, `"bars"`. `Bars` is never nil, so it marshals as `[]`.
  - `func Check(ctx context.Context, run git.Runner, q Query) (Report, error)`. This bar hardcodes `Trunk: "origin/main"`, `Branch: "main"`, `Verdict: Done`, each bar `Landed` with `Landing = Commit`, and the report's `Landing` set to the last bar's commit.
- extend package `gittest` (`git/gittest/gittest.go`), created by BIT-49.9 with `Env`, `Run` and `Isolate`, so the `landing` and `cmd` tests share one repo fixture:
  - `func New(t testing.TB) *Repo`: `exec.LookPath("git")`, else `t.Skip`. It calls `Isolate(t)` first, so `git.ExecRunner` in the code under test never inherits the hook's `GIT_*` variables. Under `t.TempDir()`: `git init --bare -b main origin.git`, then `git init -b main work`, `git remote add origin <abs origin.git>`, one commit (`Commit("chore: root")`), and `git push -u origin main`. `type Repo struct{ t testing.TB; Dir, Origin string; n int }`.
  - `func (r *Repo) Git(args ...string) string`: runs `git -c user.name=bit -c user.email=bit@example.com <args>` in `r.Dir`, with `cmd.Env = append(Env(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")`, so neither the hook's `GIT_*` variables nor the developer's config leak in. It returns trimmed stdout and calls `t.Fatalf` with stderr on failure.
  - `func (r *Repo) Commit(subject string) string`: writes a new file `f<n>.txt`, runs `add` and `commit -m subject`, and returns `rev-parse HEAD`, the full SHA. A real tree change keeps later merges and squashes non-empty.
- `cmd/serve_mcp.go`:
  - `taskLandingTool = "task_landing"`. `taskLandingDescription` says the tool reports where a track's bars landed on trunk (`origin/main`, else `main`), that it's read-only, and that it never fetches or writes. It contains `testTrackSentence`.
  - `taskLandingInput{ID string \`json:"id"\`}`.
  - `taskLandingHandler(root string, run git.Runner) mcp.ToolHandlerFor[taskLandingInput, landing.Report]`:
    - Get the store with `mcpStore(ctx, root)` (BIT-46.9), then `store.Load(in.ID)`.
    - Map `store.Children(id)` to `[]landing.Bar`, with `Commit` from `Task.Commit` (BIT-46.14).
    - `dir, err := sessionDir(root)` (BIT-47.6's `cmd/session_head.go`, which returns `(string, error)`), returning the error as `mcpStore` does. Then call `landing.Check(ctx, run, landing.Query{Dir: dir, Bars: bars})`.
    - Wrap errors as `fmt.Errorf("checking landing of %s: %w", id, err)`.
  - Register the tool in `runMCPServer(ctx, root, run, transport)` (BIT-47.6) with the same `run` the other git-aware handlers get.
- Test file: `cmd/serve_mcp_test.go`. BIT-46.9 already converted it, so there's no step 0. BIT-47.2's `testGitKeepsApproval` row must stay green.

## References
- research `review-2026-10-02` §4 (the suggested `task_landing` shape), via `mcp__bit__research_read BIT-50`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskLandingHandler/a track pushed to trunk is done` (a new top-level test with subtests)
     - **Behavior:** the skill can ask whether a track's work is on trunk and get the landing commit back.
     - **Setup:** the MCP sandbox (`mcpSandbox(t)`, BIT-46.9). `r := gittest.New(t)`, and register `r.Dir` as `BIT` (`seedProject(t, orm.CreateProjectParams{Path: <canonical r.Dir>, Code: "BIT"})`). Commit `a := r.Commit("feat(bit): landing core")` and `b := r.Commit("feat(bit): task_landing tool")`, then `r.Git("push", "origin", "main")`. Through `project.OpenStore`, create track `BIT-1` and bars `BIT-1.1` (`Commit: a`) and `BIT-1.2` (`Commit: b`), both set to `done`. Then `mcpSessionWithGit(t, r.Dir, git.ExecRunner)` (BIT-47.6) and `callTool(task_landing, {id: "BIT-1"})`.
     - **Assertions:** `verdict == "done"`, `landing == b`, `trunk == "origin/main"`, `branch == "main"`, and `bars[0]` has `id == "BIT-1.1"`, `status == "done"`, `commit == a`, `class == "landed"` and `landing == a`. Assert field by field, not whole-map equality, because later bars add `repoint` to each bar.
     - **Boundary:** every bar pushed straight to trunk, the walking-skeleton happy path.
   - [ ] Confirm fails: unknown tool `task_landing`.

2. **Implement (GREEN):**
   - [ ] `landing` types and the hardcoded `Check`, the `gittest` repo fixture, then the tool, its input, handler and registration.

3. **More tests (RED → GREEN):**
   - [ ] `TestTaskLandingHandler/an unknown track is a tool error`: `{id: "BIT-9"}` gives `IsError`. *Boundary:* a missing record.
   - [ ] `TestMCPToolDescriptions/carry the domain/task_landing`: add the row `{name: taskLandingTool, tool: taskLandingTool, want: []string{testTrackSentence}}`.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none, deterministic

## Commit
`feat(bit): task_landing reports a pushed track as done`