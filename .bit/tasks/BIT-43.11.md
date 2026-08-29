---
id: BIT-43.11
title: No bp reference survives the plugin-wide sweep
status: todo
phase: 5
phase_label: Full cycle
---
## **Verse 5**

Verses 1–3 each swept one file at a time. This bar sweeps the whole plugin at once, which is the
scope's own done-when condition, and catches anything the per-file passes missed — a reference
added while a later verse was in flight, or one in a file no verse touched.

No RED step: this is an audit, and the sweep either comes back at the expected total or it does
not. That total is the falsifiable observation.

## Scope
- Any of `bit/skills/*/SKILL.md` or `bit/agents/*.md` still carrying a driving reference — expected
  to be none, but a straggler found here is fixed here, through skill-creator, exactly as the
  earlier verses did.

## Method
- [ ] Run the sweep across the whole plugin:
      ```
      PAT='bp (task|feedback|instructions)|`(task|feedback) [a-z]+'
      grep -cE "$PAT" bit/skills/*/SKILL.md bit/agents/*.md
      ```
- [ ] **Expected: every file reports 0 except `bit/skills/plan/SKILL.md`, which reports 1** — the
      User-verify *example* documented on that file's migration bar, a command the human runs in
      their own terminal, which stays by design.
- [ ] Confirm the two deliberate CLI references survive, neither of which matches the pattern:
      `grep -rn 'bp approve' bit/` finds `bit/skills/do/SKILL.md`'s refusal message and
      `bit/agents/bot-dev.md`'s prohibition. Approval is deliberately not a tool; if these have
      been "migrated" to a tool call, that is a regression to undo.
- [ ] Fix any straggler via skill-creator, then re-run the sweep.
- [ ] Push anything this bar changed. If it changed nothing, say so rather than pushing an empty
      commit.

## Claude verifies
- [ ] The sweep totals 1 across all nine files, and that hit is in `bit/skills/plan/SKILL.md`
- [ ] `grep -rn 'bp approve' bit/` finds exactly the two lines above
- [ ] `claude plugin validate ./bit` passes

## User verifies
- [ ] none — deterministic. The real-cycle check is the next bar.

## Commit (user)
`chore(skills): sweep the last bp references out of the plugin`

(Skip the commit entirely if the sweep found nothing to fix.)