# bit-pro

A markdown-backed task tracker built for Claude Code. You describe the work; Claude
researches, scopes, plans and builds it one commit at a time, and every step is tracked.

The CLI is `bp`. Tasks live in `~/.local/share/bit` (or `$XDG_DATA_HOME/bit`), keyed by
project code. Claude reads and writes them through the `bit` MCP tools; you watch them in
`bp ui`.

## Install

Needs: Go, git, [Claude Code](https://claude.com/claude-code), `jq` and
[`sqlc`](https://sqlc.dev).

```
git clone https://github.com/B4Dmonkey/bit-pro.git
cd bit-pro
./scripts/install.sh
```

This builds `bp` into your Go bin dir (`go env GOBIN`, else `$(go env GOPATH)/bin`) and
registers the `bit-pro` plugin marketplace with auto-update on. Make sure that dir is on your
`PATH`, then check with `bp --version`.

Then register each project you want to track:

```
cd your-project
bp add .        # asks for a project code, e.g. ACME
```

The first `bp add` also installs the `bit` plugin and the `bit` MCP server in Claude Code
(user scope). If that step fails, it prints the commands to run by hand.

Already have a v1 `.bit/` folder in a project? Run `bp migrate` there instead of `bp add`.

## Workflow

Two agents do the work: **theRuler** plans, **theCreator** builds.

**1. Plan with theRuler**

```
claude --agent bit:theRuler
> add CSV export to the reports page
```

theRuler creates a track (`ACME-7`), researches the code, and writes a scope: why, and the
order to deliver it in. It **stops for you to approve the scope**. Ask questions and it
researches again. Once you approve, it breaks the scope into bars (`ACME-7.1`, `ACME-7.2`, …):
one TDD step and one commit each. It never writes code.

**2. Build with theCreator**

```
claude --agent bit:theCreator
> do ACME-7.1
```

theCreator implements one bar, runs its checks, **asks before committing**, records the
commit on the bar and pushes if the repo has a remote. Then it stops. Run it again for the
next bar.

**3. Close the track**

Once the work is pushed or merged, run `/bit:complete ACME-7` in Claude to record where it
landed and move the track out of the active list.

Use `bp ui` in another terminal to follow along. It refreshes as Claude edits tasks.

## What's included

**CLI**

| Command | What it does |
|---|---|
| `bp add <path>` | Register a project under a code |
| `bp list` | List registered projects |
| `bp migrate` | Move a v1 `.bit/` folder into the store and register the project |
| `bp task list` | Tasks across every project. `-p <track>` shows one plan in order |
| `bp task read <id>` | One task. `--body` prints just the markdown |
| `bp ui` | Terminal UI: arrow keys or `hjkl` to move, `enter` to open, `?` for help |
| `bp serve mcp` | The MCP server Claude Code talks to (started for you) |

**Plugin agents**

- `bit:theRuler`: describe the work, get an approved scope and plan
- `bit:theCreator`: implement, commit and push one bar
- `bit:bot`: general helper that knows the task tools ("what's left on ACME-7?")

**Plugin skills**: callable directly when you want one step on its own:

| Skill | Does |
|---|---|
| `bit:analyze` | Research the code before scoping |
| `bit:scope` | Write the why and delivery order |
| `bit:plan` | Break a scope into bars |
| `bit:do` | Build one bar |
| `bit:commit` | Commit a bar after asking |
| `bit:check` | Audit finished work against the plan |
| `bit:complete` | File a landed track as completed |
| `bit:feedback` | Note a correction mid-run |
| `bit:retro` | Find patterns across feedback notes |
| `bit:learn` | Turn retro findings into skill changes (bit-pro repo only) |

Plugin updates arrive automatically; no reinstall of `bp` needed.

## Uninstall

There's no script yet. Run:

```
claude plugin uninstall bit@bit-pro --scope user
claude plugin marketplace remove bit-pro
claude mcp remove -s user bit
rm "$(go env GOBIN)/bp" 2>/dev/null || rm "$(go env GOPATH)/bin/bp"
```

To delete every task in every project too (can't be undone):

```
rm -rf "${XDG_DATA_HOME:-$HOME/.local/share}/bit"
```

## More

- [hierarchy.md](./hierarchy.md): tracks, bars, verses, and the rest of the vocabulary

## Inspiration

- [Backlog.md](https://github.com/MrLesk/Backlog.md) — git-native markdown tasks, board UI.
- [kanban-md](https://github.com/antopolskiy/kanban-md) — markdown-backed kanban with filtering.
- [ccusage](https://github.com/ryoppippi/ccusage) — Claude Code usage and cost analysis from local transcripts.
