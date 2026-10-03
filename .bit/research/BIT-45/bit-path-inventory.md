# Everything that touches `.bit/` paths

**Checked:** Q2 — full inventory, and what breaks if the store root moves to `~/.local/share/bit/<CODE>/`.

## Go code
| Where | How it gets the root |
|---|---|
| `bitdir/bitdir.go` | `defaultDir = ".bit"`; `Current/Canonical/ForRoot/Root` |
| `cmd/root.go:136` | `bitdir.Resolve()` in PersistentPreRunE; `:27` `bitdir.Root()` for plugin state |
| `cmd/init.go:39,75` | `task.New(bitdir.Current()).SaveConfig/Config` |
| `cmd/approve.go`, `cmd/feedback_add.go`, `cmd/tui.go:24` | `task.New(bitdir.Current())` |
| `cmd/task/{list,create,read,update,move,delete,complete}.go` | `taskstore.New(bitdir.Current())` |
| `cmd/serve_mcp.go:294..489` (10 handlers) | `task.New(bitdir.ForRoot(root))`, root = `$CLAUDE_PROJECT_DIR` |
| `cmd/serve_mcp.go:76,84,102,109` | tool **descriptions** mention `.bit/completed/`, `.bit/archive/tasks/`, `.bit/research/<track>/` (agent-visible text) |
| `cmd/add.go:51,66` | hardcoded `filepath.Join(abs, ".bit")` |
| `daemon/loop.go:64` | hardcoded `filepath.Join(p.Path, ".bit")` |
| `task/*.go` | only relative subdirs under `s.root`: `tasks`, `completed`, `archive/tasks`, `feedback`, `research/<track>`, `config.toml` |

`tui/` has no `.bit` reference (gets a loader func). `store/store.go` is the **other** data dir (`~/.local/share/bit-pro`), see [sqlite-registry](sqlite-registry.md).

## Plugin skills / agents (`bit/`) — instructions
- Via MCP only (fine if MCP moves): analyze, scope, plan, do, check, feedback, complete — they mention `.bit/...` in prose/"never hand-edit" rules; cosmetic.
- **Direct file access (breaks):**
  - `bit/skills/retro/SKILL.md:18,29` lists/reads `.bit/feedback/*.md` **directly** ("feedback notes have no tool of their own for reading"), and `:76,110` **writes** `.bit/retro/<x>-proposals.md` directly. No MCP tool covers either.
  - `bit/skills/learn/SKILL.md:3` expects a `.bit/retro/*-proposals.md` path.
  - `bit/agents/bot-dev.md:27-28` stages `.bit/` in the commit; `bit/skills/do/SKILL.md:99` says `.bit/tasks/*.md` changes join the commit. (Already moot inside worktrees, since bp writes to the main checkout's `.bit/`.)
- `bit/.claude-plugin/plugin.json` description says "tracked in .bit/".

## Other
- `README.md` (several lines, incl. a `.bit/` tree diagram at ~110).
- `update/normalize.sh` + `normalize_test.sh` — one-shot uppercase migration, takes dirs containing `.bit/`. Precedent for a migration tool; README explicitly says "nothing detects an un-migrated project".
- `.claude/hooks/*` and `Justfile`: **no** `.bit` references. `.gitignore` doesn't ignore `.bit/` (tracked here: ~370 files: 321 completed, 27 feedback, 10 tasks, 8 archive, research, config).
- Tests: ~11 test files set up `.bit` via `t.Chdir` (init_test 11×, add/serve/start 7× each, loop_test 5×). `bitdir_test.go` pins the worktree-cut behaviour.

## What breaks if the root moves
Nothing in `task/` — it's root-agnostic. Breaks: every `bitdir` caller (needs the new resolver), the two hardcoded `.bit` joins (add, daemon), MCP resolution, retro/learn skills (direct file I/O), bot-dev/do commit instructions, README, normalize.sh (v1-only tool, fine to leave), and the `.bit`-seeding tests.
