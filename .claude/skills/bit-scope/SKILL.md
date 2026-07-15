---
name: bit-scope
description: Create or refine a high-level scope (an RFC-style overview) for a feature or change BEFORE any detailed planning. Use this first — whenever the user wants to frame WHAT is changing and WHY at a high level, sketch the shape of a feature request, decide the order of delivery, or surface risks and unknowns to de-risk before committing to a detailed plan. Triggers on "scope this out", "give me a high-level overview", "what's the shape of this feature", "let's write an RFC", "think through the delivery order", "before we plan", or describing a feature request where implementation detail isn't wanted yet. Produces a markdown scope doc: the motivation (WHY), a coarse checklist of value-delivering phases (each a usable vertical slice, ordered for incremental value), light pointers to the code areas each phase touches, and a highlighted risks/unknowns section. This is the overview that feeds bit-plan — bit-scope owns the WHY and the delivery order; bit-plan turns each phase into detailed TDD steps; bit-do executes. Reach for bit-scope for the high-level shape, bit-plan for the detailed plan, bit-do to build it.
---

# Scope Creator

You write and refine a **scope** — a short, high-level overview of a proposed software change. You do NOT write code, name functions, or produce a granular task list; a later skill (bit-plan) does that. Your job is clarity about **what** is changing, **why**, and **in what order** value gets delivered.

A scope is a markdown file the user refines until they're happy with the shape of the work. It is the first of three artifacts:

- **bit-scope** (this skill) — the high-level shape: why, and the order of delivery. Owns the WHY.
- **bit-plan** — turns the scope's phases into detailed, contradiction-driven TDD steps.
- **bit-do** — executes the plan, keeping both the scope and the plan in sync as progress lands.

Because bit-scope owns the WHY, the plan won't repeat it — it will point back here. That makes the motivation in this document load-bearing: get it right.

---

## Two modes

**Create** — start from a feature request or problem description and build the scope from scratch.
**Refine** — improve an existing scope (the user points to a file with @, or pastes it). The user will typically loop here several times, tightening the phases and de-risking, until they're satisfied enough to move to bit-plan.

---

## The altitude: value, not implementation

The whole point of a scope is to reason about **delivery**, not construction. Hold this altitude carefully — it's the most common way a scope goes wrong (it drifts into being a mini-plan).

### What "value" means

A phase delivers value when, after it lands, **someone can do something they couldn't before** — end to end, however narrow. It's a *vertical slice* (thin but complete: input → behavior → observable result), not a *horizontal layer* (all of the database work, usable by no one yet).

The litmus test for every phase:

> **Could a real user or operator exercise this phase and get a benefit — even a tiny one?**

If yes, it's an increment of value and belongs in the scope. If the phase only produces internal plumbing nobody can touch yet, it's a *task* — and tasks belong in the plan, not here.

### How this shapes the phases

- **Order by value and risk, not by architecture.** Sequence phases so each one is usable and each builds on the last. Start with a *walking skeleton* — the thinnest thing that works end to end — then widen. Avoid "layer 1, layer 2, layer 3" orderings where nothing is usable until the last layer. If a risky assumption sits underneath everything, an early phase should be the one that validates it, so failure is cheap.
- **Name phases by the capability unlocked, not the component built.** "User can search voters by district" — not "Add a search index." The reader should see the shape of *what becomes possible*, in what order.
- **Deliberately coarse.** A handful of phases, each a sentence or two. If you're tempted to write five sub-bullets about how a phase works, you've dropped into plan altitude — pull back up.

### The line you don't cross

The scope says *what becomes possible and in what order*. It never says *how*. No function signatures, no test strategy, no algorithms, no schemas. Those are bit-plan's job, and putting them here just creates a second place that drifts out of sync.

**One deliberate exception — a light "touches" pointer.** Each phase may name the *code area* it affects — a file, a module, a component — as a **locator so the reader can spot-check that work is on the right path**. This is a "where to look," not a "how to build it."

- In: `Touches: the ETL aggregation step (district_rollup.py)`
- Out: `add a GROUP BY on district_id and switch the write to an UPSERT`

If you can't name the area without prescribing the change, leave it vague ("the reporting layer") rather than sliding into implementation.

---

## Risks & unknowns are first-class

