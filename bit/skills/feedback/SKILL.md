---
name: bit_feedback
description: Record a correction that landed mid-cycle as a feedback note against the track, so the gap between what the plan said and what the work actually required is captured as evidence instead of being lost. Use whenever the user says "capture that", "record that", "note that for the retro", "add a feedback note", or corrects you with something like "that's wrong, do X instead" — and also reach for it yourself, mid-run, the moment you notice you had to decide something the plan didn't decide for you. It writes one short note into `.bit/feedback/` through the `bp` CLI and stops there — it records an observation and does **not** evaluate, classify, diagnose, or fix anything. If the correction means the plan itself was wrong, that is a separate hand-back to bit_plan — this skill only writes down what happened.
---

# Feedback Capture

You record **one note**: a correction that landed during a cycle, written down as evidence. A note has one job — preserve what the plan said and what the work actually turned out to require, at the moment that gap was visible. Nothing more happens here. You don't decide what caused it, you don't fix it, and you don't revise anything.

Capture is worth doing only if it's cheap. A note the user has to sit down and compose won't get written at the moment it's worth writing, so the whole point of this skill is that **one sentence in produces a well-formed note out** — you supply the track, the bar, and the structure; the user supplies the correction.

**Before you drive the CLI, run `bp instructions`** — the shared command contract (find the track, record a note, read a body). Every write into `.bit/` goes through `bp`; never hand-write a file under `.bit/feedback/`.

---

## When to capture

The trigger question is:

> **Did I make a decision the plan didn't make for me?**

That's sharper than "note anything that went wrong," and it converges with what the notes are for. Every unspecified decision is either a gap in the plan or something too trivial to have planned — and a plan that leaves nothing to decide is a plan a small model could one-shot. Sorting which is which is a later job; noticing that a decision happened is this one.

So the answer being yes is enough. You don't need the correction to have been serious, or to know whether it was preventable.

## 1. Find the track

A note keys to a **track**, not a bar.

Usually the session settles it — the work in flight is the track. Two tracks being open isn't ambiguity; the bar you were working in names its own parent. It's ambiguous when the correction's *subject matter* belongs to a different track than the bar you were in — or when it reaches back into work from an earlier session. In that case list the candidates with `bp task list` and **ask**. Don't guess: a note filed under the wrong track is a note the retro can't use, and the whole value of a note is that a future reader trusts where it came from.

Both active and archived tracks are accepted, so a note about finished work still has somewhere to go.

## 2. Write the note

The body has three parts and nothing else:

- **Where it happened** — cite the bar as prose: `Happened at BIT-11.4.` It's data in the text, not a key, because replanning renumbers bars and replanning is frequently the fix itself.
- **What the plan said** — quote or paraphrase the instruction as written.
- **What the work actually turned out to require** — the decision that got made, or the correction that was needed.

```markdown
Happened at BIT-11.4.

The plan said: fall back to `plugin install` when `plugin update` fails.
The work required: deciding whether the fallback also runs `marketplace add`,
which the plan did not settle.
```

**Observations only.** No cause, no blame, no proposed fix, no lesson learned. This isn't modesty — it's that right after being corrected, attribution is unreliable: it skews toward faulting the artifact ("the plan was unclear") over the choice actually made ("I didn't read the file I was told to read"). Facts are cheap and accurate in flight; judgment isn't. Classifying a note is the retro's job, and it does that job better with clean evidence than with a conclusion it has to unpick.

**Keep it short.** A few sentences. If it's growing past that, you're explaining rather than recording.

## 3. Record it

The body is multi-line prose, so build it in a file and pass it the same way a task body is passed:

```bash
bp feedback add "$TRACK" -d "$(cat note.md)"
```

The command prints the note's path — report that back to the user, so they can see exactly what was written and where.

---

## What this skill does not do

- **Evaluate notes** — no cause, no category, no severity. A note is evidence; reading across notes is a separate cycle.
- **Revise the scope or plan** — if the correction means the plan was wrong, hand back to **bit_plan** (or **bit_scope** if the direction is off). Recording a note repairs nothing, and treating it as a repair is what turns capture into reconstruction instead of evidence.
- **Fix the code** — that's the cycle you were already in. Write the note, then carry on with it.
- **Read, rewrite, or delete a note** — `add` is the only write. A new note can never damage one already recorded.
- **Hand-write `.bit/feedback/*.md`** — the note goes in through `bp feedback add`.
