---
id: BIT-17.7
title: The offer reaches the points corrections land
status: todo
phase: 3
phase_label: Offered, not remembered
---
## **Verse 3**

`bit_do` and `bit_plan` learn to offer the note at the moments a correction actually lands, so the
user answers yes or no instead of having to remember. Skill text, no code — and per the scope's
decision this is convenience on top of Verses 1 and 2, which already work without it.

## Scope
- `bit/skills/do/SKILL.md` — two insertion points, both places where a correction is already the
  subject:
  - the close-out follow-up paragraph (currently line 70, "The user often follows up with small
    cleanup…"). Handling the tweak in place stays exactly as written; what is added is that once it
    is handled, offer the note — a tweak the automated checks did not catch is a decision the plan
    did not make.
  - `### Not as expected` (currently line 105). All three cases there are corrections, and the
    first two — *the scope is wrong*, *the plan is wrong* — are plan gaps by definition. Offer the
    note as part of the hand-back, before control leaves for bit_scope or bit_plan, because that is
    the moment the evidence still exists.
- `bit/skills/plan/SKILL.md` — two insertion points:
  - `### Trap 1: a decision wearing a checkbox` (currently line 209). A launder caught here *is* a
    decision the plan did not make; the hand-back to bit_scope gains an offer to record it.
  - `## Refining an existing plan` (currently line 328). A bar reworded or split because it was
    wrong is the same event seen after the fact — worth an offer at the end of the pass, once the
    edits are agreed.

At every point the wording is an **offer**, one line, answered yes or no: nothing records a note
autonomously. A skill that fired itself mid-run would derail the run being repaired, and the yes/no
keeps the human as the judge of what counts as feedback. Point each offer at `/bit:feedback` rather
than at `bp feedback add` — composing the note is that skill's job, and duplicating the note's shape
into two more skills is exactly the drift the single-home rule exists to prevent.

Do not touch `bit/skills/scope/SKILL.md` or `bit/skills/check/SKILL.md`. The scope names do and plan
only, and bit_check is retro's neighbour — it belongs to the later consumer scope, not this one.

**Deliberately not verified here:** whether the offer actually fires. The scope decides that the
recommendation is best-effort and that this is accepted — the user invoking `/bit:feedback`
explicitly is the path that always works, so a missed offer is not a defect and no hook or
enforcement is in scope. The checks below prove the text ships, not that a model acts on it.

## Claude verifies
- [ ] `just test` and `just lint` — no Go changed, but the tree must stay green
- [ ] `claude plugin validate ./bit` exits 0; the missing-`author` warning is pre-existing
- [ ] `grep -c 'bit:feedback' bit/skills/do/SKILL.md` is 2 and the same for
      `bit/skills/plan/SKILL.md` — all four insertion points landed
- [ ] `git diff --stat bit/skills/` lists exactly two files
- [ ] `grep -rn 'bit:feedback' bit/skills/scope/SKILL.md bit/skills/check/SKILL.md` finds nothing

## User verifies
- [ ] Whole slice: commit, push, `claude plugin marketplace update bit-pro`, restart — then invoke
      `/bit:do` and `/bit:plan` and confirm the loaded text carries the capture offer in both. That
      is the verse delivered: the offer reaches the pipeline you actually run. Whether a given run
      acts on it is deliberately not a check here — see the note above.

## Commit (user)
`feat(plugin): offer feedback capture where corrections land`