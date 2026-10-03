# Which fields get written, and through which tool

**Checked:** what completion writes and whether BIT-47's MCP inputs cover it.

- Fields (BIT-46 adds them empty, `.bit/tasks/BIT-46.md:60-62`): tracks and bars carry `branch` and single `commit`. Track `commit` = landing/merge/squash commit. Bar `commit` = own hash, repointed to the squash when rung (b) fires.
- BIT-47 (`.bit/tasks/BIT-47.md:26-28`): `task_update` and `task_complete` accept `commit` and `branch`; only the commit skill (bars) and completion (tracks) record HEAD for tracks/bars. Today `taskUpdateInput` (`cmd/serve_mcp.go:150-157`) and `taskCompleteInput` (`:160-162`) have no such fields.
- Write order constraint: `Store.Load`/`Update` only read `.bit/tasks/` (`task/store.go:37-39,164`), and `Complete` moves files to `.bit/completed/`. So bars must be repointed via `task_update{commit}` **before** `task_complete`, and the track's commit goes in `task_complete{id, commit, branch}` (or a `task_update` just before). Re-check after BIT-46's JSON storage.
- Track `branch` at completion: **unknown** — the trunk (`main`) or the work branch the bars were on? The stub doesn't say; the trunk is what "landed" means.
- No author identity is needed for any command used (all are commit/message based).
- Where git runs: BIT-49's git helper (shells out, stubbed in tests), with `-C` the session dir. BIT-50 extends it with is-ancestor, cat-file, rev-list, log --grep, branch -r --contains. The helper doesn't exist yet on v2 (no git package in tree).

**Verdict:** BIT-47's inputs suffice if they land as described; the only new MCP surface BIT-50 might want is a read-only "where did this track land?" tool so the skill doesn't run git itself — **open choice**: logic in Go (testable, reusable by CLI) vs. git commands in the skill.