A scope is the cheapest place to discover what you don't know. Surfacing an unknown here — before a detailed plan exists — lets the user go de-risk it (a spike, an experiment, a question answered) so the plan starts from as much certainty as possible.

So don't treat risks as an afterthought. For each one, pair it with **what would resolve it**, and flag whether it's worth de-risking *before* planning:

```markdown
- **Unknown:** Does the county API return district codes, or only county codes?
  **Resolve by:** 30-min spike hitting the staging endpoint with 3 sample counties.
  **De-risk before planning?** Yes — the whole aggregation approach depends on this.
```

A risk with no path to resolution is just an anxiety; a risk with a resolution is a task the user can act on. Prefer the latter.

---

## Gathering context (new scopes)

Before drafting, get the WHY right — it's the part the plan will lean on, so it has to stand on its own. Ask:

1. What's the problem or goal, in user/operator terms — what breaks, is missing, or is painful today? (Not a restatement of the change.)
2. What triggered this now — a bug report, wrong data, a deadline, a new requirement?
3. Any constraints — things we must not touch, production concerns, ordering forced by external dependencies?

Then do *light* research — enough to name the code areas each phase touches and to spot the real risks, but not a deep dive (that's bit-plan's job):
- Locate the parts of the codebase each phase would affect, so the "touches" pointers are accurate.
- Notice genuine unknowns — external services, ambiguous data shapes, assumptions the whole approach rests on.

Don't over-research. If you find yourself reading function bodies to design the change, you've gone past scope altitude.

---

## Scope format

Write the scope to a `.md` file in the project root unless the user says otherwise. A clear default name is `<feature>-scope.md`, so bit-plan and bit-do can find its partner plan (`<feature>-plan.md`). Use this structure:

```markdown
# [Title]

## Why
[The motivation. What breaks, is missing, or is painful today — business or user impact,
not a restatement of the change. 2–4 sentences. This is the one home for the WHY; the
plan will point back here.]

## Summary
[The change in a few high-level sentences.]

## Phases
[A coarse, markable checklist. Each phase is a usable vertical slice, named by the
capability it unlocks, ordered for incremental value. bit-do checks these off (- [x])
as the underlying plan steps land.]

- [ ] Phase 1 — <capability unlocked>: one or two sentences on what a user/operator can now do.
  Touches: <code area / files> — where to look to verify.
- [ ] Phase 2 — <capability unlocked>: …

## Visual aid
[Where it clarifies the shape, an ASCII diagram or a ```mermaid``` block — data flow,
component relationships, or before/after. Skip if it wouldn't add clarity.]

## Risks & unknowns
[Each unknown paired with how to resolve it and whether to de-risk before planning.
Omit the section only if there genuinely are none.]

- **Unknown:** …
  **Resolve by:** …
  **De-risk before planning?** Yes / No — why.
```

Keep it tight. This is an overview a reader skims to grasp the shape of the work before a detailed plan exists. Prefer clarity over completeness.

---

## Refining an existing scope

1. Read the whole scope first.
2. Check the WHY: does it say *why*, not *what*? Would a reader who knows nothing about the codebase understand the motivation? If not, flag it and offer a rewrite — this is the section the plan depends on.
3. Check the phases at value altitude:
   - Is each phase a usable vertical slice (passes the litmus test), or is it a horizontal layer / internal task that belongs in the plan?
   - Is the order delivering incremental value — walking skeleton first, riskiest assumptions early?
   - Is each named by the capability unlocked, not the component built?
4. Check that no phase has slid into implementation detail. The "touches" pointer is a locator only — flag anything prescribing *how*.
5. Check risks: is each unknown paired with a resolution? Are the ones worth de-risking before planning called out?
6. Propose edits with reasoning — don't silently rewrite large sections. Confirm before rewriting more than a few lines.

The user drives this loop; keep refining with them until they're happy enough to move to bit-plan.

---

## Handoff to bit-plan

When the user is satisfied with the scope, the next step is bit-plan, which turns each phase into detailed TDD steps. Remind them, briefly, that:
- The plan will reference this scope for the WHY rather than repeating it.
- Each plan step will be tagged with the scope phase it serves, so progress rolls up.
- bit-do later keeps both documents in sync as work lands.

You don't build the plan — just point them at bit-plan when they're ready.
