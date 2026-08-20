---
id: BIT-29.7
title: A project with no .bit gets wired, then enrolled
status: todo
approved: true
phase: 2
phase_label: add
---
## **Verse 2**

A directory bit has never touched gets wired for Claude and then enrolled — the init flow minus
`config.toml`. Contradicts BIT-29.4 through BIT-29.6, all of which assume a project that `bp init`
has already been run in, and none of which can produce a `.claude/settings.json`.

## Scope
- `cmd/init.go` — extract the Claude-wiring half of `bp init` into a helper both commands call
- `cmd/add.go` — call it when `.bit` is absent; take a `claude.Runner`
- `cmd/root.go` — pass the runner through
- `cmd/add_test.go` — the test

"Runs the init flow minus `config.toml` creation" is taken literally: the settings write, the
`Bringing the bit plugin current...` line, and `claude.SyncPlugin` come out of `bp init` as one
helper that both commands call, rather than being reassembled in `cmd/add.go`. Reassembling them is
where the two flows would drift, and it would also make that progress line a wording choice this
plan has no business making — extracting it keeps it `bp init`'s string, unchanged.

`claude.SyncPlugin` shells `claude plugin update bit@bit-pro --scope project`, and `--scope project`
resolves against the process working directory — there is no `--cwd` flag on either
`claude plugin install` or `claude plugin update`. That is the whole reason the scope restricts this
path to `bp add .`; running it against some other directory would wire the plugin into the wrong
project, and the scope puts that explicitly out of this track.

## TDD cycle

1. **Write test (RED):** `cmd/add_test.go`
   - [ ] `TestAddCmd_InitialisesAProjectWithoutBit`
     - **Behavior:** enrolling a project the daemon will drive also makes it a project Claude can be
       driven *in* — the operator runs one command, not `bp init` followed by `bp add`. And it stops
       short of writing `config.toml`, because the task prefix is `bp init`'s to own.
     - **Setup:** `home := t.TempDir()`; `t.Setenv("HOME", home)`; `t.Setenv("XDG_DATA_HOME", "")`;
       `t.Chdir(t.TempDir())` — a bare directory, no `bp init`. Collect the shelled commands with a
       recording runner, as `TestInitCmd_SyncsThePlugin` does. `want, err := filepath.Abs(".")` after
       the chdir. Run `runWithRunner(t, run, "BIT\n", addCmdUse, ".")`.
     - **Assertions:** `err` is nil, and
       `out == "Project code: Bringing the bit plugin current...\nadded BIT " + want + "\n"` — exact,
       which pins the no-default prompt form (a scope Decision) and the order the flow runs in.
       Separately assert `!strings.Contains(out, "(")`, the way
       `TestInitCmd_PromptShowsExistingPrefix` does for the same property. `.claude/settings.json`
       exists and contains `bit@bit-pro`. `os.Stat(".bit/config.toml")` returns an error satisfying
       `errors.Is(err, fs.ErrNotExist)`. The recorded calls equal the two-element slice
       `TestInitCmd_SyncsThePlugin` already asserts. `ListProjects` has one row with `Code == "BIT"`.
     - **Boundary:** the presence of `.bit/` — the absent state, the other side of the branch
       BIT-29.4 through BIT-29.6 all exercise on the present side. And the prompt's no-default
       state, the second of its two.
   - [ ] Confirm fails: the prompt assertions pass already — BIT-29.4's config read tolerates a
         missing file, so `existing` is `""` and the no-paren form is what prints. The failures are
         the ones that matter: `out` has no `Bringing the bit plugin current...` line,
         `.claude/settings.json` does not exist, and the recorded calls are empty because
         `newAddCmd` has no runner to call.

2. **Implement (GREEN):**
   - [ ] `cmd/init.go`: extract
         `writeClaudeWiring(cmd *cobra.Command, run claude.Runner, dir string) error` — the
         `claude.WriteSettings(filepath.Join(dir, claudeDir, "settings.json"))` call, the
         `Bringing the bit plugin current...` line, and `claude.SyncPlugin`. `newInitCmd` calls it
         with `"."`; `filepath.Join(".", ".claude", "settings.json")` is `.claude/settings.json`, so
         `bp init`'s behaviour and its tests are unchanged.
   - [ ] `cmd/add.go`: `newAddCmd(run claude.Runner)`, and `cmd/root.go`:
         `rootCmd.AddCommand(newAddCmd(run))` — the same runner `newInitCmd` already receives.
   - [ ] `cmd/add.go`: after the code is read and normalised, `os.Stat(filepath.Join(abs, ".bit"))`;
         when it fails with `errors.Is(err, fs.ErrNotExist)`, call `writeClaudeWiring(cmd, run, abs)`
         before inserting. No `task.SaveConfig` call — that is the "minus `config.toml`" half.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. Verse 2's whole-slice check lands on BIT-29.8.

## Commit (user)
`feat(add): wire Claude into a project that has no .bit yet`