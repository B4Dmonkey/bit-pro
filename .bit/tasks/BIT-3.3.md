---
id: BIT-3.3
title: A parent flag mints a dotted ID
status: done
phase: 2
phase_label: A step belongs to a scope
---
## Step 3 (Phase 2 — A step belongs to a scope) — A parent flag mints a dotted ID

**Status:** ✅ Done — verified 2026-07-17

`bit task create --parent BIT-2` must produce `BIT-2.1`. One child is satisfiable by
appending `.1` — that hardcode is the point, and Step 4 breaks it.

**Scope:**
- `cmd/task_create.go` — new `--parent` / `-p` flag
- `task/store.go` — new `NextChildID(parent string) (string, error)`

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskCreateCmd_ParentMintsDottedID`
     - **Behavior:** a step records which scope it belongs to, addressably — the parent
       link the README has had open since CRUD landed.
     - **Setup:** `initProject(t, "BIT")`, `createTask(t, "Track", "...")` → BIT-1, then
       `mustRun(t, "task", "create", "A step", "-d", "...", "--parent", "BIT-1")`.
     - **Assertions:** `mustRun(t, "task", "read", "BIT-1.1")` succeeds and its first
       line is `"BIT-1.1\ttodo\tA step\n"`.
     - **Boundary:** child count under BIT-1 goes 0→1 — the lower bound, first child.
   - [x] Confirm fails: unknown flag `--parent`

2. **Implement (GREEN):**
   - [x] Add `--parent` (shorthand `-p`) to `newTaskCreateCmd`
   - [x] Add `NextChildID(parent)` returning `parent + ".1"` — hardcoded on purpose
   - [x] In `RunE`: if `--parent` is set, use `NextChildID(parent)`; otherwise keep
         `NextID(cfg.Prefix)` unchanged

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**Commit (user):** `feat(create): add --parent to create a task under a scope`