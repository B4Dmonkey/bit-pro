---
id: BIT-29.4
title: A project with a bit prefix enrolls on one keypress
status: todo
approved: true
phase: 2
phase_label: add
---
## **Verse 2**

`bp add .` in a project that already has a `.bit` prefix enrolls it on one keypress: the prompt
offers the prefix as the default, Enter accepts it, and the row lands in the registry. The walking
skeleton for this verse — BIT-29.1 and BIT-29.2 built a registry with no way to put anything in it
from the CLI.

## Scope
- `cmd/add.go` (new) — the command, the prompt helper, the insert
- `cmd/root.go` — register it
- `cmd/add_test.go` (new) — the test

The `.bit` directory is read as `filepath.Join(abs, ".bit")` off the resolved argument, not from the
package-level `bitDir`. `bitDir` is derived from the process working directory (and remapped for
`.claude/worktrees` checkouts), which is right for a command with no path argument and wrong for one
that takes one. Enrolling from inside a worktree is not a case this track handles — the scope already
restricts `bp add` to being run from inside the project it names.

Args are `cobra.ExactArgs(1)`: the scope writes the command as `bp add <path>` throughout, and
`bp add .` is the invocation its Decisions are built around.

## TDD cycle

1. **Write test (RED):** `cmd/add_test.go`
   - [ ] `TestAddCmd_EnrollsUsingTheBitPrefix`
     - **Behavior:** a project bit already tracks needs no new information to be enrolled — the
       prefix it was initialised with is the code, and accepting it is one keypress. This is the
       common case, so it has to be the cheapest one.
     - **Setup:** `home := t.TempDir()`; `t.Setenv("HOME", home)`; `t.Setenv("XDG_DATA_HOME", "")`;
       then `initProject(t, "BIT")` for a project with a `config.toml`. Compute the expected path as
       `want, err := filepath.Abs(".")` **after** the chdir — not from `initProject`'s return value,
       which on macOS is a `/var/folders/…` path while `os.Getwd()` reports the resolved
       `/private/var/…` one. The implementation calls `filepath.Abs` too, so computing it the same
       way is what makes the assertion meaningful rather than platform trivia.
       Run `out, err := runWithStdin(t, "\n", addCmdUse, ".")`.
     - **Assertions:** `err` is nil. `out == "Project code (BIT): added BIT " + want + "\n"` — exact
       equality, which pins the prompt's parenthesised-default form and the success line's
       code-then-path order in one assertion, both scope Decisions. Then reopen the registry with
       `db.Open()` and assert `orm.New(sqlDB).ListProjects(t.Context())` has length 1 with
       `Code == "BIT"` and `Path == want`.
     - **Boundary:** the prompt's default-present state — one of its two states, the other being
       BIT-29.7's. And registry row count 0 → 1.
   - [ ] Confirm fails: `Error: unknown command "add" for "bp"`.

2. **Implement (GREEN):**
   - [ ] `cmd/add.go`: `const addCmdUse = "add"` and `newAddCmd() *cobra.Command` with
         `Use: "add <path>"`, `Args: cobra.ExactArgs(1)`, and a `Short`. No `claude.Runner`
         parameter yet — BIT-29.7 is what needs one.
   - [ ] `cmd/add.go`: `readProjectCode(cmd *cobra.Command, existing string) (string, error)`,
         mirroring `readInteractivePrefix` in `cmd/init.go` — same `bufio.NewReader(cmd.InOrStdin())`
         read, same `err != nil && line == ""` guard, same fall back to `existing` on empty. It takes
         `existing` as a parameter rather than looking it up, so BIT-29.7 can pass `""`.
   - [ ] `cmd/add.go`: in `RunE` — `abs, err := filepath.Abs(args[0])`; read the prefix with
         `task.New(filepath.Join(abs, ".bit")).Config()`, tolerating the error the way
         `readInteractivePrefix` does (`if cfg, err := …; err == nil`); prompt with
         `readProjectCode`; open the registry with `db.Open()` and `defer sqlDB.Close()`; insert via
         `orm.New(sqlDB).CreateProject(cmd.Context(), orm.CreateProjectParams{Path: abs, Code: code})`;
         print `added %s %s\n`.
   - [ ] `cmd/root.go`: `rootCmd.AddCommand(newAddCmd())`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. Verse 2's whole-slice check lands on BIT-29.8.

## Commit (user)
`feat(add): enroll a project using its bit prefix as the code`