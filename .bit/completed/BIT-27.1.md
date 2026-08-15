---
id: BIT-27.1
title: Cwd inside a Claude worktree resolves to the main checkout
status: done
phase: 1
phase_label: Canonical store
---
## **Verse 1**

`bp` derives the canonical `.bit/` from its own working directory: when cwd sits inside
`.claude/worktrees/<slug>`, everything from `.claude/` onward is cut off and the store resolves
against the main checkout. The existing `TestBitDirEnvVar_DefaultIsRelativeDotBit` is the
contradiction that stops "always cut" from passing — a cwd with no `.claude/worktrees/` segment
must still resolve to the relative `.bit`.

## Scope
- `cmd/root.go` — add the derivation to the `PersistentPreRunE` resolution point. The existing
  `--bit-dir` / `BIT_DIR` branches stay in place for now and still win; they come out in a later
  bar, so this bar leaves every current test green.
- `cmd/root_test.go` — the new RED test.

## References
- `automation-notes.md`, section **Measured facts → "Worktrees are imposed, not opted into"** —
  the observed path convention this derivation keys on (`<repo>/.claude/worktrees/<random-slug>`,
  branch `worktree-<slug>`, session cwd set to it). Read it before choosing what to match on.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestBitDir_InsideClaudeWorktreeResolvesToMainCheckout`
     - **Behavior:** a `bp` command run from inside a Claude Code worktree reads the main
       checkout's store, not the worktree's divergent copy — the whole point of the fix.
     - **Setup:** `root := initProject(t, "BIT")`; `createTask(t, "Track", "...")` so `BIT-1`
       exists in `<root>/.bit`. Then `os.MkdirAll(filepath.Join(root, ".claude", "worktrees",
       "hazy-pondering-star", ".bit"), 0o755)` — a bare empty `.bit` standing in for the
       worktree's snapshot copy. Then `t.Chdir(filepath.Join(root, ".claude", "worktrees",
       "hazy-pondering-star"))`. Do **not** set `BIT_DIR` or pass `--bit-dir`.
     - **Assertions:** `mustRun(t, "task", "list")` output contains `BIT-1`.
     - **Boundary:** cwd depth relative to the `.claude/worktrees/` segment == 0 (cwd *is* the
       worktree root, the exact shape Claude Code produces) — the lower bound of the
       inside-a-worktree case, and the one the operator actually hits.
   - [ ] Confirm fails: output does not contain `BIT-1` — resolution used the relative `.bit`,
     which is the empty directory the setup created (the run may instead error on the
     configless store; either failure is the same missing-derivation cause).

2. **Implement (GREEN):**
   - [ ] In `cmd/root.go`, add an unexported helper that takes a working directory and returns
     the store path: split it on the OS separator, find the **first** index where a segment is
     `.claude` immediately followed by `worktrees`, and return `filepath.Join(<everything before
     that index>..., ".bit")`. No match → return `".bit"` unchanged. Reuse the existing
     `claudeDir` constant for the `.claude` segment.
   - [ ] Call it from `PersistentPreRunE` as the new default, before the flag/env branches:
     `wd, err := os.Getwd()`; on error leave `bitDir = ".bit"` and return no error — today's
     behaviour is preserved and no new failure path is introduced; otherwise
     `bitDir = <helper>(wd)`. The `bitDirFlag` and `BIT_DIR` branches still override it.
   - [ ] Match on path **segments**, not a substring — a directory literally named
     `my.claude/worktrees-old` must not match.

Note for the implementer: assert on behaviour (`BIT-1` in the output), never on the resolved
path string. On macOS `t.TempDir()` hands back `/var/folders/...` while `os.Getwd()` returns the
physical `/private/var/...`; segment-cutting works identically on either, but a string
comparison against the value `initProject` returned would fail for an unrelated reason.

## Claude verifies
- [ ] `just test` passes — including `TestBitDirEnvVar_DefaultIsRelativeDotBit`, which is the
  contradiction proving the derivation didn't swallow the non-worktree case
- [ ] `just lint` passes
- [ ] `just install`

## User verifies
- [ ] none — deterministic; the verse's end-to-end check lands on its last bar

## Commit (user)
`fix(cmd): resolve .bit from the main checkout when run inside a Claude worktree`