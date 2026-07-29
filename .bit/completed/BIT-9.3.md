---
id: BIT-9.3
title: Prompt shows the existing prefix as default
status: done
phase: 2
phase_label: Re-run init keeps prefix
---
The prompt tells the human what a bare enter will reuse — an already-initialized project shows `Task ID prefix (BIT): `, a fresh one still shows `Task ID prefix: `. Forces composing the prompt string from the existing prefix (bar 1 read it and fell back silently; this makes the affordance visible), and contradicts a naive always-`(…)` prompt that would render `Task ID prefix (): ` in a fresh project.

**Scope:**
- `cmd/init.go` — build the prompt text from the `existing` prefix read in the prior step: when non-empty, `fmt.Fprintf(cmd.OutOrStdout(), "Task ID prefix (%s): ", existing)`; when empty, the current `"Task ID prefix: "`.
- `cmd/init_test.go` — new table-driven test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestInitCmd_PromptShowsExistingPrefix` (table-driven, two cases)
     - **Behavior:** the prompt offers the existing prefix as a default only when one exists.
     - **Setup:** case "existing config" — seed with `mustRun(t, "init", "--prefix", "BIT")`, then `out, _ := runWithStdin(t, "BIT\n", "init")`. case "fresh project" — no seed, `out, _ := runWithStdin(t, "BIT\n", "init")` in a fresh `t.TempDir()`. (Both feed `"BIT\n"` so init completes without blocking; the assertion is on the printed prompt, not the outcome.)
     - **Assertions:** existing-config case — `strings.Contains(out, "Task ID prefix (BIT): ")`. fresh case — `strings.Contains(out, "Task ID prefix: ")` and `!strings.Contains(out, "(")`.
     - **Boundary:** prompt composition at both config states — present (renders the `(BIT)` default) vs absent (no parens). The absent case is what rules out an always-`(…)` bug.
   - [ ] Confirm fails: today the prompt is the literal `"Task ID prefix: "` in both cases, so the existing-config assertion (`"Task ID prefix (BIT): "`) fails for the right reason.

2. **Implement (GREEN):**
   - [ ] Replace the fixed `fmt.Fprint(..., "Task ID prefix: ")` with a conditional on `existing`: non-empty → `fmt.Fprintf(..., "Task ID prefix (%s): ", existing)`; empty → the unchanged `"Task ID prefix: "`.

**Claude verifies:**
- [ ] `just test` passes — new test plus the Phase 2 bar-1 test and the existing init tests.
- [ ] `just lint` clean.

**User verifies:**
- [ ] In an initialized project, `bit init` shows `Task ID prefix (BIT): `; in a fresh directory it still shows `Task ID prefix: `.

**Commit (user):** `feat(init): show the existing prefix as the prompt default`