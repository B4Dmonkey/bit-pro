---
id: BIT-17
title: Capture feedback notes when a correction lands
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

Two failure modes are worth recording. **Divergence**: what got implemented differs from what
was planned. **Gap**: the plan was silent on something the work turned out to require, so a
decision got made during execution that should have been made during planning. Gaps are the more
valuable of the two, because a plan that leaves nothing to decide is exactly the plan a small
model can execute without supervision.

## Summary

This track builds **capture only** — the evidence side of the loop. A correction becomes a
**note**: a small markdown file in `.bit/feedback/`, written through `bp`, recording which track
and bar it happened at, what the plan said, and what the work actually turned out to require.

Notes are recorded three ways, in increasing convenience: by hand with a `bp` subcommand, by
invoking a `bit_feedback` skill that turns a correction into a well-formed note, and by `bit_do`
and `bit_plan` *recommending* capture at the moment a correction lands so the user answers yes or
no rather than having to remember.

Evaluating notes is explicitly **not** in this track. `bit_retro` is untouched here; it becomes
the consumer of `.bit/feedback/` in a later scope. What this track has to get right is that the
notes exist, survive, and are worth reading when that consumer arrives.

## Visual aid

```
  in-flight (cheap, factual)                     later, separate scope

  a correction lands
         │
         ├── user runs /bit:feedback ────┐
         │                              │
         └── bit_do / bit_plan           │
             recommend it → yes/no ──────┤
                                         ▼
                                  bp feedback add
                                         │
                                         ▼
                              .bit/feedback/BIT-17-001.md      ┄┄▶  bit_retro
                              (create-only; one file per note)       (not built here —
                                                                      reads these later)
```

## Decisions

- **Capture and evaluation are separate tracks of work.** Capture has to be cheap enough to
  happen mid-run; evaluation needs the whole cycle in view. Combining them is what made the old
  skill produce reconstruction instead of evidence. This track is capture; `bit_retro` is not
  touched.
- **Notes live in `.bit/feedback/`, one file per note.** A new note is a file *create*, so it can
  never damage a note already recorded — which matters because capture fires at the least
  reliable moment in the cycle, right after a run went wrong. It also keeps notes out of the
  track body, which the skills rewrite wholesale via `task update -d`. Cost accepted: a folder of
  many small files reads worse by hand than one document per track would.
- **A note's filename carries the track and a sequence: `BIT-17-001.md`.** The track ID keys the
  note to its work; the sequence gives deterministic ordering without depending on filesystem
  order. Picking the next number is a folder scan, the same pattern `task/store.go` already uses
  for task IDs.
- **The write path is a `bp` subcommand — `bp feedback add`.** Every write into `.bit/` goes
  through `bp` and nothing hand-edits files under it; feedback is no exception. The skill decides
  *what* the note says, the CLI owns *where and how* it's stored.
- **Capture is a skill the user invokes, and agents may recommend it.** `/bit:feedback` is called
  deliberately — the user is the one who knows a correction just happened. `bit_do` and `bit_plan`
  suggest calling it when they see a correction land, and the user answers yes or no. Nothing
  records a note autonomously: a skill that fires itself mid-run would derail the run being
  repaired, and the yes/no keeps the human as the judge of what counts as feedback.
- **The recommendation is best-effort and that is accepted.** Whether `bit_do` and `bit_plan`
  actually notice a correction worth a note is an instruction, followed with varying fidelity, and
  no amount of building settles it. It doesn't need to be reliable: the user invoking
  `/bit:feedback` explicitly is the path that always works, and the recommendation only saves them
  the trouble of remembering. So a missed recommendation is not a defect, and nothing in this
  track needs to prove the trigger fires — no hook, no enforcement.
- **Notes key to the track and cite the bar as data** ("happened at BIT-11.4"). Replanning
  renumbers bars, and replanning is frequently the fix itself, so a note keyed to a bar would be
  orphaned by the very event it describes.
- **Archiving a track leaves its notes in place.** `task archive` relocates files within
  `.bit/tasks/` → `.bit/archive/`; `.bit/feedback/` is untouched by it. That's the behavior we
  want, not an accident to tidy up: retro reads *finished* tracks, so notes have to outlive the
  track's active life.
- **A note records observations only. There is no cause field yet.** In the moment right after
  being corrected, attribution skews toward blaming the artifact ("the plan was unclear") over
  the model's own choice ("I didn't read the file I was told to read"). Facts are cheap and
  reliable in-flight; judgment isn't. Classifying a note is retro's job, so the field arrives with
  retro rather than being guessed at now.
- **The trigger question is "did I make a decision the plan didn't make for me?"** Sharper than
  "note any problems," and it converges with the real goal: every unspecified decision is either a
  plan gap or something too trivial to plan, and a plan that leaves nothing to decide is one a
  small model can one-shot.
- **Divergence is observed from the working diff, not from commit history.** The user does the
  commits, not the agent, so commit shape carries no signal about how the work went.
- **A third failure mode is out of reach and the design should not pretend otherwise.** A plan can
  be complete, internally consistent, and simply *wrong* — the wrong pattern, executed
  faithfully. No divergence, no gap, a clean diff, and nothing for capture to see. That is caught
  at the plan review gate by the user, and notes will not find it.
- **Aggregation above `.bit/` is out of scope.** Rolling notes up across repos and clients, and
  the periodic review that reads accumulated counts to rewrite the skills themselves, both need a
  store outside the project directory. Per-repo capture has to work first.

## Verses

- [ ] Verse 1 — **A correction leaves a durable record instead of vanishing**: `bp feedback add`
  writes a note into `.bit/feedback/`, and the note survives the scope and plan being rewritten
  afterward and the track being archived. Run by hand at first — the value is that repair stops
  destroying its own evidence.
  Touches: a new command under `cmd/`, and the store in `task/`.
- [ ] Verse 2 — **A correction becomes a good note without the user composing it**: invoking
  `/bit:feedback` turns "that's wrong, do X instead" into a well-formed note — the right track and
  bar, what the plan said, what the work actually required — so recording one costs a sentence
  rather than a writing exercise.
  Touches: a new `bit_feedback` skill under the plugin's `skills/`.
- [ ] Verse 3 — **The user gets offered the note instead of having to think of it**: `bit_do` and
  `bit_plan` recommend capture at the points where corrections land, so a yes is all it takes.
  Convenience on top of Verses 1 and 2, which already work without it.
  Touches: the `bit_do` and `bit_plan` skill text.