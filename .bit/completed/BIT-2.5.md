---
id: BIT-2.5
title: Contradiction forces real ID assignment
status: done
phase: 1
phase_label: Init wizard + create
---
## Step 5 (Phase 1 — Init wizard + create) — Contradiction forces real ID assignment
**Status:** ✅ Done — verified 2026-07-15
Step 4 hardcodes `id := cfg.Prefix + "-1"`, so a second `create` call would silently
overwrite `BIT-1.md`. A test seeding **non-contiguous** existing task files (`BIT-1` and
`BIT-3`, no `BIT-2`) forces real scanning — and specifically rules out a naive
"count existing files + 1" approach, which would produce `BIT-3` and collide with the
existing `BIT-3.md`.

**Scope:**
- `cmd/task_create.go` — replace the hardcoded ID with a real scan
- `cmd/task_create_test.go` — new test

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskCreateCmd_AssignsNextIDWhenTasksExist`
     - **Behavior:** proves task IDs are assigned by scanning existing files for the
       highest number, not by counting them — the scope's Risks section calls this
       exact ambiguity out as needing a decision in this plan.
     - **Setup:** `init --prefix BIT`; manually write `.bit/tasks/BIT-1.md` and
       `.bit/tasks/BIT-3.md` (minimal valid frontmatter, any body); run
       `task create "Third real task" --description "..."`.
     - **Assertions:** `err` is `nil`; `.bit/tasks/BIT-4.md` exists (`max(1,3)+1`);
       `BIT-1.md` and `BIT-3.md` are unchanged (still exist, original content intact).
     - **Boundary:** existing IDs `{1, 3}` — a gap at 2 — proves max-based assignment,
       not count-based (count would wrongly produce `3`, colliding with `BIT-3.md`).
   - [x] Confirm fails: `task create` overwrites `.bit/tasks/BIT-1.md` instead of
     creating `BIT-4.md` — Step 4's hardcoded `-1` suffix.

2. **Implement (GREEN):**
   - [x] `cmd/task_create.go`: add `nextTaskID(prefix string) (string, error)`:
     ```go
     func nextTaskID(prefix string) (string, error) {
         matches, err := filepath.Glob(filepath.Join(tasksDir, prefix+"-*.md"))
         if err != nil {
             return "", fmt.Errorf("scanning %s for existing task IDs: %w", tasksDir, err)
         }
         re := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `-(\d+)\.md$`)
         max := 0
         for _, m := range matches {
             if sub := re.FindStringSubmatch(filepath.Base(m)); sub != nil {
                 if n, _ := strconv.Atoi(sub[1]); n > max {
                     max = n
                 }
             }
         }
         return fmt.Sprintf("%s-%d", prefix, max+1), nil
     }
     ```
   - [x] Replace the hardcoded ID in `RunE` with `nextTaskID(cfg.Prefix)`.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] max-based (not count-based) ID assignment is the right call — it means a manually
  deleted task's ID is never reused, which is probably desirable but worth confirming

**Commit (user):** `feat(task-crud): assign task IDs by scanning for the highest existing number`