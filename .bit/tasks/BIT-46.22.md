---
id: BIT-46.22
title: bp add refuses a removed project's code from another folder
status: todo
phase: 3
phase_label: bp remove soft-deletes a project
---
## **Verse 3**

A removed project keeps its code. Asking for that code from a different folder is refused before any wiring, so the operator either revives the old project or picks another code. Without this, the UNIQUE constraint would surface as an opaque `registering ...` error. A test that types a removed project's code forces the check.

## Scope
- `project/resolve.go`: `func ByCode(projects []Project, code string) (Project, bool)` (`strings.EqualFold`), and `var ErrCodeRemoved = errors.New("code belongs to a removed project")`. **Seam for BIT-49:** migrate refuses a code that belongs to another project, removed or not, through `ByCode`.
- `cmd/add.go`: after `project.ValidateCode` and before wiring, if `p, ok := project.ByCode(projects, code); ok && p.Removed`, return `fmt.Errorf("%s at %s: %w; run `bp add` there to revive it, or pick another code", p.Code, p.Path, project.ErrCodeRemoved)`. An active owner still falls through to the UNIQUE error, as in Verse 1.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAddCmd/refuses a removed project code` (subtest in `TestAddCmd`)
     - **Behavior:** a removed project's code can't be taken by a new folder.
     - **Setup:** sandbox; `bp add <tmp>/old` (`FOO`); `bp remove` there (`"y\n"`); recording runner; `bp add <tmp>/new` with stdin `"foo\n"`.
     - **Assertions:** `errors.Is(err, project.ErrCodeRemoved)`; the message contains `FOO` and the old path; the runner recorded no calls for the second add; `ListProjects` still has one row (old, removed).
     - **Boundary:** a lowercase-typed code matching a removed uppercase code, from a different path.
   - [ ] Confirm fails: the error is the wrapped UNIQUE failure (not `ErrCodeRemoved`), after wiring has run.

2. **Implement (GREEN):**
   - [ ] `ByCode`, `ErrCodeRemoved`, the check in `add`.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
Whole verse, in a fresh sandbox (set up and build as in BIT-46.11, with the fake `claude` on `PATH`):
- [ ] `cd $SB/proj && printf 'demo\n' | bp add .`; `bp task create "Open track"`; `bp task create "Step" -p DEMO-1`.
- [ ] `bp remove`, answer `n` → prints `Outstanding work:` with `DEMO-1` and `DEMO-1.1`, then `not removed`; `bp task list` still lists both.
- [ ] `bp remove`, answer `y` → `removed DEMO ...`; `ls $SB/data/bit/DEMO/archive/tasks` shows the four files; `bp list` no longer shows DEMO; `bp task list` fails with `this project was removed; run `bp add` here to revive it`.
- [ ] `mkdir $SB/other && cd $SB/other && printf 'demo\n' | bp add .` is refused, naming DEMO and `$SB/proj`.
- [ ] `cd $SB/proj && bp add .` prints `revived DEMO ...`; `bp list` shows DEMO again; `bp task create "Next"` prints `DEMO-2`.
- [ ] Whole slice: removing a project loses nothing, hides it, and `bp add` brings it back.

## Commit (user)
`feat(bit): bp add refuses a removed project's code from another folder`