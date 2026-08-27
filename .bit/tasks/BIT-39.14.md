---
id: BIT-39.14
title: bp add makes a project ready even when it is already added
status: todo
---
## Why

The operator's flow is meant to be three steps: install `bp` once, `bp add` a project, then work
with Claude. Step two has to make the project *ready* — every dependency the daemon will need
present, and present in **that** project. Today it does not, and the failure is silent: `bp add`
prints success, and the gap only shows up as a dispatched session dying with `--agent
'bit:bot-dev' not found`.

Measured 2026-08-26 against `tools/example`, which had been added and still could not resolve the
agent. The install record shows the plugin pinned somewhere else entirely:

```
bit@bit-pro | project | /Users/appstack/Developer/UniqueDataManagement/tools/bit-pro
```

## The problem

Three defects compound. Any one of them alone would leave a project unready.

**1. `bp add` short-circuits on an enrolled project.** `cmd/add.go:45-48`:

```go
if enrolled {
    fmt.Fprintln(cmd.OutOrStdout(), "already added")
    return nil
}
```

Re-running `bp add .` is the natural way to repair a project, and it is exactly the thing that
does nothing. Enrollment is a row in the registry; readiness is everything else. Conflating them
means the setup can only ever run once, on the one day the project happened not to exist yet.

**2. The wiring is gated on `.bit/` being absent.** `cmd/add.go:66-70`:

```go
if _, err := os.Stat(filepath.Join(abs, ".bit")); errors.Is(err, fs.ErrNotExist) {
    if err := writeClaudeWiring(cmd, run, abs); err != nil {
```

So a project that ran `bp init` first — which is the documented order — never gets the wiring from
`bp add` at all. The two commands each assume the other did it.

**3. The install lands in the wrong project.** `claude.SyncPlugin` runs `claude plugin install
bit@bit-pro --scope project`, and `--scope project` resolves against the **current working
directory**. `claude.Runner` has no directory parameter and `ExecRunner` never sets `cmd.Dir`, so
the install inherits wherever the operator was standing — not the `abs` path `bp add` was given.
`RegisterMCP`'s `claude mcp add` has the same property: local scope is per-project, keyed on cwd.

There is a fourth, subtler one worth pricing while the file is open: `SyncPlugin` tries `plugin
update` first and only falls back to `install` if the update fails. An update of a plugin
installed for *some other* project can succeed, which swallows the fallback and leaves this
project without it.

## Scope
- `cmd/add.go` — the enrolled short-circuit and the `.bit/` gate.
- `cmd/init.go` — `writeClaudeWiring`.
- `claude/sync.go` — `Runner`, `ExecRunner`, `SyncPlugin`, `RegisterMCP`.
- `cmd/add_test.go`, `cmd/init_test.go`, `claude/sync_test.go`.

## Method
- [ ] Split enrollment from readiness in `bp add`: an already-enrolled project skips the
      `CreateProject` insert and still runs the wiring. Keep a line of output saying so.
- [ ] Delete the `.bit/` existence gate — the wiring is idempotent by construction, so there is
      nothing to guard.
- [ ] Give the runner a directory. `claude.DirRunner` already exists in `claude/dispatch.go` with
      the right shape; the cheapest move is to widen `Runner` to carry `dir` rather than introduce
      a second type, and have `ExecRunner` set `cmd.Dir`. Both call sites pass the project path.
- [ ] Reverse `SyncPlugin`'s order, or verify after: `install` is what makes the plugin present
      *here*, and a succeeding `update` is not evidence of that.
- [ ] `bp add` should say what it did — enrolled, wired, plugin synced, MCP registered — so a
      re-run is legible rather than looking like a no-op.

## Claude verifies
- [ ] `just test`, `just lint`.
- [ ] A test that `bp add` on an **already-enrolled** project still calls the wiring.
- [ ] A test that `bp add` on a project that already has `.bit/` still calls the wiring.
- [ ] Fake-runner tests asserting both the argv **and** the working directory for
      `plugin install` and `mcp add`.

## User verifies
- [ ] `bp add .` in `tools/example`, which is already enrolled, then confirm `claude --agent
      bit:bot-dev ...` resolves there instead of erroring.
- [ ] Confirm `installed_plugins.json` gains a `bit@bit-pro` entry whose `projectPath` is the
      example project.

## Commit (user)
`fix(add): setup runs again, and in the project being added`