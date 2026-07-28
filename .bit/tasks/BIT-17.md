---
id: BIT-17
title: Retro notes and a bit_retro skill to evaluate them
status: todo
---
## Why

When a run goes sideways the fix is to stop, repair the scope and plan, and carry on. That
repair is **lossy**: the broken plan is overwritten, so a track that was rewritten three times
mid-flight ends up looking identical to one that ran clean. The evidence of what went wrong is
destroyed by the act of fixing it, which means the same mistake is available to be made again
next cycle.

The existing `bit-retro` skill tried to close this loop by asking the model to reconstruct how a
cycle went after the fact. It plateaued, and the reason is structural rather than a matter of
prompt quality: with the failure evidence already gone, reconstruction produces fluent prose
that is unfalsifiable — nothing in the repo can contradict it — and uncountable, so there is no
way to tell whether a lesson is new or the fifth restatement of one that already got "fixed."

Two failure modes are worth catching. **Divergence**: what got implemented differs from what was
planned. **Gap**: the plan was silent on something the work turned out to require, so a decision
got made during execution that should have been made during planning. Gaps are the more valuable
of the two, because a plan that leaves nothing to decide is exactly the plan a small model can
execute without supervision.

## Summary

Split the loop into capture and evaluation. **Capture** is a `bit` subcommand that appends a
factual observation to a track — which bar, what the plan said, what the work actually required,
what was missing — called from inside `bit_do` and `bit_plan` at the moment a correction lands.
It records observations only; the cause is left blank. **Evaluation** is a new `bit_retro` skill
that reads a finished track's notes against what actually changed, assigns each note a cause
from a small fixed set, and routes it to the stage that should have caught it.

The output of a retro is routing, not insight: each note becomes a plan-format constraint, a
line of always-loaded context, or a hook — or it's marked not preventable and dropped.

## Visual aid

```
  in-flight (cheap, factual)              end of cycle (analysis)

  bit_plan ─┐                             bit_retro
            ├─ correction lands              │ reads: track notes + plan + actual diff
  bit_do  ──┘        │                       │ assigns: cause (fixed set)
                     ▼                       ▼
              bit <note cmd>            routes each note to one of:
                     │                    · a plan-format constraint (validator enforces)
                     ▼                    · a line in always-loaded context
              note on the TRACK           · a hook (mechanically blocks it)
              (cites the bar as data)     · not preventable → dropped
```

## Risks & unknowns

- **Unknown:** What the capture command is called and how a note is stored. Storage is the
  substantive half: a markdown section in the track body is readable and diffs well, but the
  body is prose the skills already rewrite wholesale, so appending has to not clobber. YAML
  frontmatter is structured and safely appendable but invisible in the TUI.
  **Resolve by:** user's call, informed by how `task update -d` currently round-trips a body.
  **De-risk before planning?** Yes — every verse reads or writes notes, so the shape has to be
  settled first. It's also a naming and format choice, which building cannot answer.

- **Unknown:** The cause taxonomy. It must be **small and closed** so recurrence is countable
  across cycles — free prose can't be aggregated, and without counts a retro can't tell whether
  a fix already fired and failed, so it re-derives the same lesson forever. Every cause has to
  name something you would change; if it doesn't, it's a description of what happened, not a
  cause. Starting candidates: *missing reference* → the plan must cite an in-repo exemplar as
  `file:line`; *unasked question* → a question joins the scope or plan checklist; *step too big*
  → a decomposition rule. It must include an explicit **not preventable** bucket, or every note
  reads as a process failure and the checklists bloat with defensive questions for things that
  will never recur.
  **Resolve by:** user's call. Worth drafting against the last few tracks that actually went
  sideways, so the categories come from real failures rather than imagined ones.
  **De-risk before planning?** Yes — it's the vocabulary the retro skill is written against.

- **Risk:** In-flight capture may not fire reliably. The trigger lives inside `bit_do` and
  `bit_plan`, which are already running when a correction lands — that's the whole reason it
  isn't a skill — but it's still an instruction, and instructions get followed with varying
  fidelity. **Resolve by:** Verse 2 puts the trigger in and the next few real cycles show whether
  notes actually appear. If they don't, the fallback is a hook or an end-of-bar prompt, which is
  a bigger change and belongs in its own scope.
  **De-risk before planning?** No — only running real cycles answers it, and Verse 2 is where
  that starts.

