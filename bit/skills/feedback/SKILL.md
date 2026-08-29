---
name: bit_feedback
description: Record a correction that landed mid-cycle as a feedback note against the track, so the gap between what the plan said and what the work actually required is captured as evidence instead of being lost. Use whenever the user says "capture that", "record that", "note that for the retro", "add a feedback note", or corrects you with something like "that's wrong, do X instead" — and also reach for it yourself, mid-run, the moment you notice you had to decide something the plan didn't decide for you. It writes one short note into `.bit/feedback/` through the `bp` CLI and stops there — it records an observation and does **not** evaluate, classify, diagnose, or fix anything. If the correction means the plan itself was wrong, that is a separate hand-back to bit_plan — this skill only writes down what happened.
---

# Feedback Capture

You record **one note**: a correction that landed during a cycle, written down as evidence. A note has one job — preserve the scope and plan text that bears on the correction, the actual exchange that surfaced it, and what was discovered, so a reader who wasn't in the session can see the whole gap without reopening the track. Nothing more happens here. You don't decide what caused it, you don't fix it, and you don't revise anything.

Capture is worth doing only if it's cheap **for the user** — a note they have to sit down and compose won't get written at the moment it's worth writing. That cheapness is about their input, not the note's length: **one sentence in produces a complete note out** — you gather the scope text, the plan text, and the actual exchange from what's already in front of you; the user only ever supplies the correction.

Three tools cover everything this skill does: `mcp__bit__task_list` and `mcp__bit__task_read` to find the track and read what it says, and `mcp__bit__feedback_add` to record the note. The old prohibition on hand-writing a file under `.bit/feedback/` still holds, and it holds harder now — `feedback_add` isn't the preferred way in, it's the only one.

---

## When to capture

The trigger question is:

> **Did I make a decision the plan didn't make for me?**

That's sharper than "note anything that went wrong," and it converges with what the notes are for. Every unspecified decision is either a gap in the plan or something too trivial to have planned — and a plan that leaves nothing to decide is a plan a small model could one-shot. Sorting which is which is a later job; noticing that a decision happened is this one.

So the answer being yes is enough. You don't need the correction to have been serious, or to know whether it was preventable.

## 1. Find the track

A note keys to a **track**, not a bar.

Usually the session settles it — the work in flight is the track. Two tracks being open isn't ambiguity; the bar you were working in names its own parent. It's ambiguous when the correction's *subject matter* belongs to a different track than the bar you were in — or when it reaches back into work from an earlier session. In that case list the candidates with `mcp__bit__task_list` and **ask**. Don't guess: a note filed under the wrong track is a note the retro can't use, and the whole value of a note is that a future reader trusts where it came from.

Active, completed, and archived tracks are all accepted, so a note about finished work still has somewhere to go.

## 2. Write the note

A note gathers whatever a future retro needs to see the whole gap — it isn't a fixed number of one-liners to fill in. Ask, for each note: cold, with no memory of this session, what would someone need to read to know what went wrong and where to look? That's usually four kinds of material:

- **Where it happened** — cite the track and, if there is one yet, the bar, as prose: `Happened at BIT-11.4, under track BIT-11.` Data in the text, not a key, because replanning renumbers bars and replanning is frequently the fix itself.
- **The relevant scope and plan text, quoted verbatim** — whatever part of the track's WHY/Decisions/Risks and the bar's body actually bears on the correction. Quote it directly; don't describe it from memory. Paraphrasing loses the exact wording a retro would need to check the claim without reopening the track. If what's being corrected is a verdict another skill's own mechanism produced — not something this session decided — name that mechanism and quote what it produced. A retro can't revisit a behavior it was never told the name of.
- **The exchange itself, quoted verbatim** — the actual question and correction as they were said, not a summary of them. A summary is already someone's interpretation of what mattered; the real exchange lets the retro form its own read of it instead of inheriting yours.
- **What was discovered, and what the work actually required** — the concrete fact that surfaced, stated plainly: what existed, what was checked, what was chosen, and the correction that followed. Resist the pull to generalize this into "the underlying issue is X" — noticing that several notes share an underlying issue takes reading across all of them, which is bit_retro's job. A single note reaching for that generalization on its own is doing the retro's job with one data point, and it's exactly the same move as writing a cause: it substitutes your inference for the fact.

