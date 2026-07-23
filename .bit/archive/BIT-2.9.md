---
id: BIT-2.9
title: Contradiction forces path-traversal-safe `taskPath`
status: done
phase: 2
phase_label: List & read
---
## Step 9 (Phase 2 — List & read) — Contradiction forces path-traversal-safe `taskPath`
**Status:** ✅ Done — verified 2026-07-16
Step 8's `taskPath(id)` does `filepath.Join(tasksDir, id+".md")` with no containment.
`read`/`update`/`delete` all take `id` straight from `args[0]` — user-typed, untrusted —
so a crafted ID like `../../README` resolves outside `.bit/tasks/` entirely (`filepath.Join`
cleans `..` but doesn't stop it escaping the root), landing on the project's real
`README.md`. It has to be a real `.md` file specifically: `taskPath` always appends
`.md` to `id`, so a shallower escape like `../config` (one level up, to `.bit/`) would
only ever resolve to a nonexistent `config.md` — never the real `.bit/config.toml` —
which would make the RED test pass for the wrong reason (file-not-found) even against
the unguarded implementation. A test supplying the `../../README` ID and asserting the
command refuses rather than touching the file it resolves to can't pass against the
current unguarded `filepath.Join`, forcing real containment. This matters before Phase
4's `delete` can be trusted not to remove a file outside `.bit/tasks/` on a crafted or
badly-mistyped ID — the exact class of accident the scope's confirmation requirement
exists to prevent, which a path escape sidesteps entirely.

**Scope:**
- `go.mod` / `go.sum` — add `github.com/spf13/pathologize@v1.1.0`
- `cmd/task_model.go` — `taskPath` uses `pathologize.Join` to contain `id` under
  `tasksDir` instead of a raw `filepath.Join`
- `cmd/task_read_test.go` — new test

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskReadCmd_RejectsPathTraversalID` (in `cmd/task_read_test.go`)
     - **Behavior:** proves a task ID can't be used to read a file outside
       `.bit/tasks/` — `README.md` at the project root is a concrete, real `.md` file
       two `..` segments away from `tasksDir`.
     - **Setup:** `dir := t.TempDir()`; `t.Chdir(dir)`; write a fixture
       `os.WriteFile(filepath.Join(dir, "README.md"), []byte("# real project readme\n"),
       0o644)` (the temp dir has no repo README, so the test supplies its own target);
       `init --prefix BIT`; `task create "Real task" --description "..."`; fresh
       `rootCmd` with `SetOut(buf)`, `SetArgs([]string{"task", "read", "../../README"})`.
     - **Assertions:** `err` is not `nil`; `buf.String()` does not contain `"real
       project readme"` (the content that would appear if `README.md` were printed);
       `os.ReadFile(filepath.Join(dir, "README.md"))` still succeeds and returns the
       unchanged fixture content (the read attempt didn't corrupt or move anything).
     - **Boundary:** `id` containing `..` that resolves two levels above `tasksDir` —
       escaping both `.bit/tasks/` and `.bit/` to reach the project root. One level
       (`../config`, landing in `.bit/`) isn't a viable escape target here: `taskPath`
       always appends `.md`, so it would resolve to a nonexistent `config.md` rather
       than the real `.bit/config.toml`, and the test would pass for the wrong reason.
   - [x] Confirm fails: `err` is `nil` and `buf.String()` contains the fixture's raw
     content (`"real project readme"`) — `filepath.Join(tasksDir, "../../README"+".md")`
     cleans to `README.md` at the project root, which `os.ReadFile` happily opens.

2. **Implement (GREEN):**
   - [x] `go get github.com/spf13/pathologize@v1.1.0`
   - [x] `cmd/task_model.go`: add `"github.com/spf13/pathologize"` to imports; replace
     `taskPath`:
     ```go
     func taskPath(id string) string {
         return pathologize.Join(tasksDir, id+".md")
     }
     ```
     `pathologize.Join` sanitizes `id` and guarantees the result stays lexically under
     `tasksDir` — a `..`-laden ID can no longer escape it (see the `fileflow-pathologize`
     skill: `tasksDir` is the trusted root, `id` is the untrusted part).
   - [x] Since a contained, sanitized ID no longer matches a real file on disk,
     `loadTask`'s existing `os.ReadFile` failure naturally produces a "no such file"
     error for a traversal attempt — no separate validation branch needed.

**Claude verifies:**
- [x] `go test ./...` passes — Steps 4–8's tests stay green (legitimate IDs like
  `BIT-1` are untouched by `pathologize.Join`)
- [x] `go vet ./...` passes

**User verifies:**
- [x] none — this closes a gap the scope didn't call out explicitly, but it's a direct
  consequence of Phase 4's "warns/confirms before removing anything" requirement: a path
  escape would let a crafted ID bypass that protection entirely

**Commit (user):** `fix(task-crud): contain task ID file paths under .bit/tasks/`