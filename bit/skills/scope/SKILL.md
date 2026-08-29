---
name: bit_scope
description: Create or refine a high-level scope (an RFC-style overview) for a feature or change BEFORE any detailed planning. Use this first — whenever the user wants to frame WHAT is changing and WHY at a high level, sketch the shape of a feature request, decide the order of delivery, or surface risks and unknowns to de-risk before committing to a detailed plan. Triggers on "scope this out", "give me a high-level overview", "what's the shape of this feature", "let's write an RFC", "think through the delivery order", "before we plan", or describing a feature request where implementation detail isn't wanted yet. Also use it when a **spike has come back with a result** — "the spike worked", "here's what we learned", "the plugin update does/doesn't apply" — because the answer turns an unknown into a decision and the verses that depended on it have to be revised here before they can be planned. Authors the scope as a track in `.bit/` through the `bp` CLI — the motivation (WHY), a coarse checklist of value-delivering verses (each a usable vertical slice, ordered for incremental value), light pointers to the code areas each verse touches, the open risks/unknowns, and the decisions that iteration has settled. This is the overview that feeds bit_plan — bit_scope owns the WHY and the delivery order; bit_plan turns each verse into detailed TDD steps; bit_do executes. Reach for bit_scope for the high-level shape, bit_plan for the detailed plan, bit_do to build it.
---

# Scope Creator

You write and refine a **scope** — a short, high-level overview of a proposed software change. You do NOT write code, name functions, or produce a granular task list; a later skill (bit_plan) does that. Your job is clarity about **what** is changing, **why**, and **in what order** value gets delivered.

A scope lives as a **track** in `.bit/` — a top-level task whose body holds the scope prose — authored and refined through the local `bp` CLI. The user refines it until they're happy with the shape of the work. It is the first of three artifacts:

- **bit_scope** (this skill) — the high-level shape: why, and the order of delivery. Owns the WHY.
- **bit_plan** — turns the scope's verses into detailed, contradiction-driven TDD steps, one **bar** (child task) per step under this track.
- **bit_do** — executes the plan, moving each bar's status and rolling the track up as progress lands.

A **verse** is one value slice in the delivery order (the checklist you write here); bit_plan tags each bar to the verse it serves with the `--phase`/`--phase-label` flags (the CLI flag keeps the name `phase`; the scope calls the slice a verse), and bit_do checks a verse off once all its bars are done.

Because bit_scope owns the WHY, the plan won't repeat it — the bars live under this track, so a reader gets the WHY by reading the track body. That makes the motivation in this document load-bearing: get it right.

Three tools cover everything this skill does: `mcp__bit__task_create` to mint the track, `mcp__bit__task_read` to read a body back, and `mcp__bit__task_update` to write a refined one. Never hand-edit `.bit/tasks/*.md` — that rule is the one the whole tool surface exists to enforce.

---

## Two modes

**Create** — start from a feature request or problem description and build the scope from scratch as a new track.
**Refine** — improve an existing scope. The user names the track (by ID like `BIT-7`, or by title); read its body with `mcp__bit__task_read`, taking the `body` field, then write the refined body back. The user will typically loop here several times, tightening the verses and resolving unknowns into decisions, until they're satisfied enough to move to bit_plan. Refine is also where a **spike's result** lands: the user runs the spike, comes back with the answer, and the scope is revised around it — see *Refining after a spike*.

---

## The altitude: value, not implementation

The whole point of a scope is to reason about **delivery**, not construction. Hold this altitude carefully — it's the most common way a scope goes wrong (it drifts into being a mini-plan).

### What "value" means

A verse delivers value when, after it lands, **someone can do something they couldn't before** — end to end, however narrow. It's a *vertical slice* (thin but complete: input → behavior → observable result), not a *horizontal layer* (all of the database work, usable by no one yet).

The litmus test for every verse:

> **Could a real user or operator exercise this verse and get a benefit — even a tiny one?**

If yes, it's an increment of value and belongs in the scope. If the verse only produces internal plumbing nobody can touch yet, it's a *task* — and tasks belong in the plan, not here.

### How this shapes the verses

- **Order by value and risk, not by architecture.** Sequence verses so each one is usable and each builds on the last. Start with a *walking skeleton* — the thinnest thing that works end to end — then widen. Avoid "layer 1, layer 2, layer 3" orderings where nothing is usable until the last layer. If a risky assumption sits underneath everything, an early verse should be the one that validates it, so failure is cheap.
- **Name verses by the capability unlocked, not the component built.** "User can search voters by district" — not "Add a search index." The reader should see the shape of *what becomes possible*, in what order.
- **Deliberately coarse.** A handful of verses, each a sentence or two. If you're tempted to write five sub-bullets about how a verse works, you've dropped into plan altitude — pull back up.

