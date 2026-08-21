---
id: BIT-29.8
title: An empty code is refused before anything is written
status: done
phase: 2
phase_label: add
---
## **Verse 2**

An empty project code is refused and nothing is written — no Claude wiring, no registry row.
Contradicts BIT-29.7, which reaches the no-default prompt for the first time and so is the first bar
where pressing Enter yields `""` with nothing to fall back on.

## Scope
- `cmd/add.go` — the guard
- `cmd/add_test.go` — the test

This is the one case the scope did not settle when it was written; the behaviour and the message were
decided during planning and belong in the track's Decisions.

The guard sits immediately after the read and before every write, which is where `bp init` puts its
equivalent — `TestInitCmd_RejectsBadInvocations` asserts the same property from the other side, that
a rejected prompt leaves no `config.toml` behind.

## TDD cycle

1. **Write test (RED):** `cmd/add_test.go`
   - [ ] `TestAddCmd_RejectsAnEmptyCode`
     - **Behavior:** a project with no code is not a project the registry can name, so the command
       refuses rather than storing a blank one — and it refuses before touching anything, so a
       mistyped enrollment leaves the directory exactly as it was.
     - **Setup:** `home := t.TempDir()`; `t.Setenv("HOME", home)`; `t.Setenv("XDG_DATA_HOME", "")`;
       `t.Chdir(t.TempDir())` — a bare directory, so the prompt has no default. Run
       `runWithStdin(t, "\n", addCmdUse, ".")`.
     - **Assertions:** `err` is non-nil and `err.Error() == "project code cannot be empty"`.
       `os.Stat(".claude/settings.json")` returns an error satisfying `errors.Is(err,
       fs.ErrNotExist)` — the wiring never ran. `ListProjects` returns zero rows.
     - **Boundary:** code length at 0 — the lower bound of the valid range, and the only value the
       registry must refuse.
   - [ ] Confirm fails: `err` is nil. BIT-29.7 wires the directory and inserts a row whose `Code` is
         `""`, printing `added  <abs>` with the doubled space where the code should be.

2. **Implement (GREEN):**
   - [ ] `cmd/add.go`: immediately after `readProjectCode` returns and before anything is written,
         `if code == "" { return errors.New("project code cannot be empty") }`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice, in a real directory that has no `.bit/`: run `bp add .`, press Enter at the
      prompt — the command refuses and the directory is untouched (no `.claude/`). Run it again and
      type a code — the prompt offers no default, the plugin sync runs, and the last line reads
      `added <CODE> <abspath>`. Run it a third time — it prints `already added` with no prompt. Then
      in a project that already has `.bit/config.toml`, `bp add .` and press Enter — the prompt
      offers the prefix, and one keypress enrolls it. That sequence is the capability this verse
      exists for: any project, tracked or not, is one command away from being registered.

## Commit (user)
`feat(add): refuse an empty project code before writing anything`