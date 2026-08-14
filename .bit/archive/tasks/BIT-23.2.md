---
id: BIT-23.2
title: bp check <ID> executes the declared check command
status: done
phase: 2
phase_label: Runnable checks
---
## **Verse 2**

The check command is the only trustworthy verdict source for a bar's work: a dispatched session returns no structured exit status, so the model's own claim "it passed" cannot be relied on. `bp check <ID>` executes the bar's declared check command and lets its exit code be the verdict.

## Scope
- `task/task.go` — add `Check string \`yaml:"check,omitempty"\`` field to `Task`
- `cmd/check.go` — new file; top-level `bp check <ID>` command that loads the task, shells out with `exec.Command("sh", "-c", t.Check)`, streams stdout/stderr, and returns the command's error (exit code propagates)
- `cmd/root.go` — add `rootCmd.AddCommand(newCheckCmd())`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheckCmd_PassingCheck`
     - **Behavior:** `bp check BIT-1` exits 0 when the declared check is a passing shell command
     - **Setup:** `initProject(t, "BIT")`; write `.bit/tasks/BIT-1.md` directly with frontmatter `check: "exit 0"` and valid `id`, `title`, `status` fields
     - **Assertions:** `mustRun(t, "check", "BIT-1")` returns no error
     - **Boundary:** check command exits 0 — the passing case; proves execution reaches the shell command
   - [ ] Confirm fails: `bp check` command does not exist; Execute returns "unknown command"
   - [ ] `TestCheckCmd_FailingCheck` (contradiction)
     - **Behavior:** `bp check BIT-1` propagates a non-zero exit to the caller
     - **Setup:** same but frontmatter `check: "exit 1"`
     - **Assertions:** `run(t, "check", "BIT-1")` returns a non-nil error
     - **Boundary:** check command exits non-zero — the failing case; forces real `exec.Command` execution rather than hardcoded nil return

2. **Implement (GREEN):**
   - [ ] Add `Check string \`yaml:"check,omitempty"\`` to `Task` struct in `task/task.go`
   - [ ] Create `cmd/check.go` with `newCheckCmd()` that runs `exec.Command("sh", "-c", t.Check).Run()`; if `t.Check == ""`, return `fmt.Errorf("no check declared for %s", id)`
   - [ ] Stream the command's output by wiring its Stdout/Stderr to `cmd.OutOrStdout()`/`os.Stderr`
   - [ ] Wire into root: `rootCmd.AddCommand(newCheckCmd())`

3. **More tests (RED → GREEN):**
   - [ ] `TestCheckCmd_NoCheckDeclared`
     - **Behavior:** `bp check` on a task with no check field returns an error
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Track", "...")` (no check field)
     - **Assertions:** `run(t, "check", "BIT-1")` returns non-nil error; error message contains "no check declared"
     - **Boundary:** check field absent — proves the empty guard

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] `bp check BIT-23.1` (once BIT-23.1's check is declared via Bar 3) — command runs and exits 0 when the check passes, exits non-zero and shows output when it fails

## Commit (user)
`feat(check): add bp check command that executes a bar's declared check`