---
id: BIT-10.8
title: Force overrides the guard
status: todo
phase: 1
phase_label: Archive
---
Let `force` bypass the guard for the deliberate case (abandoning in-progress work). Contradicts bar 1.3, whose guard ignores `force` and still errors.

**Scope:**
- `task/store.go` — gate the unfinished-bars guard behind `!force`.
- `task/store_test.go` — new test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreRelocate_ForceOverridesGuard`
     - **Behavior:** `force=true` relocates a track even with unfinished bars.
     - **Setup:** same unfinished setup as 1.3; `s.Relocate("BIT-1", true)`.
     - **Assertions:** all three in `archive/`; none in `tasks/`; no error.
     - **Boundary:** `force=true` against the exact input 1.3 rejected — the override case.
   - [ ] Confirm fails: 1.3's guard ignores `force` → still returns `UnfinishedBarsError`.

2. **Implement (GREEN):**
   - [ ] Run the guard only when `!force`.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**Commit (user):** `feat(task): --force relocate past the unfinished guard`