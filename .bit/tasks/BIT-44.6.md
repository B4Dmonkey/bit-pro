---
id: BIT-44.6
title: bit:scope cites existing research instead of copying it
status: done
approved: true
phase: 2
phase_label: Scope reads research
---
## **Verse 2**

bit:scope builds on research when it exists, so a scope written after analyze stays short: it cites the notes instead of copying the evidence. Without this, bit:scope still does its own light research and never sees the notes.

## Scope
- `bit/skills/scope/SKILL.md`: in Gathering context and in Refine, call `mcp__bit__research_read` for the track first. If there are topics, read `index`, open only the topics a verse or risk needs, and cite topic names in the body instead of pasting findings. With no research, behave exactly as today. Add a research line to the References guidance so the body points at `.bit/research/<track>/`.

## Method
- [ ] Edit with `skill-creator:skill-creator`. Keep the existing light-research path intact for tracks that have no research.

## Claude verifies
- [ ] `uv run --quiet --with pyyaml python ~/.claude/plugins/cache/claude-plugins-official/skill-creator/unknown/skills/skill-creator/scripts/quick_validate.py bit/skills/scope/` (kebab-case noise expected)
- [ ] `claude plugin validate ./bit`

## User verifies
- [ ] In `claude --plugin-dir ./bit`, create a throwaway track, run `/bit:analyze` on it, then `/bit:scope` to refine it. The session calls `research_read` before drafting, and the resulting body names research topics instead of restating the findings. Then `bp task delete` the throwaway track. **Whole slice:** a scope written after analyze stays short.

## Commit (user)
`feat(bit): bit:scope builds on track research`