### The line you don't cross

The scope says *what becomes possible and in what order*. It never says *how*. No function signatures, no test strategy, no algorithms, no schemas. Those are bit_plan's job, and putting them here just creates a second place that drifts out of sync.

**One deliberate exception — a light "touches" pointer.** Each verse may name the *code area* it affects — a file, a module, a component — as a **locator so the reader can spot-check that work is on the right path**. This is a "where to look," not a "how to build it."

- In: `Touches: the ETL aggregation step (district_rollup.py)`
- Out: `add a GROUP BY on district_id and switch the write to an UPSERT`

If you can't name the area without prescribing the change, leave it vague ("the reporting layer") rather than sliding into implementation.

---

## Risks and unknowns become decisions

A scope is the cheapest place to discover what you don't know — and the place to drive those unknowns to rest before a detailed plan exists.

Keep the distinction sharp:

- An **unknown** is any genuinely-open question the approach rests on. Three kinds all count: a *choice the user still has to make* (a name, some wording, which columns), an *ambiguous data shape*, or a *technical-feasibility assumption* — does this library or technique actually do what the approach needs? Who answers it doesn't matter — the user, a spike, or an early verse that proves it by building. What matters is that it's open and load-bearing, so it gets named here rather than discovered late. Pair each one with **what would resolve it** and whether it's worth resolving *before* planning:

  ```markdown
  - **Unknown:** Does the county API return district codes, or only county codes?
    **Resolve by:** 30-min spike hitting the staging endpoint with 3 sample counties.
    **De-risk before planning?** Yes — the whole aggregation approach depends on this.
  ```

- A **decision** is a question that's been *answered* — the user made a call, or a spike came back. It's settled.

**The refine loop drains unknowns into decisions.** The moment an unknown gets answered, it stops being an unknown — so move it out of Risks & unknowns and into Decisions. Don't leave it in the risk list re-labelled "Resolved": an answered question sitting in the unknowns section is clutter, and it hides what's actually still open behind a wall of settled ones. Each pass should make the risk list *shorter* and the decision list *longer*.

An unknown has exactly two honest resting places: **answered now** — it becomes a Decision — or **named in Risks, tied to the specific early verse that resolves it by building** (walking-skeleton-first, so a wrong assumption fails cheap and early). The target state at handoff is a **Risks & unknowns section that's empty or close to it**: every question either settled into a Decision or, at most, a named risk a first verse is about to prove.

### Spikes

Both resting places involve exploring the unknown; the difference is *when*, and it decides whether a spike is a verse at all:

- **An inline probe** you can run during this scoping session — hit the staging endpoint, read the library's source, try the command. Cheap and immediate. Run it, then write the **Decision**. It never becomes a verse, and leaving it as one wastes a whole planning cycle on something a five-minute check would have settled. Reach for this first, always.
- **A spike verse** — only building the thing can answer it. This becomes a real verse in the delivery order, and because a wrong assumption should fail cheap, it goes first.

**Mark a spike verse as one:** `- [ ] Verse 1 (spike) — …`. The marker is load-bearing because bit_plan treats spikes differently — a spike's deliverable is an *answer*, so it can't be planned as a TDD step, and verses downstream of it often can't be planned at all until the answer is in. Unmarked, bit_plan reads it as ordinary work and writes confident detailed steps for work whose shape isn't known yet.

The verse line itself stays as coarse as any other — the detail lives in the **Risks & unknowns** entry, beside `Resolve by`, so there's one home for the question rather than two that drift. That entry needs two things an ordinary unknown doesn't:

- **The question phrased so an answer is recognisable** — what you'd see if yes, what you'd see if no. "See whether the update mechanism works" isn't a question; "does pushing a skill change deliver it to a repo that already has the plugin installed" is. Vague questions produce spikes that mill around and conclude "seems fine," which is the same as not running one.
- **What its answer could change** — name the downstream verses whose shape depends on it, or say plainly that none do. This is the fact bit_plan needs and cannot derive from anywhere else: if later verses hang on the answer, it plans the spikes and stops; if they don't, this scope is doing two independent jobs and probably wants splitting. Both readings lead somewhere expensive if wrong, so state it even when it feels obvious.

Also say whether the spike's artifact is **kept or thrown away**. It changes how carefully the thing gets built, and a spike that's vague about it is how throwaway code ends up load-bearing.

