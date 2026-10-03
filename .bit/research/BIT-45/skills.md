# Skills and agents vs the central store (Q6)

**Checked:** which `bit/skills/*/SKILL.md` and `bit/agents/*.md` reference `.bit/` paths or file shapes, and which need new MCP tools. This supersedes the skills section of [bit-path-inventory](bit-path-inventory.md).

The constraints are the operator's:
- Bodies stay markdown, with JSON metadata alongside.
- Feedback and retro move to top-level dirs, as JSON with a `project` field.

## Breaks: direct file I/O
- **retro** (`bit/skills/retro/SKILL.md`)
  - Lists and reads `.bit/feedback/*.md` directly (:18, :29, which says "Feedback notes have no tool of their own").
  - Writes `.bit/retro/<track-or-album>-proposals.md` (:27, :76).
  - Globs `.bit/retro/*-proposals.md` to de-duplicate (:110).
  - Once feedback is top-level, retro needs a tool that lists by project and track.
- **learn** (`bit/skills/learn/SKILL.md:3,16`) takes a proposals file path under `.bit/retro/`, or pasted content. The paste path survives. The path form goes stale.
  - With a top-level retro dir, learn (which runs only in bit-pro) could list proposals from *every* project through a tool. That reverses the "carried over by hand" premise, which is worth a scope decision.
- **check** (`bit/skills/check/SKILL.md`) writes `<track-id>-check.md` in the **repo root**, not `.bit/` (:101), and re-reads it (:154-155). v2 doesn't affect it.
  - It says retro consumes it (:8, :101), but retro never reads it. This mismatch already exists in v1 and is out of scope.

## Semantics change (delete, don't reword)
- `bit/agents/bot-dev.md:27-28` stages `.bit/` in the bar commit.
- `bit/skills/do/SKILL.md:99` says `.bit/tasks/*.md` changes join the commit.
- With state outside the repo, both go.

## Cosmetic only (all access is through MCP)
feedback (:3, :12, :113), do (:8, :104, :108, :131-132), complete (:3, :8, :12, :23), scope (:10, :110, :205, :227), analyze (:3, :17, :64), plan (:8), `agents/ruler.md:12`, `agents/bot.md:8,16,22`. Also `README.md` (7, 34, 42, 59, 65-67, 103, a tree at 110, 137, 141) and `bit/.claude-plugin/plugin.json:5`.

## Body-structure reliance: unaffected
- Verse checklists (scope, do :94, complete :22, check :43) are text inside the body.
- The "User verifies" branching (do :63-89, bot-dev :20-22) is text inside the body.
- analyze's `[x](x.md)` index links are text inside the body.
- Bodies stay markdown, so none of this changes.

## New MCP tools implied
1. `feedback_list(track?)` and `feedback_read(id)`. Retro needs them. `feedback_add` keeps its signature, and the server fills in `project`.
2. `retro_write(name, body)` plus `retro_list()` and `retro_read(name)`, stored as JSON + md in the top-level `retro/`. Learn can use `retro_read` or keep the paste path.
3. Optional and not required for centralization: `check_write`/`check_read`.

## Other consumers
- `.opencode/bit-list.tsx` (untracked) shells `bp task list/read` with `cwd`. It has no `.bit` paths.
- `opencode.json` registers `bp serve mcp` with no `CLAUDE_PROJECT_DIR`. See [project-resolution](project-resolution.md).
- Skills ship from GitHub `main`. v2 skill edits therefore land at cutover, or through `--plugin-dir` during dev. See [testing](testing.md).
