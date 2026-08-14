---
id: BIT-23.10
title: bit:do gates on approved before starting a bar
status: todo
phase: 6
phase_label: Skill update
---
## **Verse 6**

The approved field now exists on every bar, but bit:do doesn't know about it yet. It starts bars without checking whether they've been approved. This bar updates the bit:do skill so it reads the approved field from the bar's frontmatter and stops if the gate isn't cleared.

## Scope
- `bit/skills/do` — the source skill document (not the compiled asset — edits go through the plugin route via skill-creator, then `just install` + `bp init`)

The change the skill needs:
1. **Before starting a bar** — read `approved:` from the bar's frontmatter via `bp task read <BAR>`. If false or absent, stop and tell the user: "BIT-X.N is not approved — approve it first with `bp approve BIT-X.N` or in the TUI."

## Method
- [ ] Invoke the skill-creator skill pointing at the bit:do skill document
- [ ] Provide the approved-gate requirement above as the change brief
- [ ] Review the draft: does the approved gate read from frontmatter via the correct CLI command?
- [ ] Accept or iterate; once satisfied, run `just install && bp init` to bake the updated skill into the installed plugin

## Claude verifies
- [ ] `just install` succeeds
- [ ] `bp init` succeeds (embeds the updated skill)
- [ ] `bp instructions` output for bit:do references `approved`

## User verifies
- [ ] Open a test bar that has `approved: false`; run `bit:do` on it — it stops and prompts for approval before doing any work

## Commit (user)
`feat(skills): bit:do gates on approved field before starting a bar`