What you must never do is **dissolve an open question into a verse's prose** — "how X is composited is deferred to plan-time," "the technical question is left for later." That reads like a resolution but hides an unknown where nothing tracks it. And it misreads bit_plan: **bit_plan is not a discovery phase.** It turns a *settled* shape into TDD steps; it has no mechanism to answer an open feasibility question, so "defer to plan-time" just launders the question downstream into a bogus step or a fake "User verifies" check — the very trap the design-choices note below warns about. If a quick spike can answer it, run the spike and write the Decision. If only building can prove it, say exactly that *in Risks*, pointed at the verse that proves it — don't bury it in the verse's description.

**Decisions double as acceptance criteria.** They're the constraints bit_plan and bit_do must honour — the things that have to hold true for the work to be right. Writing them down here is what keeps a downstream plan from quietly relitigating a call the user already made.

**Design choices are decisions too.** A naming call (is the command `archive` or `stash`?), user-facing wording, the shape of stored data, an ID strategy — these are choices, and a choice left open here doesn't vanish; it resurfaces downstream as a bogus "User verifies: the command reads naturally" checkbox, because bit_plan had nowhere to defer to and launders the open question into a fake check. If bit_plan hands one back — "this reads like a choice, not a check" — that's the signal: settle it here as a Decision. Deciding it once, in the open, is what stops it from being relitigated as a fake verification later.

---

## Gathering context (new scopes)

Before drafting, get the WHY right — it's the part the plan will lean on, so it has to stand on its own. Ask:

1. What's the problem or goal, in user/operator terms — what breaks, is missing, or is painful today? (Not a restatement of the change.)
2. What triggered this now — a bug report, wrong data, a deadline, a new requirement?
3. Any constraints — things we must not touch, production concerns, ordering forced by external dependencies?

Then do *light* research — enough to name the code areas each verse touches and to spot the real risks, but not a deep dive (that's bit_plan's job):
- Locate the parts of the codebase each verse would affect, so the "touches" pointers are accurate.
- Notice genuine unknowns — external services, ambiguous data shapes, assumptions the whole approach rests on.
- **Capture any reference docs the user provided.** If the user shared a file path, URL, pasted spec, design doc, or named any external artifact as an authority this scope should honor, record it in the `## References` section of the track body. A reference doc is something the user is pointing to as a source of truth — not every file mentioned casually, only the ones they're invoking as constraints or specifications.

Don't over-research. If you find yourself reading function bodies to design the change, you've gone past scope altitude.

---

## Scope format

The scope is authored as a **track body**. Draft the body (the markdown below), then create the track with it in one call: `mcp__bit__task_create`, passing the scope title as `title` and the drafted prose directly as `body`. It returns the new track's `id`, which is how bit_plan and bit_do later find this work.

Refining an existing track means reading its body, editing, and writing it back with `mcp__bit__task_update`. Report the track ID to the user — it's the handle they'll name when they move to bit_plan.

The order below is deliberate: the **pitch** first (why, what, and a picture of it), then the **problems still open**, then **what's been settled**, and only last **how the work breaks up**. A reader should be sold on the change and know what's undecided before they read the delivery order.

The track's **title** is the scope title; the **body** is this structure (no leading `# Title` needed — the title lives in the track, not the body):

```markdown
## Why
[The motivation. What breaks, is missing, or is painful today — business or user impact,
not a restatement of the change. 2–4 sentences. This is the one home for the WHY; the
plan will point back here.]

## Summary
[The change in a few high-level sentences.]

## Visual aid
[Where it clarifies the shape, an ASCII diagram or a ```mermaid``` block — data flow,
component relationships, or before/after. Skip if it wouldn't add clarity.]

## Risks & unknowns
[Only questions that are genuinely still open — things the user still has to answer.
Each paired with how to resolve it and whether to de-risk before planning. As iteration
answers them, they move down into Decisions. Aim to empty this section before handoff;
omit it entirely once nothing is open.]

- **Unknown:** …
  **Resolve by:** …
  **De-risk before planning?** Yes / No — why.

[When only building can answer it, `Resolve by` names the spike verse and adds two lines —
this is what bit_plan reads to decide how much of the scope it can plan:]

- **Unknown:** …
  **Resolve by:** Verse N (spike). Answer is yes if <observation>; no if <observation>.
  **Downstream:** Verses N+1…M depend on the answer / nothing downstream depends on it.
  **Artifact:** kept and built on / thrown away.
  **De-risk before planning?** Yes — <why>.

## Decisions
[The calls that are settled — often unknowns that iteration resolved. These are the
acceptance criteria the plan and build must honour. One line each: what it commits to,
and briefly why.]

- **<the decision>.** <what it commits to and, briefly, why.>

## Verses
[A coarse, markable checklist. Each verse is a usable vertical slice, named by the
capability it unlocks, ordered for incremental value. bit_do checks these off (- [x])
as the underlying plan steps land — it keys on the `Verse N` text, so keep the checkbox
and `Verse N` on one line. A verse that exists to answer an unknown is marked `(spike)`
after its number; its question and downstream impact live in Risks & unknowns.]

- [ ] Verse 1 (spike) — <the question settled>: one or two sentences on what gets built to answer it.
  Touches: <code area / files> — where to look to verify.
- [ ] Verse 2 — <capability unlocked>: one or two sentences on what a user/operator can now do.
  Touches: <code area / files> — where to look to verify.

## References
[Omit this section if no reference docs were provided. List only external artifacts the
user explicitly pointed to as authorities — specs, design docs, API references, pasted
content. One line each: the path or URL, and which verses it informs.]

- `path/to/doc.md` — what this doc is and which verses it informs
```

