# Operator decisions

These are operator decisions, not research findings, unless an entry is marked "Claude default".

# 2026-09-30

## Scope
This track was split out of the original BIT-46. It covers shared feedback and retro, `bp migrate` and the `.bit/` sweep (the old Verses 3–5, now this track's Verses 1–3). It also owns the git helper that migrate uses to read HEAD.

## Confidentiality
- `feedback_list` returns only the current project's notes by default. Notes quote exchanges verbatim, and retro promises it's safe to run inside a confidential client project, so it must never read another project's raw notes.
- Retro proposal file names include the project code. Otherwise names could collide in the shared `retro/` folder, since today's names are `<track-or-album>-proposals.md` (`bit/skills/retro/SKILL.md:76`).
- learn's description is rewritten. Today it describes the v1 model, with proposals "carried over by hand" and learn "never inside the project the proposals came from" (`bit/skills/learn/SKILL.md:3`). In v2, learn reads every project's proposals through MCP.

## Cutover
The operator alone decides when v2 is ready and merges the branch. No track gates cutover. Migrating each project is a step on the cutover checklist, run when the operator chooses.

# 2026-10-01

## migrate and the wiring
- Re-running migrate on a registered project doesn't set up the wiring. Only the first migration, the one that registers the project, ensures it. The wiring is global, user-scope, and owned by BIT-48.

## What migrate writes for git info, per record kind
- Every record carries git info, for traceability (BIT-46 topic `decisions`).
- Tracks and bars (`tasks/`, `completed/`, `archive/`): `branch` = the branch at migration, and `commit` = HEAD at migration.
- Research topics, feedback notes and retro proposals: `commits` = one entry, `{sha: HEAD at migration, branch: branch at migration, at: migration time}`.
- In a folder with no git, every git field stays empty.

## The git helper
- Git facts come from the session dir, not the registered project path. The session dir is the worktree the session runs in: `CLAUDE_PROJECT_DIR` for MCP, and the current folder for the CLI. A Claude worktree has its own branch and HEAD (BIT-45 topic `history-anchors`).
- BIT-47 and BIT-50 extend it.

## The sweep and bot-dev
- The sweep only removes the `.bit/` staging lines from do and bot-dev.
- In v2, bot-dev doesn't commit unattended: it stops and asks the operator's permission before every commit, through BIT-47's commit skill. Dispatching bot-dev doesn't count as permission. BIT-47 rewrites bot-dev's commit text, so the sweep mustn't reword it to keep the unattended behaviour.

## Claude defaults (the operator can veto these)
- The new MCP tools are `feedback_list`, `feedback_read`, `retro_write`, `retro_list` and `retro_read` (BIT-45 topic `skills`). `feedback_add` keeps its signature, and the server fills in `project`.
- `feedback_list` has no cross-project option in this track.
- Retro proposal names are `<CODE>-<track-or-album>-proposals`.
- A migrate re-run on a registered project stops with "already migrated" and changes nothing.
- migrate refuses a code that belongs to another project, including a removed one, the same as `bp add`.
- In a Claude worktree, migrate reads the main checkout's `.bit/`.
- migrate prints the cleanup step (`git rm -r .bit` or removing the folder) and never runs it.
- A git error, or no git, gives empty values and never fails a command.

## Readiness pass, 2026-10-01 (Claude defaults; the operator can veto these)
- **migrate registers the folder that holds the `.bit/` it reads.** In a Claude worktree that's the main checkout, never the worktree, so the project resolves from the checkout and every worktree under it.
- **migrate applies BIT-46's code format and reserved names** (`^[A-Z][A-Z0-9]*$`; `FEEDBACK` and `RETRO` reserved so no project dir collides with the top-level `feedback/` and `retro/` on case-insensitive APFS).
- **Every bar that edits a skill or agent validates it** with skill-creator's `quick_validate.py` and `claude plugin validate ./bit` (memory `bit-skill-edits-run-skill-creator`). The analyze evals that name `.bit/` are swept too.
- **Dev runs never use `just install` on `v2`,** and sandbox `XDG_DATA_HOME` and `HOME`. Migrate is tried against a copy of a real `.bit/`.
- **The README sweep includes its command table:** `bp init` out; `bp add`, `bp remove` and `bp migrate` in.

Background research for this track is in BIT-45 topics `skills`, `migrate`, `json-schema` and `history-anchors` (those notes predate some decisions; the track body wins) and BIT-46 topic `file-inventory`.