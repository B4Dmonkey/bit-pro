---
id: BIT-9.2
title: Bare enter reuses the existing prefix
status: done
phase: 2
phase_label: Re-run init keeps prefix
---
Re-running `bit init` in an initialized project lets a bare enter reuse the existing prefix, instead of erroring "cannot be empty." Forces reading the existing config and falling back to its prefix when the prompt input is empty.

**Scope:**
- `cmd/init.go` — in the interactive branch, before prompting, read the existing config (`task.New(bitDir).Config()`); a present config's `Prefix` becomes the fallback, a missing/unreadable config yields no fallback. After trimming input, if `prefix == ""` and a fallback exists, use the fallback. The final `if prefix == "" { error }` still fires when there was no existing config (fresh project).
- `cmd/init_test.go` — new test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestInitCmd_ReuseExistingPrefixOnEnter`
     - **Behavior:** a bare enter in an already-initialized project keeps the existing prefix rather than erroring.
     - **Setup:** `t.Chdir(t.TempDir())`, seed a config first with `mustRun(t, "init", "--prefix", "BIT")`, then `runWithStdin(t, "\n", "init")` (bare enter, no `--prefix`).
     - **Assertions:** the second run returns no error; `task.New(".bit").Config()` still reports `Prefix == "BIT"`.
     - **Boundary:** empty prompt input *with* a pre-existing config — the case today's `prefix == ""` guard rejects. (Its complement, empty input with no config, stays an error — guarded by the existing `TestInitCmd_RejectsBadInvocations`.)
   - [ ] Confirm fails: today `"\n"` trims to `""` and `init` returns `errors.New("task ID prefix cannot be empty")`, so `runWithStdin` returns a non-nil error and the "no error" assertion fails for the right reason.

2. **Implement (GREEN):**
   - [ ] Before the prompt, load the existing prefix into a local (e.g. `existing string`): call `task.New(bitDir).Config()`, and on success set `existing = cfg.Prefix`; on error (missing config) leave it empty. After `prefix = strings.TrimSpace(line)`, add `if prefix == "" { prefix = existing }`. The existing empty-check stays and errors only when `existing` was also empty.

**Claude verifies:**
- [ ] `just test` passes — including the new test and the existing `TestInitCmd_RejectsBadInvocations` (fresh + empty still errors) and `TestInitCmd_IsIdempotent`.
- [ ] `just lint` clean.

**User verifies:**
- [ ] In an already-initialized project, run `bit init` and press enter at the prompt — it completes and keeps the existing prefix.

**Commit (user):** `feat(init): reuse the existing prefix on a bare enter`