Keep it tight. This is an overview a reader skims to grasp the shape of the work before a detailed plan exists. Prefer clarity over completeness.

---

## Refining an existing scope

1. Read the whole scope first.
2. Check the WHY: does it say *why*, not *what*? Would a reader who knows nothing about the codebase understand the motivation? If not, flag it and offer a rewrite — this is the section the plan depends on.
3. Check the risks and decisions — this is where refinement does its real work:
   - Is anything in Risks & unknowns actually *answered*? If so, it's a decision — move it down into Decisions, don't leave it re-labelled "Resolved" in the risk list.
   - Does anything still in Risks & unknowns have a resolution path and a de-risk-before-planning call?
   - Are the Decisions phrased as acceptance criteria a reader could check against?
   - Push toward the goal: fewer open unknowns each pass, ideally an empty risk section at handoff.
4. Check the verses at value altitude:
   - Is each verse a usable vertical slice (passes the litmus test), or is it a horizontal layer / internal task that belongs in the plan?
   - Is the order delivering incremental value — walking skeleton first, riskiest assumptions early?
   - Is each named by the capability unlocked, not the component built?
5. Check that no verse has slid into implementation detail, **and that no verse is smuggling an open question** — a "TBD", a "deferred to plan-time", a "the technical question is…". The "touches" pointer is a locator only; flag anything prescribing *how*, and flag any unknown hiding in a verse: it belongs in Risks & unknowns, or answered in Decisions — never in the verse text.
6. Check the spikes: is every verse that exists to answer an unknown marked `(spike)`, and does its Risks entry state the question, the yes/no observation, what's downstream, and whether the artifact is kept? A spike missing the downstream call is the one bit_plan can't act on. And check the inverse — a verse marked `(spike)` whose question an inline probe could answer right now shouldn't be a verse; run the probe and write the Decision.
7. Check the References section: did the user provide any reference docs during this session that aren't captured there? If so, add them. If no references exist yet and none were provided, omit the section entirely.
8. Propose edits with reasoning — don't silently rewrite large sections. Confirm before rewriting more than a few lines.

The user drives this loop; keep refining with them until they're happy enough to move to bit_plan.

### Refining after a spike

When the user comes back with a spike's result, this is the most valuable refine pass there is — a real answer has arrived and the scope was written not knowing it. Three things happen, in order:

1. **The unknown becomes a Decision.** It's answered, so it leaves Risks entirely. Record what the answer actually was and what was observed, not just the conclusion — a future reader needs to know it was measured rather than assumed.
2. **The downstream verses get revised against the answer.** This is the point of the whole loop, so don't skim it. Reread each verse the risk entry named and ask what the answer changes: does the capability still make sense, is it still the right order, is a verse now unnecessary, does a new one need to exist? A spike that comes back "no" is a *successful* spike, and it usually means real rework here — that's the cost it was run to expose early, so take it rather than patching around it.
3. **The spike verse gets checked off** — bit_do normally does this when the last bar lands, so it may already be `[x]`. If the answer was "no" and the approach changed, the spike still delivered: it's done.

Then hand back to bit_plan for the verses that were waiting. Say explicitly which ones are now plannable, so the next planning pass doesn't have to re-derive it.

---

## Handoff to bit_plan

When the user is satisfied with the scope, the next step is bit_plan, which turns each verse into detailed TDD steps — one bar per step under this track. Give them the track ID and remind them, briefly, that:
- The plan's bars live under this track, so they inherit the WHY rather than repeating it.
- Each bar will be tagged (`--phase`/`--phase-label`) with the verse it serves, so progress rolls up.
- bit_do later moves each bar's status and rolls the track up — checking off a verse — as work lands.

If the scope has a spike verse, say what bit_plan will do with it so the handoff isn't a surprise: it plans the spikes, and then either stops there (if the risk entry says later verses depend on the answer) or asks whether to split the scope in two (if it says they don't). Either way the user comes back here with the result before the dependent verses get planned.

You don't build the plan — just point them at bit_plan, naming the track, when they're ready.
