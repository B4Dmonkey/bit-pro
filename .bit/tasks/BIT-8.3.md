---
id: BIT-8.3
title: task move reflected in the parent list
status: todo
phase: 1
phase_label: Resequence
---
Wire `bit task move <bar> --before|--after <anchor>` end-to-end. This is the walking skeleton's top: a reorder from the CLI shows up in `bit task list --parent` — the same ordered list bit_do resumes from — so "correct next step" is proven here.

**Scope:**
- `cmd/task_move.go` (new) — `newTaskMoveCmd`: `Use: "move <bar>"`, `Args: cobra.ExactArgs(1)`, flags `--before` and `--after` (string, anchor bar IDs). Validate exactly one is set (error if both or neither). Call `task.New(bitDir).Move(args[0], anchor, before)`.
- `cmd/task.go` — register with `taskCmd.AddCommand(newTaskMoveCmd())`.
- `assets/bit-cli.md` — document `task move` under Commands (append-appends / move-rewrites invariant; every write through `bit`).
- `assets/skills/bit_plan/SKILL.md` — in the Refine section, note that reordering an existing plan now uses `bit task move --before|--after` instead of delete-and-recreate.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestTaskMoveCmd_ReordersParentList` (cmd pkg, using `initProject`/`createTask`/`mustRun`)
     - **Behavior:** a `task move` from the CLI changes what `task list --parent` returns — the surface bit_do reads for the next step.
     - **Setup:** init `BIT`; create track `BIT-1`; create bars `BIT-1.1`, `BIT-1.2` under it; run `task move BIT-1.2 --before BIT-1.1`.
     - **Assertions:** `task list --parent BIT-1` output lines are `BIT-1.2` then `BIT-1.1`.
     - **Boundary:** two bars, N=2, moving the second ahead of the first — the smallest reorder that changes observable output.
   - [ ] `TestTaskMoveCmd_RejectsBadFlags` (table-driven)
     - **Behavior:** the command refuses ambiguous/empty positioning instead of guessing.
     - **Setup/Assertions:** both `--before` and `--after` given → error; neither given → error. No task file is modified.
     - **Boundary:** the flag cardinality constraint (exactly one), tested at both invalid extremes (zero, two).
   - [ ] Confirm fails: `unknown command "move"`.

2. **Implement (GREEN):**
   - [ ] Add `newTaskMoveCmd`, the mutual-exclusion flag check, and register it. Follow the cobra-viper conventions already used by the sibling `task_*` commands (local flag vars, `RunE`, `cmd.OutOrStdout`).
   - [ ] Update `assets/bit-cli.md` and the bit_plan skill copy.

**Claude verifies:**
- [ ] `just test` passes — including `init_test.go` (seeded `.claude` copies still byte-match the edited `assets/` sources).
- [ ] `just lint` clean.

**User verifies:**
- [ ] `just install` then `bit task move` on a real track reads naturally; the bit-cli.md wording is right.

**Commit (user):** `feat(cli): add task move to resequence a bar`