## Decisions

- **Capture and evaluation are separate.** Capture has to be cheap enough to happen mid-run;
  evaluation needs the whole cycle in view. Combining them is what made the old skill produce
  reconstruction instead of evidence.
- **Capture is a `bit` subcommand, not a skill.** A skill would have to be invoked at the precise
  moment things went wrong, which is the least reliable moment to ask for a context switch, and
  loading a second instruction set mid-run derails the run being repaired. The *trigger* lives in
  `bit_do` and `bit_plan`, which are already loaded.
- **Notes hang off the track and cite the bar as data** ("happened at BIT-11.4"). Replanning
  renumbers bars, and replanning is frequently the fix itself, so a bar-level note would be
  orphaned by the very event it describes.
- **Capture records observations only; cause is left blank and assigned at retro.** In the moment
  right after being corrected, attribution skews toward blaming the artifact ("the plan was
  unclear") over the model's own choice ("I didn't read the file I was told to read"). Facts are
  cheap and reliable in-flight; judgment isn't.
- **Divergence is captured at the end of each bar from the working diff, not from commit
  history.** The user does the commits, not the agent, so commit shape carries no signal about
  how the work went. Before/after commit hashes are still worth recording as a pointer to the
  range, but the comparison is diff-versus-plan.
- **Evaluation is a new `bit_retro` skill** — the underscore-family port of the hyphen-family
  `bit-retro`, which currently has no counterpart in the CLI-driven set.
- **A retro's output is routing, not insight.** Each note becomes exactly one of: a constraint on
  the plan format that a validator can enforce, a line in always-loaded repo context, or a hook
  that mechanically blocks the mistake. If a note can't become one of those three, the skill says
  so plainly and drops it rather than writing it up as a lesson. Prose that isn't enforced
  anywhere is what the old skill produced.
- **The in-flight trigger is "did I make a decision the plan didn't make for me?"** Sharper than
  "note any problems," and it converges with the real goal: every unspecified decision is either
  a plan gap or something too trivial to plan, and a plan that leaves nothing to decide is one a
  small model can one-shot.
- **A third failure mode is out of reach and the design should not pretend otherwise.** A plan
  can be complete, internally consistent, and simply *wrong* — the wrong pattern, executed
  faithfully. No divergence, no gap, a clean diff, and nothing for capture to see. That is caught
  at the plan review gate by the user, and notes will not find it.
- **Aggregation above `.bit/` is out of scope.** Rolling notes up across repos and clients, and
  the periodic review that reads accumulated counts to rewrite the skills themselves, both need a
  store outside the project directory. Per-track counting inside one repo has to work first.

## Verses

- [ ] Verse 1 — **A correction leaves a durable record instead of vanishing**: a `bit` subcommand
  appends an observation to a track, and the note survives the scope and plan being rewritten
  afterward. Run by hand at first — the value is that repair stops destroying its own evidence.
  Touches: a new command under `cmd/`, and the track storage in `task/`.
- [ ] Verse 2 — **Notes accumulate without anyone remembering to write them**: `bit_do` and
  `bit_plan` call capture at the points where corrections land, so a cycle that went sideways
  ends with notes on the track and a clean cycle ends with none.
  Touches: the `bit_do` and `bit_plan` skill text.
- [ ] Verse 3 — **A finished track can be reviewed against what actually happened**: `bit_retro`
  reads the track, its notes, and the real diff, and routes each note to the stage that should
  have caught it — or marks it not preventable. Turns a pile of observations into specific
  changes to specific artifacts.
  Touches: a new `bit_retro` skill.
- [ ] Verse 4 — **Recurrence is visible across cycles**: each note carries a cause from the fixed
  set, so counts can be read across tracks and a lesson that keeps coming back is distinguishable
  from a new one. This is what makes the loop a loop rather than a sequence of one-off reviews.
  Touches: the note format, and `bit_retro`'s reporting.