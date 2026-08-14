---
id: BIT-23.10
title: bit:do reads check and approved from each bar
status: todo
phase: 6
phase_label: Skill update
---
## **Verse 6**

The Go work added `check` and `approved` fields to every bar, but bit:do doesn't know about them yet. It improvises a verification step instead of running the declared check, and it doesn't gate on approval before starting. This bar updates the bit:do skill so it reads both fields from the bar's frontmatter and acts on them.

## Scope
- `bit/skills/do` — the source skill document (not the compiled asset — edits go through the plugin route via skill-creator, then `just install` + `bp init`)

The changes the skill needs:
1. **Before starting a bar** — read `approved:` from the bar's frontmatter via `bp task read <BAR>`. If false or absent, stop and tell the user: "BIT-X.N is not approved — approve it first with `bp approve BIT-X.N` or in the TUI."
2. **Verification** — use the declared `check:` field instead of improvising. The verification step becomes: `bp check <BAR>`. If the field is absent, fall back to `just test` + `just lint` and note that a declared check was not found.
3. Do not run `bp check` if `Check` is empty — only skip to the fallback.

## Method
- [ ] Invoke the skill-creator skill: `Skill("skill-creator")` pointing at the bit:do skill document
- [ ] Provide the three delta requirements above as the change brief
- [ ] Review the draft for accuracy: does the approved gate read from frontmatter via the correct CLI command? Does the check step call `bp check <BAR>` and fall back correctly?
- [ ] Accept or iterate; once satisfied, run `just install && bp init` to bake the updated skill into the installed plugin

## Claude verifies
- [ ] `just install` succeeds
- [ ] `bp init` succeeds (embeds the updated skill)
- [ ] `bp instructions` output for bit:do references `approved` and `check`

## User verifies
- [ ] Open a test bar that has `approved: false`; run `bit:do` on it — it stops and prompts for approval before doing any work
- [ ] Open a bar with `check: "just test"` and run `bit:do` — the verification step runs `bp check <BAR>` (which shells out to `just test`) rather than improvising a check

## Commit (user)
`feat(skills): bit:do reads check and approved from bar frontmatter`