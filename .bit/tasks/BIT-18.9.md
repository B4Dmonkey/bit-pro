---
id: BIT-18.9
title: A live sign-off completes instead of archiving
status: doing
phase: 3
phase_label: The session says complete
---
## **Verse 3**

bit_do's Track sign-off is the one place a session decides where finished work goes, and it
says `bp task archive`. Until it says `complete`, every sign-off files a shipped track into
the soft-delete bin no matter what the CLI supports. This is the last bar of Verse 3 and of
the track, so it carries the integration check.

## Scope
- `bit/skills/do/SKILL.md` — the `description:` frontmatter line, the Track-status rollup
  note, the Track sign-off section (its opening paragraph and steps 2–3), and the
  *What this skill does not do* line.
- `bit/skills/feedback/SKILL.md` — the "Both active and archived tracks are accepted" line.

## Edits
- [ ] `do/SKILL.md` sign-off step 2: `bp task complete <track>` relocates the track and all
      its bars into `.bit/completed/`. Drop the "Use `archive`, not `delete`" clause — it
      exists to disambiguate two verbs pointing at one folder, which is the confusion being
      removed. Replace the `--force` note too — `complete` has no override, so the line
      becomes: every bar must be `done`, which a signed-off track already satisfies.
- [ ] `do/SKILL.md` sign-off step 3 and the section's opening paragraph, the rollup note, the
      `description:`, and the *does not do* line: they all say the sign-off "archives" the
      work. It *completes* it. Say `completed/`, not "files it away", so a reader can't read
      it as the archive.
- [ ] `feedback/SKILL.md`: active, completed, or archived tracks are all accepted.
- [ ] Keep the frontmatter structurally untouched apart from the `description:` value — same
      keys, same `name: bit_do`, no reflowing. The colon-space YAML trap lives here.

## Claude verifies
- [ ] `grep -rn "task archive" bit/skills/` returns nothing.
- [ ] `grep -q "task complete" bit/skills/do/SKILL.md` and
      `grep -q "completed" bit/skills/feedback/SKILL.md`.
- [ ] skill-creator frontmatter validation on both edited skills:
      `SC=~/.claude/plugins/cache/claude-plugins-official/skill-creator/unknown/skills/skill-creator`
      then `uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py"
      bit/skills/do/` and the same for `bit/skills/feedback/`. The kebab-case `name:`
      complaint is known standing noise — every bit skill fails it and the plugin loads
      fine. Any *other* failure is real.
- [ ] `claude plugin validate ./bit`
- [ ] `just test` and `just lint` — neither skill is compiled, but the tree has to stay green.

## User verifies
- [ ] Whole slice, after the commit is pushed: `claude plugin marketplace update bit-pro`,
      restart, then start a `/bit:do` session on a track whose bars are all done. The
      sign-off step it offers says `bp task complete <track>`, and running it puts the track
      and its bars in `.bit/completed/` — a live session now files finished work in the right
      place without being told.

## Commit (user)
`docs(skills): sign a track off with task complete`