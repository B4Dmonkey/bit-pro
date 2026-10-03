# BIT-46 research index

**Start at [decisions](decisions.md):** the operator decisions of 2026-09-30 (the split into BIT-46/48/49) and 2026-10-01 (`bp init` deleted in Verse 1, the order of `bp add`'s checks, `bp remove`, and git info on every record).

**Then read [review-2026-10-02](review-2026-10-02.md):** an independent re-check of every pointer against the code. Headlines:
- The registry already exists in v1 (`store/store.go:22` `bit-pro`, `db/open.go:25` `bit.db`, 4 migrations, `code` not UNIQUE). Verse 1 must rename the dir and file and replace all migrations and `projects.sql`.
- `task.Store` has no project code (`task/store.go:25-31`), so dropping config.toml changes the `task.New` API across about 60 test calls. `initProject` has 78 call sites (not about 75) and must be rewritten in the same bar that deletes `bp init`.
- The MCP server already builds its store per call (`cmd/serve_mcp.go:294…489`); only the registry lookup is new.
- The `/var` vs `/private/var` symlink and the typed-case `$PWD` that `os.Getwd` returns break naive prefix matching (probed). `bp add` stores `filepath.Abs` unresolved.
- The removed-project resolver error and the revive path belong in Verse 3. Verse 3's archive only works on JSON if Verse 2 rewrites `relocateInto` to move pairs.
- The record-format gaps bit:plan would guess at (field names, empty vs omitted, timestamps, globs, returned path) are listed there with defaults.

The notes below predate the split. Where they say "Verse 3", "Verse 4" or "Verse 5", they mean the original BIT-46 numbering: Verse 3 is now BIT-49 Verse 1 (feedback and retro), Verse 4 is BIT-49 Verse 2 (migrate), and Verse 5 is BIT-49 Verse 3 (the sweep). The wiring parts of the old Verse 1 are now BIT-48.

## Soundness pass 2 (global Claude wiring) — read first
- **Resolved by BIT-49 (was "Blocking"):** cutover never created the global wiring. Only `bp add` ensured it, `bp add` refused a folder with `.bit/`, and migrate leaves `.bit/` in place. BIT-49's `bp migrate` now ensures the wiring on a first migration, and `bp add` checks "already registered" first (2026-10-01). → [soundness-2](soundness-2.md)
- `install --scope user` fails for a marketplace declared only in project settings (sandbox-verified). The ensure step needs `claude plugin marketplace add B4Dmonkey/bit-pro` first. `WriteSettings` becomes dead code. → [soundness-2](soundness-2.md)
- `RegisterMCP`'s `claude mcp get bit` is scope-blind. From bit-pro or a client project it finds the local entry and skips the user add. `mcp add -s user` exits 1 when the entry exists. Local beats user, and identical endpoints don't conflict. → [soundness-2](soundness-2.md)
- The "plugin behind" notice is per-project: it matches on `projectPath` and its text says `--scope project` (`claude/plugin.go:26-31`, `cmd/root.go:119`). With a user-scope install it goes silent. → [soundness-2](soundness-2.md)
- A dev `bp add` would write to the real `~/.claude*`, because the XDG sandbox doesn't cover Claude. Global MCP and skills load in every session, so unregistered-folder behaviour needs deciding. → [soundness-2](soundness-2.md)
- `init-and-update` below predates the global-wiring decision. Its per-project recommendation (option a) is superseded.

## Init & update (v2 @ 6a1d345) — resolves the "how does the operator update wiring" unknown
- All wiring steps are already idempotent: the settings.json merge, `plugin update||install --scope project`, and `mcp get||add`. The MCP entry runs `bp serve mcp` by name, so a new bp needs no re-registration. The only recurring update is the **per-project plugin version**, and the post-command notice already prints the fix. → [init-and-update](init-and-update.md)
- `db.Open` creates the data dir and file and applies the embedded migrations, so **no setup command is needed** for `main.db`. → [init-and-update](init-and-update.md)
- **Surprises:** `bp add` skips wiring when `.bit/` exists and returns `already added` without re-wiring. `ExecRunner` has no `Dir`, so `bp add <other path>` wires bp's cwd and not `<path>` (question). A migrated project keeps `.bit/`, so check registration before refusing. → [init-and-update](init-and-update.md)
- Recommendation: make `bp add .` safe to re-run (registered → re-wire only). No `bp update` yet. **Superseded (2026-10-01):** a registered `bp add` is a no-op and doesn't re-wire. → [init-and-update](init-and-update.md)

## Soundness pass (v2 @ 6a1d345)
- **Verse 1 blocker:** track creation reads the prefix from `config.toml` (`task/store.go:216-221`), which the Decisions drop. Verse 1 must take the code from the `projects` row. → [soundness](soundness.md)
- Gaps: `bp init` is in no verse, `bp add` on a folder with `.bit/` is unspecified, `code` isn't UNIQUE in the schema, and migrate needs new git code and should require uppercase IDs (`update/normalize.sh`). All since decided: `bp init` is deleted in Verse 1, `bp add` checks registration first, `code` is UNIQUE, and the git helper and uppercase check are BIT-49's. → [soundness](soundness.md)
- Verse 1 Touches miss `cmd/task/*`, `approve.go`, `feedback_add.go`, `init.go` and `cmd/tui.go`. About 18 test files create `.bit/`. → [soundness](soundness.md)
- BIT-45 topics `migrate`/`json-schema` are stale (bit.db, lowercase dir, git-log timestamps, copy-unknown-files). The track is right. → [soundness](soundness.md)

## Plugin-dir testing (earlier pass)
Resolves the scope's open unknown ("does `claude --plugin-dir ./bit` conflict with installed `bit@bit-pro`?"). Answer: **no conflict — verified empirically.**

- The inline `--plugin-dir` copy **replaces** `bit@bit-pro` for the session (one `bit` plugin, `source: bit@inline`). There are no duplicate `bit:*` skills or agents, whether or not the project enables the plugin.
- The plugin declares no MCP server. `--mcp-config dev.json --strict-mcp-config` cleanly swaps the local-scope `bp serve mcp` for a dev binary. The server gets `CLAUDE_PROJECT_DIR` = the launch dir, plus any `env` from the config (for example a sandbox `XDG_DATA_HOME`).
- Headless (`-p`) runs need `--allowedTools mcp__bit__*` for the dynamic server.
- Recommendation: test v2 with `claude --plugin-dir <v2>/bit --mcp-config <dev.json> --strict-mcp-config`. It touches no settings, so the Risk entry can be closed and BIT-49 Verses 1 and 3 can be tested this way.

## Topics
- [decisions](decisions.md): the operator decisions of 2026-09-30 and 2026-10-01. Read first.
- [review-2026-10-02](review-2026-10-02.md): the independent pointer re-check (confirmed vs wrong, with file:line), the missing inventory items, verse/order hazards, the probed symlink and case behaviour, and suggested defaults for the under-specified record format. Open it before planning any verse.
- [file-inventory](file-inventory.md): the files each v2 track changes, grouped by owning track and verse. See review-2026-10-02 for its gaps (`move_test.go`, the store/db path tests, MCP descriptions, `pluginState`).
- [soundness-2](soundness-2.md): sandbox-verified `claude` CLI behaviour for user-scope plugin and MCP, the notice changes, the unregistered-folder consequences, the cutover wiring gap (resolved by BIT-49), and the dev-test hazard. Open before planning BIT-48 Verse 1.
- [init-and-update](init-and-update.md): step-by-step init/add/wiring with file:line, the per-machine vs per-project table, the db bootstrap, and options (a)/(b)/(c). Its per-project recommendation is superseded by the global decision.
- [soundness](soundness.md): per-verse issues with file:line evidence (original BIT-46 numbering), the ordering check, and contradictions with BIT-45 notes and v2-sketch.md.
- [plugin-dir-testing](plugin-dir-testing.md): probe method, init-event evidence, MCP wiring, and the exact test recipe. Open it before planning how BIT-49 Verses 1 and 3 are tested.
- [claim-audit-2026-10-02](claim-audit-2026-10-02.md): CONFIRMED/WRONG/STALE verdict per claim in every topic and the track body, the full call-site inventory for Verse 1 (incl. MCP seed helpers the fixture count misses), a green bar order for Verses 1–3, forced API changes, and what BIT-48/49 need (incl. the worktree-cut gap for migrate). Open before bit:plan.
