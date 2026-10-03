---
id: BIT-49.8
title: retro, learn and feedback work through the new tools with no file I/O
status: todo
approved: true
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

The tools exist now, but retro still lists `.bit/feedback/*.md` and writes `.bit/retro/` by hand, and learn still asks for a hand-carried file. This bar rewrites those three skills and bot's tool list to match the shared model. It edits skill text only, so it has no Go test. It's checked with skill-creator, `claude plugin validate` and a sandboxed `--plugin-dir` run. This bar completes the verse.

## Scope
Line numbers are from v2 @ 6a1d345; re-grep before editing.
- `bit/skills/retro/SKILL.md`:
  - `:3` (description): reads the current project's feedback notes through `mcp__bit__feedback_list`/`feedback_read`, not `.bit/feedback/*.md`. Drop "runs somewhere else entirely". Proposals go to the shared store, where bit_learn reads them.
  - `:10`: retro's one artifact is a proposals record, not a file. It still edits nothing and makes no CLI change. Drop "leaves this project … runs somewhere else" and the pointer to the *Handoff* section, which doesn't exist. Say that bit_learn picks proposals up from the shared store.
  - `:18`: the notes are listed by `mcp__bit__feedback_list` (this project only) and read by `mcp__bit__feedback_read`.
  - `:27`: retro *writes* through the tool surface too, with `mcp__bit__retro_write`. Delete the "plain file you write directly / no retro tool" asymmetry.
  - `:29`: replace "no tool of their own … List `.bit/feedback/*.md` directly" with `feedback_list`, then `feedback_read` on each ID.
  - `:64`: an accepted proposal goes in the record, not the file. `retro_write` replaces the record by name, so retro writes the accepted proposals together in one call once the walk-through is done (a later call resends the whole body).
  - `:76`: "Each accepted proposal goes in one record, written with `mcp__bit__retro_write` under the name `<track-or-album>-proposals`. The server prefixes the project code, so the stored name is `<CODE>-<track-or-album>-proposals`." Keep the record's body format as it is.
  - `:102`: report the stored name `retro_write` returns. bit_learn finds it with `retro_list`.
  - `:110`: de-duplicate against `mcp__bit__retro_list` plus `retro_read`, not a `.bit/retro/*-proposals.md` glob.
  - `:123`: delete the "Move the proposals file anywhere …" bullet. There's no file to move. `:122` (learn runs in bit-pro) stays.
  - The confidentiality promise stays, and is now backed by `feedback_list` only ever returning this project's notes.
- `bit/skills/learn/SKILL.md`:
  - `:3` (description): reads retro proposals from every project through `mcp__bit__retro_list`/`retro_read`. Drop "carried over by hand" and "never inside the project the proposals came from". Keep that learn runs inside bit-pro, because that's where the skills and CLI live.
  - `:8`: retro runs in whatever project produced the feedback and stores its proposals in the shared store, and learn reads them from there.
  - `:16`: list with `retro_list`. If the user didn't name one, show the list and ask which to work through. Read the chosen one in full with `retro_read`. A pasted proposal is still accepted.
  - `:22`: "before they're carried over" → "before learn reads them".
  - Nothing else about how learn works changes (the scope revisits it later).
- `bit/skills/feedback/SKILL.md` `:3`, `:12`, `:113`: drop `.bit/feedback/`. The note "goes in through `mcp__bit__feedback_add`, the only way in", with no path. The tool list on `:12` is unchanged.
- `bit/agents/bot.md`:
  - `:22`: add `mcp__bit__retro_write` to "the whole write surface" list. Leave the rest of the line, including "Never hand-edit `.bit/tasks`", for BIT-49.28.
  - `:57`: "handing over a retro proposals file" → "working through retro proposals". The `only inside bit-pro itself` column stays.
- Text only: no Go change and no test files.

## Change checklist
- [ ] Edit the lines above.
- [ ] `grep -n '\.bit' bit/skills/retro/SKILL.md bit/skills/learn/SKILL.md bit/skills/feedback/SKILL.md` finds nothing.
- [ ] `grep -n 'carried over\|carry elsewhere\|Handoff\|runs somewhere else' bit/skills/retro/SKILL.md bit/skills/learn/SKILL.md` finds nothing.

## Claude verifies
- [ ] `SC=$(ls -d ~/.claude/plugins/cache/claude-plugins-official/skill-creator/*/skills/skill-creator | head -1); for s in retro learn feedback; do uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/$s/; done`. Only the known kebab-case `name:` failure is allowed.
- [ ] `claude plugin validate ./bit` passes
- [ ] `just lint` and `just test` pass

## User verifies
Sandbox as in BIT-46.11's steps (never `just install`), with two projects, `demo` (`DEMO`) and `other` (`OTHER`), each with a track and a note (`bp feedback add DEMO-1 -d "demo note"` / `bp feedback add OTHER-1 -d "other note"`). For the MCP sessions, use BIT-46.11's `$SB/dev.json` and `claude --plugin-dir <v2 checkout>/bit --mcp-config $SB/dev.json --strict-mcp-config` from a shell without the `HOME` export.
- [ ] From `$SB/demo`, run `/bit:retro DEMO-1` interactively. The transcript shows `mcp__bit__feedback_list` and `feedback_read` calls and no Read/Glob of a `.bit` path. It never shows `other note`. Accept one proposal, and it reports a stored name starting with `DEMO-`.
- [ ] From `$SB/other`, run `/bit:learn`. It lists the `DEMO-…-proposals` record with project `DEMO` through `retro_list`, and opens it with `retro_read` when chosen.
- [ ] `ls $SB/data/bit/retro` shows the `DEMO-…-proposals.{json,md}` pair, and `ls $SB/data/bit/feedback` shows both projects' notes.
- [ ] Whole slice: feedback and retro live in the shared folders, retro sees only its own project's notes, and learn sees every project's proposals, all without a file path in sight.

## Commit (user)
`feat(bit): retro, learn and feedback use the shared feedback and retro tools`