---
name: bit_plan
description: Create or refine an implementation plan for a large task — bug fix, refactor, or new feature. Use when the user says "make a plan", "let's plan this out", "let's revise the plan", names a scope track to plan, or describes a change that is too large to implement in one session. Also triggers on casual phrasing like "let's think through this" or "how should we approach X" when the scope is clearly multi-step. This skill authors and refines the plan only — one bar (child task) per step under the scope's track in `.bit/`, through the `bp` CLI. When the user wants to frame the high-level WHY and delivery order first, use bit_scope; when they want to carry out an existing plan ("implement the plan", "continue our implementation", "do the next step"), use bit_do instead. Produces contradiction-driven step bars (each one red-green cycle and one commit) tagged with the verse they serve, TDD-first checklists, and an explicit split between what Claude verifies and what the user verifies. If the scope marks a verse `(spike)` — work whose deliverable is an answer to an open unknown — this skill plans every spike first and then either stops there or proposes splitting the scope, depending on whether the remaining verses depend on the answer.
---

# Implementation Plan Creator

You create and refine implementation plans. A plan is a set of **bars** (child tasks) under a scope's **track** in `.bit/`, authored through the `bp` CLI — one bar per step, detailed enough to work from autonomously across sessions, minimal enough not to waste tokens.

**Before you drive the CLI, run `bp instructions`** — the shared command contract (read the scope from its track body, create bars under the track, tag each bar's verse). Every write goes through `bp`; never hand-edit `.bit/tasks/*.md`.

## Two modes

**Create** — start from a scope (an existing track) and build its steps as bars from scratch
**Refine** — improve the existing bars under a track (add, split, reword, or re-order steps)

---

## Context — defer to the scope

The WHY lives in the **scope** — the track body (bit_scope's work), not here. Because the bars you create live *under* that track, the linkage is structural: a reader opening a bar reads the parent track (`bp task read <track> --body`) for the WHY. You don't repeat the motivation in each bar, and there's no separate Context pointer to maintain — the parent relationship is the pointer.

If **no scope track exists** — the user came straight to a plan — you still need the WHY before drafting. Either suggest writing a quick scope first (bit_scope creates the track), or capture 2–3 sentences of motivation and put them at the top of the track body yourself. A reader who knows nothing about the codebase should understand *why* this work is needed:

**Wrong:** "This plan covers fixing the prime distribution queries and cleaning up the streaming code."
**Right:** "Congressional and house district voter files show wrong prime totals because each county's ETL job overwrites the district-level aggregate — only the last county to finish is reflected. Voter targeting for non-county geographies is unreliable until this is fixed."

If the user gives you only a "what", push back once to get the "why". Ask: what breaks or fails today because this isn't done?

---

## Gathering context (new plans)

**Start from the scope track.** The user names it (by ID or title); read its body end to end with `bp task read <track> --body` — the WHY, the verses, the "touches" pointers, the risks, and the References section if one exists. If the scope has a `## References` section, note which reference docs are relevant to which verses before drafting bars — the right ones need to be carried forward into the right bar bodies, not left in the scope where they'll be out of context when bit_do executes the step. The scope hands you the delivery order and the code areas each verse affects — your job is to turn its verses into TDD steps, one bar each under that track.

Default to planning every verse in one pass. Splitting into multiple planning sessions exists to route around a genuine unknown, not as a default posture just because a scope has more than one verse. If the scope is clear and nothing is open, plan it end to end now; don't ask the user to pick a verse to be cautious.

### When the scope has spike verses

A **spike** is a verse whose deliverable is an *answer*, not a capability. bit_scope marks it `Verse N (spike)` in the checklist and, in the Risks & unknowns entry it resolves, states the question, what counts as a yes or a no, whether the artifact is kept, and — the field you need most — what's **downstream** of the answer. Read that before drafting any bars, because spikes change *how much of the scope you plan*, and getting that wrong wastes a whole cycle.

**Plan every spike.** More than one unknown means more than one spike, they're independent questions, and there's no reason to serialise learning. All of them get bars.

Then use the downstream note:

- **Later verses depend on the answer → plan the spikes and nothing else.** Don't ask about this one; it's the correct move, so make it and say so. Name the verses you're leaving unplanned, and state the loop back: run the spikes with bit_do, take the results to bit_scope, which turns the unknown into a Decision and revises those verses, then plan again. The reason to stop rather than plan optimistically is that a plan which gets rewritten is worse than no plan — it reads as settled, so nobody re-reads it, and its detailed steps quietly become the shape of the work even after the answer contradicted them.

- **Later verses stand on their own → this is a signal to split the scope, and you ask first.** A scope carrying both settled work and an open question is doing two independent jobs: the known work could be planned, built, and signed off without ever waiting on the answer, and holding it hostage to a spike is pure delay. So propose the split — one track for the knowns, one for the spikes and whatever follows from them — and say what it buys. But **confirm before doing it**: it creates a second track and changes what the user is tracking day to day, which is theirs to decide, not a mechanical consequence of the scope's shape. If they agree, hand back to bit_scope to author the split (it owns tracks; you own bars), then plan each track. If they'd rather keep one track, plan the spikes and the independent verses together in it and defer only what genuinely hangs on an answer.

If the scope has a spike but no downstream note, that's the one thing you can't work around — ask, or hand back to bit_scope to settle it. Guessing picks between "plan almost nothing" and "split the user's track," and both are expensive to be wrong about.

If there's no scope, ask before researching:

1. What's the problem or goal — user-facing impact, not technical description?
2. What triggered this now — bug report, wrong data, a deadline?
3. Any constraints — things we must not touch, production concerns, time pressure?

Then research the codebase (a scope's light "touches" pointers are a starting point, not a substitute — go deeper here):
- Find every file the change will touch
- Read adjacent code to understand the existing pattern before proposing changes
- Note specific function names, line numbers, and current vs. desired behavior
- Identify tests that will need updating

Also identify any verse that will need real-shaped data from an external system — a mock response, a poll condition, a citation, an example payload. The user is the actual authority on what real data looks like here, not the model's recollection, so work out when to loop them in:
- **Obtainable now** (an existing endpoint, an existing query, a doc you can point to) — ask the user for it before drafting, so the bar is written against something real instead of a plausible guess.
- **Only obtainable once part of the feature is built** (you can't capture a real response until the client wrapper this plan is about to create actually exists) — don't block drafting on it. Instead, give that bar a `**Needs real data:**` note (see Plan format) naming the actual mechanism to get it once the bar is reached — a command to run, an endpoint to hit, a query to make. bit_do won't have this section's context by the time it gets there, so the note has to name the real mechanism, not just flag that one exists.

Don't start drafting until you've done this research. A plan that names wrong files or misunderstands the existing pattern wastes more time than it saves.

---

## TDD-First — Outside-In, Always

This is non-negotiable. Every step follows strict Test-Driven Development:

### The cycle within each step

1. **Write the test first (RED)** — Write the smallest possible test from the highest level that exercises the behavior this step adds. The test must fail, and it must fail for the *correct reason* (e.g., "function doesn't exist" or "returns wrong value", not "import error in unrelated file").

2. **Implement the minimum to pass (GREEN)** — Write only enough production code to make the failing test pass. No speculative code, no "while we're here" additions.

3. **Add more tests (RED → GREEN)** — Error cases, edge cases, additional scenarios. Each follows the same red-green cycle.

4. **Refactor (REFACTOR)** — Once green, clean up duplication or structure. Tests stay green throughout.

### Outside-in via contradiction (the forcing function)

Start at the highest level — the entry point the user will call. Work inward, but never drop a level "because the plan says to." Drop a level because a test at the current level *can't pass* without real behavior below.

The mechanism is **contradiction**:

1. Write a test. Make it pass with a hardcoded return. Commit.
2. Write a *second* test with different inputs whose expected output contradicts the hardcoded value. This test *can't pass* without real logic.
3. Implement the minimum real logic to make both pass. That logic may call a lower layer — which starts the same cycle (hardcoded → contradicted → real).

Each step is a committable green state. The test is what pulls real code into existence — never a verse boundary. You never write a verse that says "now implement the child workflow" without a failing test that makes it impossible *not to*.

**Why this matters:** Plans that say "Verse 1: top level. Verse 2: next level down" produce isolated layers that were never forced to integrate. The contradiction approach means every layer exists because a higher-level test demanded it. If no test demands it, it doesn't get built (YAGNI).

### Realistic test data

The highest-level test should use **realistic data** — real-shaped inputs, realistic volumes, production-like values. The goal is to exercise the system the way it will actually be used, not with obviously fake placeholders.

When you can't use real data (external services, databases, auth), mock — but before writing mocks, **check what existing integration tests in the repo do**. Match their patterns for:
- How they set up test fixtures
- What level of realism they use in test data
- Whether they use testcontainers, in-memory fakes, or recorded responses
- Helper functions they provide for building realistic test objects

Don't invent a mocking approach when the repo already has one. Consistency with the existing test infrastructure matters more than theoretical purity.

Matching the existing pattern controls *shape* — it doesn't make the specific values true. Build the mock's field values, poll conditions, and citations from the real artifact gathered for this bar (the one the user supplied during gathering, or the one this bar's `**Needs real data:**` step produces) — not from recollection of what the response probably looks like.

### Describing tests in a plan

Every test listed in a plan needs four fields:

- **Behavior** — what system property this test proves, in plain language. The "so what."
- **Setup** — concrete inputs, mocks, preconditions. Specific values, not vague descriptions.
- **Assertions** — exact expected outputs.
- **Boundary** — name the specific input field or condition being exercised and say where in its valid range this case sits. For example: "`items` length == 0 — the lower bound; proves the loop handles zero iterations without panic." This includes: counts at 0 / 1 / N, values that must be positive tested at 0 or negative, boolean conditions tested in both states, inputs outside the valid range that the system must reject or coerce. Always anchor to a concrete input — "proves aggregation works" is just restating the behavior; "`CompanyIDs` count > 1 — exercises the multi-element aggregation path" is a boundary.

A test name and one-liner ("3 members, assert count is 3") looks obvious until implementation reveals ambiguity. These fields force you to articulate what's actually being protected, which is what the executor needs to write the *right* test.

```markdown
1. **Write test (RED):**
   - [ ] `TestThing_HappyPath`
     - **Behavior:** [what system property this proves]
     - **Setup:** [concrete inputs, mocks, preconditions]
     - **Assertions:** [exact expected outputs]
     - **Boundary:** [what range/constraint/invariant this exercises]
   - [ ] Confirm fails: [expected reason]

3. **More tests (RED → GREEN):**
   - [ ] `TestThing_ErrorCase`
     - **Behavior:** …
     - **Setup:** …
     - **Assertions:** …
     - **Boundary:** …
```

### The one exception: spike bars

A spike bar's deliverable is an *answer*, so there's nothing to test-drive. What's being proven is whether some mechanism outside your control behaves the way the approach assumes, and no test you write can settle that — a test asserting the answer you hope for is the same laundering as a decision wearing a checkbox (see Trap 1 below). It looks like verification and verifies nothing.

What replaces the red-green cycle is a **falsifiable observation**: state up front what you'd see if the answer is yes and what you'd see if it's no, exactly the way an ordinary bar states its expected failure reason. That's what stops a spike from wandering around and concluding "seems to work" — the failure mode spikes actually have. And a "no" is a real result the bar can succeed at producing, not a failed bar.

TDD still applies to whatever code the spike **keeps**. The scope says whether the artifact survives; carry that into the bar, because a probe you throw away is built as cheaply as possible while a probe whose artifact later verses build on is real code with a real test. Blurring the two is how throwaway code ends up load-bearing.

### Why this matters

- Catching the wrong failure reason means your test isn't testing what you think. This is a critical signal.
- "Test at the end" plans routinely get shipped without tests because "the code works and we're out of time." TDD-first makes this impossible.
- Outside-in ensures you design the API/interface before the internals. The test IS the first consumer.

---

## Step design

Call them "steps" not "verses" — a step is one red-green cycle that earns one commit. A plan typically has more steps than it would have had verses, because each step is smaller: one test + the minimum code to pass it.

Each step should:
- Be one red-green cycle (one new test or a contradicting test, then make it pass)
- Leave the system green and committable
- State what **forces** it — why does this step exist? Usually: "contradicts the hardcoded return from Step N" or "forces real implementation of the mocked layer"
- Follow YAGNI — don't add code for behavior that no test demands yet

Steps should not:
- Bundle multiple unrelated test scenarios into one step (split them — each earns its own commit)
- Include "while we're in here" cleanup
- Drop a layer without a test at the current layer that demands it

Name steps after what they prove, not what they touch.
Good: "Contradiction forces real fan-out"
Bad: "Implement child workflow"

**"Leave the system green" is a whole-module property, not a per-bar one.** A bar's own test passing doesn't mean the codebase still builds. Before sequencing bars that change a shared function's signature, check whether existing callers are tested through a fake/interface boundary or hit the real implementation directly. If they hit it directly, the steps are not actually independently green — merge them into one bar, or add a bar that introduces the boundary first.

### Tag each bar with its verse

Every bar carries the verse it serves as CLI metadata: `--phase <N>` (the verse number from the scope's checklist) and `--phase-label "<label>"` (the verse's short name — the flag keeps the name `phase`, the scope's slice is a verse). This metadata is what lets progress roll up: bit_do checks off a verse in the track body once all the bars tagged to it are done — so the `--phase` value, not any body text, is the source of truth for rollup. The bar body *also* opens by naming that verse (`## **Verse N**`) so a reader can place the step at a glance, but if the two ever disagree, trust `--phase`. A bar serves exactly one verse; if a step seems to span two, it's probably two bars. Create the bars in the scope's delivery order, so the walking skeleton lands first.

**Do not put rollup or status instructions in the bar bodies.** Naming the verse the bar serves in its opening line is fine — that's context for the reader, not a rollup instruction. But a bar body describes only its own step beyond that (the TDD cycle and checks): no "verse rollup" notes, no `**Status:**` lines — the bar's status *field* is the progress marker, and keeping the scope in sync is bit_do's job. Writing that into the body just burns tokens on what the executor already knows.

### Refactor steps

TDD is red-green-**refactor**. After accumulating 3–7 examples of a pattern, consider a refactor step. This isn't test-driven (no new failing test) — it's reshaping code while keeping tests green. Include it in the plan as a step with clear criteria for when to attempt it and what to look for (repeated structures, divergent copies of the same logic). If fewer than 3 examples exist, it's too early — leave it alone.

---

## Verification split

After the TDD cycle in each step, there are two kinds of additional check.

**Claude verifies** — deterministic, scriptable, run before the bar is handed over:
- Tests pass (`make test`, `go test ./...`, specific test file)
- Linter passes (`make lint`)
- Build succeeds (`make build`)
- Specific output assertion (e.g., count matches, format correct)

**User verifies** — a concrete manual check the automated tests can't make: something the
human *does* in the running system and *observes* a specific result. There's an action and a
pass/fail.

- "In `bp tui`, press `→`, then confirm `q`/`esc`/`ctrl+c` each quit and `ctrl+d` still scrolls."
- "`git status`: these files are added; the only diffs to `.claude/` are X."
- "Run `bp task list` against the real records — the 13 bars under BIT-2 stay in one column, nothing wraps."

The litmus test: *could the user pass or fail this by doing one thing and looking?* If there's
no action-and-observation — if you're asking them to **bless a choice** rather than **observe a
behavior** — it isn't a verification. See the three traps below.

Never put a judgment call in "Claude verifies." Never put an automatable check in "User
verifies." Not every bar needs a User-verify — a pure-plumbing bar the Claude-checks already
cover should say so ("none — deterministic") rather than manufacture one.

### Trap 1: a decision wearing a checkbox

If you catch yourself writing *"X is acceptable"*, *"reads naturally"*, *"feels right"*, *"is a
reasonable shape"*, *"is the right call — worth confirming"* — stop. That's not something the
user can pass or fail; it's a **choice**, and a choice is a scope Decision, not a plan
verification. It shows up because the scope never settled it, so the plan launders the open
question into a fake checkbox — "the `archive` command reads naturally next to `delete`" is a
naming call; "the refusal message reads clearly" is copy; neither is a check.

When this happens, **hand back to bit_scope** to record the choice as a Decision — its Decisions
section is exactly for naming, wording, data-shape, and strategy calls, and they double as
acceptance criteria. Once it's decided, the bar either needs no User-verify or gets a concrete
one that *observes* the now-settled behavior. Don't ship the launder.

A launder caught here *is* a decision the plan didn't make, so it's worth recording as well as
fixing. Offer that in one line alongside the hand-back — yes or no, and `/bit:feedback` writes the
note. Recording the Decision is what erases the trace, so the offer belongs before the hand-back,
not after.

### Trap 2: a "how does it feel" spread across every bar

Some checks are real but subjective — *does the whole slice integrate, does this capability feel
right end to end.* Legitimate, but they're about the **verse as a delivered capability**, not any
single bar. Scattering "does this feel natural" onto each bar drags a squishy check no one can
answer until the slice is whole — and manufactures busywork.

Put the verse's one integration/feel check on its **last bar** — the one that completes the verse
— phrased against the capability the verse unlocks ("Whole slice: archiving a done track moves it
and its bars out of the active view — the declutter goal actually lands"). Earlier bars carry
only their own concrete checks, or none.

### Trap 3: a real-shaped action that isn't real

An action-and-observation check can still be false — the command it names might not exist, or
might not do what the wording assumes. Passing the litmus test only confirms the *shape* is
right, not that the named action actually works.

Before writing a command the user is meant to run, confirm it actually works in this project —
check the task runner, README, or an already-correct bar elsewhere in the same track. If no
invocation convention exists yet for this new capability, that's a scope Decision, not something
to guess in the plan — hand back to bit_scope.

**Claude never commits.** The plan includes a suggested commit message per step, but committing is always the user's action.

---

## Plan format

Each step is **one bar** under the scope's track. The bar's **title** is the step name (what it proves); its `--phase`/`--phase-label` tag the verse it serves; its **body** is the step detail below. Create each bar in delivery order — `task create --parent` prints the dotted bar ID (`BIT-7.1`, `BIT-7.2`, …), and the order you create them in is the order bit_do will execute them:

```bash
BAR=$(bp task create "Contradiction forces real fan-out" \
        --parent "$TRACK" --phase 1 --phase-label "Ingest" \
        -d "$(cat step-body.md)")
```

Report the bar IDs (or just the count and the track) back to the user when you're done.

The **bar body** uses this structure. It opens with the verse the bar serves (`## **Verse N**`) so a reader knows the slice at a glance — the same verse the `--phase` metadata carries. After that: no `## Step N` heading (the title is the step name) and no `**Status:**` line (the status field is the marker). The step's sections are `##` headings so they stand apart when the body is read on its own:

```markdown
## **Verse 1**

[One sentence: what this step accomplishes and what forces it (e.g., "hardcoded return can't satisfy both tests")]

## Scope
- `path/to/file.go` — what changes here
- `path/to/other.go` — what changes here

## References
[Omit if no reference from the scope's References section applies to this step.
Only include references the implementer actually needs to read for this bar.]

- `path/to/doc.md` — what to use it for in this step

## Needs real data
[Omit unless this bar's test data can only be obtained once this bar's own work exists —
see Gathering context. Name the actual mechanism, not just that one is needed.]

- [ ] [a command to run, an endpoint to hit, a query to make]

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestName` (table-driven subtest if applicable)
     - **Behavior:** …
     - **Setup:** …
     - **Assertions:** …
     - **Boundary:** …
   - [ ] Confirm fails: [expected failure reason]

2. **Implement (GREEN):**
   - [ ] specific implementation task (may be a hardcoded return — that's fine for the first bar)

## Claude verifies
- [ ] tests pass (use the project's task runner — check CLAUDE.md or Makefile/justfile/etc.)
- [ ] linter passes

## User verifies
- [ ] [concrete manual check — do X, observe Y; omit on a pure-plumbing bar. A verse's integration/feel check goes on its *last* bar. Never a decision-in-disguise ("reads naturally", "is acceptable") — that's a scope Decision, hand it back.]

## Commit (user)
`feat(scope): short description`
```

Before finalizing a bar, trace its RED test's assertions against its GREEN checklist line by
line. If the checklist, read literally, would still leave the test failing, the checklist is
incomplete — "a test at this level forces it" is not a substitute for checking this specific
bar.

The throughline that used to live in a plan's "How this plan works" section — what the entry point is and how tests drive deeper — belongs in the **track body** (a sentence or two), not repeated per bar. If it's missing and would help, offer to add it to the track via bit_scope.

### A spike bar's body

Same frame, different middle: the TDD cycle is replaced by the question and how it gets settled. Prefix the phase label with `spike:` so it's visible wherever bars are listed — `--phase 1 --phase-label "spike: delivery"` — since a reader scanning the board should be able to tell which bars produce knowledge and which produce capability.

```markdown
## **Verse 1 (spike)**

**Question:** [the unknown, phrased so an answer is recognisable]
**Yes looks like:** [the concrete observation]
**No looks like:** [the concrete observation — a real outcome, not a failed bar]

## Scope
- `path/to/thing` — what gets built, and whether it's kept or thrown away

## Method
- [ ] the concrete steps that produce the observation

## Claude verifies
- [ ] [whatever is deterministic — a command exits 0, a file contains X. Plus tests, if this bar keeps code.]

## User verifies
- [ ] [the observation itself, when only a human can see it]

## Report back
- [ ] Take the answer to bit_scope: the unknown becomes a Decision, and Verses N–M get revised against it before they're planned.

## Commit (user)
`<type>(scope): short description`
```

The `Report back` item is what closes the loop — without it a spike's answer lives only in the session that ran it, and the scope keeps its unknown while the work has moved on.

---

## Refining an existing plan

1. Read the whole plan first: the track body (`task read <track> --body`) and every bar in order (`task list --parent <track>`, then `task read <bar> --body` for each).
2. Compute what this change actually invalidates — don't just edit what you came to touch. If a bar being edited already contains a factual/architectural claim (a deployment mechanism, a build choice), verify it against the track's current Decisions and the codebase's actual precedent before keeping it — inherited text isn't pre-vetted just because it's already there. If the track's Decisions changed, list every bar in the track (`bp task list --parent <track>`), not just the ones this pass is touching, and reset to `todo` any non-`done` bar that no longer matches.
3. Check the verse tags: is each bar tagged to a verse (`--phase`/`--phase-label`), and do the tags follow the scope's delivery order? An untagged bar, or one that jumps ahead of the walking skeleton, is a flag.
4. Check that the bar bodies don't duplicate the WHY the track owns, or carry stray `**Status:**` lines — that's drift waiting to happen.
5. Check the throughline: can you trace *why* each bar exists? Every bar after the first should be forced by a contradiction or dependency. If a bar says "now implement X" without a test that demands it, flag it — something is missing.
6. Review each bar: does it start with a test? Is it one red-green cycle?
7. Flag any bar that bundles multiple scenarios (split it into two bars — each earns its own commit)
8. Check each bar's **User verifies** against the three traps (see Verification split): a decision-in-disguise ("reads naturally", "is acceptable", "feels right") is a missing scope Decision — flag it and hand back to bit_scope; a subjective "how does it feel" on a non-final bar belongs on its verse's last bar instead; a command named without confirming it actually works in this project is a real-shaped-but-not-real check — verify it or hand back to bit_scope if no invocation convention exists yet. A concrete do-X-observe-Y check confirmed against this project is fine as is.
9. Check references: if the scope track has a `## References` section, are the relevant references carried into bars that need them? A bar touching a verse that depends on a spec or design doc should have a `## References` section pointing to it. Flag any bar where the reference is missing and the implementer would need it.
10. Check the spikes both ways. Does each spike bar state a falsifiable observation (yes looks like / no looks like), say whether its artifact is kept, and end with a `Report back` item? And the drift case that matters more: are there bars planned for verses the scope says depend on an *unresolved* spike? Those were written without the answer, so flag them — the fix is to hold them until the spike reports back, not to polish them.
11. Flag YAGNI violations, over-bundled bars, or missing verification
12. Apply edits through the CLI: reword a bar with `task update <bar> -d "…"`, add a missing step with `task create --parent …`, and **reorder a misplaced bar with `task move <bar> --before|--after <sibling>`** — the ID is stable identity, so moving a bar rewrites only the track's order list and every reference to that bar (commit messages, notes) stays valid. Don't delete-and-recreate to reorder; that churns IDs and breaks those references. Propose changes with reasoning — don't silently rewrite large sections.

Confirm with the user before rewriting more than a few bars.

Once the edits are agreed, offer to record what drove them: a bar reworded, split, or reordered
because it was wrong is the same plan gap seen after the fact. One line, yes or no, and
`/bit:feedback` writes the note. A refine pass that only tightened wording needs no offer — this is
for the passes where something was actually wrong.

---

## Detail level

Plans are executed by Claude Code with minimal back-and-forth. That means:
- Name the exact files and functions being changed
- For type/interface changes, note the downstream callers that need updating
- Include the exact test/build command for each verification step
- Show the expected test failure reason so the executor knows they're on track

Don't pad. If a step is simple, the checklist can be two items. If a verification is just "tests pass", say that.