```markdown
Happened at BIT-11.4, under track BIT-11 ("Ballot image batch rename").

Scope (BIT-11) says: "Batch-rename a directory of scanned ballot images to
the project's file-naming convention." No tooling constraint is named
anywhere in the track.

Bar BIT-11.4 says: "Rename every file under `incoming/` to
`<county>-<precinct>-<seq>.tif` using the county/precinct parsed from each
filename's existing prefix." It doesn't ask whether existing tooling should
be reused first either.

Exchange:
> Assistant: Done — I wrote a small Python script that parses each filename
> with a regex and renames the file to the convention.
> User: this project already has `scripts/rename_batch.sh` for exactly this —
> you should have used that instead of writing a new script.

The work required using `scripts/rename_batch.sh`, which the project
already has for exactly this rename, instead of the new Python script that
was written.
```

The same shape covers a correction to another skill's mechanism, even before a bar exists:

```markdown
Happened during bit_scope work on BIT-3's risk review (no bar yet — this
surfaced before bit_plan).

Scope (BIT-3) says: Phases 1-2 build a nightly export report's core
sections; Phase 3 is a spike into locale-aware currency formatting,
contingent on the upstream feed carrying a currency code. Its Risks section
reads: "De-risk before planning? No — nothing in Verses 1–3 depend on the
answer."

bit_scope's downstream-impact test produced that verdict: it asks whether
anything already planned or built would need to be redone once the risk
resolves, and judged this one non-blocking because Phase 3 is purely
additive.

Exchange:
> Assistant: [recommends moving to bit_plan, citing the reasoning above]
> User: why should we not derisk before planning?
> Assistant: [restates the downstream-impact reasoning]
> User: i want all risk resolved before planning.

The work required: the user's standard is that all open risk gets resolved
before planning, regardless of what the downstream-impact test concluded.
```

**Observations only.** No cause, no blame, no proposed fix, no lesson learned, and no naming the underlying pattern behind the instance — "the plan should have asked whether to reuse existing tooling" is the retro's conclusion, drawn from reading several notes together, not something to write into any single one. This isn't modesty — it's that right after being corrected, attribution is unreliable: it skews toward faulting the artifact ("the plan was unclear") over the choice actually made ("I didn't read the file I was told to read"), and a single instance is too little evidence to generalize from even when the attribution is right. Facts are cheap and accurate in flight; judgment isn't. Classifying a note — and reading across many of them for the pattern — is the retro's job, and it does that job better with clean evidence than with a conclusion it has to unpick.

**Brevity is not the goal.** A note missing the scope text, the plan text, or the actual exchange isn't short — it's incomplete, and the retro will have to reopen the track and reconstruct the conversation to do the job the note exists to save it from doing. The one thing that stays cheap is what the user has to supply: they give the correction, in one sentence if that's all it takes. Gathering everything else the note needs is your job, not theirs.

## 3. Record it

Before running this, state the note (or its gist) and get an explicit go-ahead — even when you noticed the gap yourself and nobody asked. Noticing it yourself means drafting and offering it, not writing it unasked.

Then call `mcp__bit__feedback_add` with `track` and `body`. The body is simply the note's prose — you pass the markdown itself, so there's nothing to stage and nothing to escape, and the quoted material a good note is mostly made of survives exactly as written.

The tool returns the note's `path` — report it back to the user, so they can see exactly what was written and where.

---

## What this skill does not do

- **Evaluate notes** — no cause, no category, no severity. A note is evidence; reading across notes is a separate cycle.
- **Revise the scope or plan** — if the correction means the plan was wrong, hand back to **bit_plan** (or **bit_scope** if the direction is off). Recording a note repairs nothing, and treating it as a repair is what turns capture into reconstruction instead of evidence.
- **Fix the code** — that's the cycle you were already in. Write the note, then carry on with it.
- **Read, rewrite, or delete a note** — `add` is the only write. A new note can never damage one already recorded.
- **Hand-write `.bit/feedback/*.md`** — the note goes in through `mcp__bit__feedback_add`.
