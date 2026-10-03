# init-and-update (v2 @ 6a1d345)

**Checked:** what `bp init` / `bp add` do today, what an "update" has to refresh, who creates the db, and what surface replaces "re-run `bp init`" once init folds into add.

## What the code does today

`bp init` (`cmd/init.go:25-44`):
1. Prompts for the prefix (default: existing `config.toml` prefix), writes `.bit/config.toml` (`task.SaveConfig`). Idempotent (overwrites with the same value).
2. `writeClaudeWiring(cmd, run, ".")` (`cmd/init.go:51-71`):
   - `claude.WriteSettings(<dir>/.claude/settings.json)` (`claude/settings.go:19`): read-merge-write of `extraKnownMarketplaces.bit-pro = {github B4Dmonkey/bit-pro}` and `enabledPlugins["bit@bit-pro"]=true`. Other keys kept. Idempotent. **Per project** (a file in the repo).
   - `claude.SyncPlugin` (`claude/sync.go:33`): `claude plugin marketplace update bit-pro` (**per machine**), then `claude plugin update bit@bit-pro --scope project`, falling back to `plugin install ... --scope project`. Idempotent. The installed version is recorded **per project** (`installed_plugins.json` entries keyed by `projectPath`; `claude/plugin.go:10-34`), so each project lags independently.
   - `claude.RegisterMCP` (`claude/sync.go:21`): `claude mcp get bit` and, if missing, `claude mcp add bit -- bp serve mcp` (default local scope: an entry in `~/.claude.json` keyed by the project dir). Idempotent. It runs `bp` by name, so a new bp binary needs **no** re-registration.

`bp add <path>` (`cmd/add.go:26-80`): `db.Open()`; if `ProjectExists(abs)` prints `already added` and returns **without wiring** (`:45-48`). Otherwise prompts for the code (default: `.bit/config.toml` prefix), runs `writeClaudeWiring` **only when `<abs>/.bit` does not exist** (`:66-70`), then `CreateProject`.

Post-command notice (`cmd/root.go:50-68`): after every non-quiet command, `claude.RefreshMarketplace()` fires `claude plugin marketplace update bit-pro` in the background, then `pluginState` compares the project's installed version with the marketplace clone's `bit/.claude-plugin/plugin.json` and prints `bit plugin X → Y available — run: claude plugin update bit@bit-pro --scope project`. It uses `bitdir.Root()` as the project root, so v2 must feed it the resolved project path instead.

## Where the pieces live

| Piece | Scope | Changes when |
|---|---|---|
| marketplace clone | machine | every bp command refreshes it (background) |
| plugin version | per project | a new plugin release, pushed to GitHub `main` |
| `.claude/settings.json` keys | per project | only if bp changes the marketplace/plugin key |
| MCP entry (`bp serve mcp`) | per project (local scope, in `~/.claude.json`) | never, since bp is called by name |

The plugin installs from GitHub (`claude/settings.go:17`; memory `bit_plugin_installs_from_github`), so a bp release does not ship skills. The only recurring per-project update is the plugin version, and the notice already names the exact command. "Re-run `bp init`" was only a convenience wrapper around `SyncPlugin`.

## main.db on a fresh machine

`db.Open` (`db/open.go:19-38`) calls `store.Dir()`, which `MkdirAll`s the data dir (`store/store.go:23`), then dbmate `CreateAndMigrate()` on embedded migrations. That creates the file and schema on first use. **No setup command is needed.** Today it is `~/.local/share/bit-pro/bit.db`. v2 changes the dir to `bit` and the name to `main.db`, and replaces the migrations (BIT-45 drops the queue).

## Surprises / questions (verify)
- `ExecRunner` (`claude/sync.go:12`) sets no `cmd.Dir`. So `bp add ../other` writes settings.json into `../other` but runs `plugin update --scope project` and `mcp add` (local scope) against **bp's cwd**. This might wire the wrong project. Is that intentional? It is harmless only when the operator runs `bp add .`.
- `bp migrate` copies and never deletes `.bit/` (Decisions). So a migrated project **still has `.bit/`** until cutover's `git rm`. Any "refuse if `.bit/`" check in `add` must come **after** the "already registered" check, or a re-run on a migrated project is refused.
- Does migrate wire Claude? It's unspecified. A v1 project is already wired identically, so it probably doesn't need to.
- `bp add` takes a path arg. It must not go through the resolver's "not migrated" error, or it can't register anything.

## Options
- **(a) `bp add` safe to re-run.** The operator types `bp add .`. If registered: skip registration, run `writeClaudeWiring`, print `already added; wiring refreshed`. Unregistered + `.bit/` → migrate error. Unregistered, no `.bit/` → prompt, wire, register. Pros: no new command, same idempotency v1 init had. Cons: "add" reads oddly as an update. Order of checks is load-bearing (registered → `.bit/` → new).
- **(b) `bp update [--all]`.** Re-wires the resolved project, or every registered one. Pros: clear name, can update all projects at once. Cons: a new command plus an `--all` loop (each needs its cwd set, see the ExecRunner issue), and it duplicates what (a) gets for free.
- **(c) Nothing new.** Updating is the plugin notice's `claude plugin update ... --scope project`. Pros: zero code. Cons: settings/MCP drift is never repaired, and a project that lost its wiring has no bp path back.

## Recommendation
(a), minimal. Registered → re-wire and stop. It keeps v1's "re-run to update" habit under the new name. Add (b) `--all` only if updating many projects one by one becomes painful.
