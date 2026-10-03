---
id: BIT-46.11
title: bitdir and config.toml are gone; the main-checkout cut survives as project.MainCheckout
status: todo
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

With every caller on the resolver, `bitdir` and `config.toml` go. bitdir's `.claude/worktrees/` cut (`bitdir/bitdir.go:57-68`) is kept as an exported helper, because BIT-49's migrate needs it to find a worktree's main-checkout `.bit/` (each worktree has its own tracked `.bit/`, so walking up won't work). The helper changes behavior: it returns the checkout root instead of `<root>/.bit`. So it gets a test, adapted from bitdir's existing worktree cases. Everything else here is removal and adds no new tests (operator rule for v2).

## Scope
- new `project/worktree.go`: `func MainCheckout(dir string) (string, bool)`. It splits on `filepath.Separator`, finds the **first** `.claude` segment followed by `worktrees`, and returns the path before it (the outermost checkout) with `true`; otherwise `"", false`. **Seam for BIT-49:** migrate joins `.bit` itself.
- delete `bitdir/` and `bitdir/bitdir_test.go`. Its worktree cases move to `project/worktree_test.go` (below).
- `task/store.go` `Create`: delete the `Config()` fallback (the top-level branch becomes `s.NextID(s.code)`) and `configFileName`. Every production store is built by `project.OpenStore` with a code, so no new guard is added (YAGNI).
- delete `task/config.go` and its tests (`TestStoreConfig` in `task/store_test.go`, converted in BIT-46.6). Switch the `SaveConfig` seeding inside `TestStoreCreate` (`:783-786`) to `NewProject(root, "BIT")`.
- `cmd/add_test.go`, since `config.toml` no longer exists anywhere (and the grep below would otherwise hit it): `TestAddCmd/initialises a project without bit`'s `.bit/config.toml` absence check (`:107`) becomes a `.bit` absence check, and `TestAddCmd/refuses an unregistered folder with bit` (BIT-46.5) seeds only an empty `.bit/` dir, not a `config.toml`. `TestTaskCreateCmd/errors without config`'s failure message (`cmd/task/create_test.go:287`, "when config.toml is absent") says "when the folder is unregistered" instead; its assertions are unchanged.
- `go mod tidy`, which drops `github.com/BurntSushi/toml` (only `task/config.go` imports it).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMainCheckout` (table, `project/worktree_test.go`, adapted from `bitdir_test.go`)
     - **Behavior:** from anywhere inside a Claude worktree, the main checkout's folder is found. Outside one, nothing is.
     - **Rows / Assertions:** `worktree root` `/r/.claude/worktrees/hazy` → `/r, true`; `deep inside a worktree` `/r/.claude/worktrees/hazy/src/pkg` → `/r, true`; `nested worktree resolves to the outermost` `/r/.claude/worktrees/outer/.claude/worktrees/inner` → `/r, true`; `outside a worktree` `/r/src` → `"", false`; `claude dir without worktrees` `/r/.claude/settings` → `"", false`.
     - **Boundary:** depth inside the worktree (0 and 2), nesting (the first occurrence wins), and a `.claude` segment not followed by `worktrees`.
   - [ ] Confirm fails: `project.MainCheckout` undefined.

2. **Implement (GREEN):**
   - [ ] `MainCheckout`.

3. **Removal:**
   - [ ] Delete `bitdir/`, `task/config.go`, the `Create` fallback, `configFileName` and `TestStoreConfig`; switch `TestStoreCreate` to `NewProject`; `go mod tidy`.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `go mod tidy` leaves no further diff
- [ ] `grep -rn "bitdir\|config.toml\|SaveConfig\|ConfigPath" --include='*.go' .` finds nothing
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
This checks the whole verse in a sandbox (never `just install`). The fake `claude` stops `bp add`'s wiring from running real `claude` commands against the sandboxed `HOME`.
- [ ] Set up and build before sandboxing `HOME`, so Go's caches stay where they are:
  `SB=$(mktemp -d); mkdir -p $SB/home $SB/data $SB/bin $SB/proj/src/pkg; printf '#!/bin/sh\nexit 0\n' > $SB/bin/claude; chmod +x $SB/bin/claude; just db-gen-queries && go build -o $SB/bin/bp .`
  then `export HOME=$SB/home XDG_DATA_HOME=$SB/data PATH=$SB/bin:$PATH`.
- [ ] `cd $SB/proj && printf 'demo\n' | bp add .` prints `Project code: Bringing the bit plugin current...`, `Registering bit MCP server...`, `bit MCP server registered (local scope).`, then `added DEMO /private/var/.../proj` (wiring runs before registering until BIT-48).
- [ ] `cd $SB/proj/src/pkg && bp task create "First track"` prints `DEMO-1`. `ls $SB/data/bit/DEMO/tasks` shows `DEMO-1.md`, and `ls -a $SB/proj` has no `.bit`.
- [ ] `mkdir -p $SB/proj/.claude/worktrees/wt && cd $_ && bp task list` lists `DEMO-1`.
- [ ] `mkdir $SB/other && cd $SB/other && bp task list` fails with ``not a bit project; run `bp add` ``. `mkdir -p $SB/v1/.bit && cd $SB/v1 && bp task list` fails with ``run `bp migrate` ``. `bp init` says `unknown command`.
- [ ] MCP, from a **new shell without the `HOME` export** (claude needs your real login): write `$SB/dev.json` as `{"mcpServers":{"bit":{"command":"<SB>/bin/bp","args":["serve","mcp"],"env":{"XDG_DATA_HOME":"<SB>/data"}}}}`, then run `cd $SB/proj/src && claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config --allowedTools 'mcp__bit__*' -p "call mcp__bit__task_list and print the task IDs"`. It shows `DEMO-1`.
- [ ] Whole slice: a registered project works from any subfolder or worktree through both the CLI and MCP, with no `.bit/` in the repo.

## Commit (user)
`refactor(bit): delete bitdir and config.toml; keep the main-checkout cut`