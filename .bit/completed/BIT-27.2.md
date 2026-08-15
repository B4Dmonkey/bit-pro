---
id: BIT-27.2
title: Nesting contradicts a last-occurrence cut
status: done
phase: 1
phase_label: Canonical store
---
## **Verse 1**

A worktree created inside a worktree must resolve to the *outermost* checkout. This contradicts
the previous bar: a `strings.LastIndex`-style search for `.claude/worktrees` passes the single-
worktree test and lands on the inner checkout here, so only cutting at the **first** occurrence
satisfies both.

## Scope
- `cmd/root_test.go` — the contradicting test.
- `cmd/root.go` — the helper added in the previous bar, if it doesn't already cut at the first
  occurrence.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestBitDir_NestedWorktreeResolvesToOutermostCheckout`
     - **Behavior:** the derivation cuts at the first `.claude/worktrees/` segment, so the
       canonical store is the true repo root even when worktrees nest — the store never lands
       on an intermediate copy that is itself a branch snapshot.
     - **Setup:** `root := initProject(t, "BIT")`; `createTask(t, "Track", "...")` so `BIT-1`
       exists in `<root>/.bit`. Build
       `nested := filepath.Join(root, ".claude", "worktrees", "outer", ".claude", "worktrees",
       "inner")` and `os.MkdirAll(filepath.Join(nested, ".bit"), 0o755)`. Also create the
       decoy the naive implementation would pick:
       `os.MkdirAll(filepath.Join(root, ".claude", "worktrees", "outer", ".bit"), 0o755)`.
       Then `t.Chdir(nested)`.
     - **Assertions:** `mustRun(t, "task", "list")` output contains `BIT-1`.
     - **Boundary:** count of `.claude/worktrees/` segments in cwd == 2 — the first value above
       the single-occurrence case the previous bar pinned, and the one that distinguishes a
       first-occurrence cut from a last-occurrence cut.
   - [ ] Confirm fails **only if** the previous bar's helper searched from the right; if it
     already cut at the first occurrence this test is green on arrival. That is a real outcome,
     not a skipped step — record it and keep the test, since it is what pins the behaviour
     against a later refactor. Expected failure reason when it does fail: output missing
     `BIT-1`, because resolution landed on the empty `<root>/.claude/worktrees/outer/.bit` decoy.

2. **Implement (GREEN):**
   - [ ] Only if RED actually failed: change the helper's segment search to return the first
     matching index rather than the last. No other change.

## Claude verifies
- [ ] `just test` passes — both worktree tests and `TestBitDirEnvVar_DefaultIsRelativeDotBit`
- [ ] `just lint` passes
- [ ] `just install`

## User verifies
- [ ] none — deterministic; the verse's end-to-end check lands on its last bar

## Commit (user)
`fix(cmd): cut at the first .claude/worktrees segment so nesting resolves outermost`