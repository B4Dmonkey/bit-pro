---
id: BIT-10.7
title: Delete honors the guard and --force
status: done
phase: 2
phase_label: Non-destructive delete
---
Give delete a `--force` flag so it can override the all-done guard, matching archive. The guard itself already reaches delete for free (both call `Relocate`); this bar wires the override.

**Scope:**
- `cmd/task_delete.go` — add `--force`/`-f`; `RunE` calls `Relocate(id, force)`.
- `cmd/task_delete_test.go` — new tests.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestTaskDeleteCmd_ForceDeletesUnfinished`
     - **Behavior:** `--force` deletes (relocates) a track with an unfinished bar.
     - **Setup:** `createTask` BIT-1; create bar `BIT-1.1` `--parent BIT-1` (todo); `mustRun(t,"task","delete","BIT-1","--yes","--force")`.
     - **Assertions:** both `BIT-1.md` and `BIT-1.1.md` in `archive/`; none in `tasks/`.
     - **Boundary:** `force=true` — the override.
   - [ ] `TestTaskDeleteCmd_RefusesUnfinishedWithoutForce` (lock-in, passes once 2.1's shared `Relocate` is in — guards against regression, not a fresh RED)
     - **Behavior:** without `--force`, deleting a track with a todo bar refuses and moves nothing.
     - **Setup:** same track+bar; `run(t,"task","delete","BIT-1","--yes")`.
     - **Assertions:** err is / wraps `*UnfinishedBarsError` naming `BIT-1.1`; both files still in `tasks/`.
     - **Boundary:** `force=false`, unfinished.
   - [ ] Confirm fails: delete hardcodes `force=false` → the `--force` test still errors.

2. **Implement (GREEN):**
   - [ ] Add `--force`/`-f`; pass it to `Relocate`. (`--yes` skips the prompt; `--force` overrides the guard — orthogonal.)

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)
- [ ] `just build` then `bit task delete --help` shows `--force`

**User verifies:**
- [ ] the guard's refusal message reads clearly

**Commit (user):** `feat(cmd): task delete honors the all-done guard and --force`