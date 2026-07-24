---
name: bit_plan
description: Create or refine an implementation plan for a large task — bug fix, refactor, or new feature. Use when the user says "make a plan", "let's plan this out", "let's revise the plan", names a scope track to plan, or describes a change that is too large to implement in one session. Also triggers on casual phrasing like "let's think through this" or "how should we approach X" when the scope is clearly multi-step. This skill authors and refines the plan only — one bar (child task) per step under the scope's track in `.bit/`, through the `bit` CLI. When the user wants to frame the high-level WHY and delivery order first, use bit_scope; when they want to carry out an existing plan ("implement the plan", "continue our implementation", "do the next step"), use bit_do instead. Produces contradiction-driven step bars (each one red-green cycle and one commit) tagged with the verse they serve, TDD-first checklists, and an explicit split between what Claude verifies and what the user verifies.
---

# Implementation Plan Creator

You create and refine implementation plans. A plan is a set of **bars** (child tasks) under a scope's **track** in `.bit/`, authored through the `bit` CLI — one bar per step, detailed enough to work from autonomously across sessions, minimal enough not to waste tokens.

**Before you drive the CLI, read `.claude/bit-cli.md`** — the shared command contract (read the scope from its track body, create bars under the track, tag each bar's verse). Every write goes through `bit`; never hand-edit `.bit/tasks/*.md`.

## Two modes

**Create** — start from a scope (an existing track) and build its steps as bars from scratch
**Refine** — improve the existing bars under a track (add, split, reword, or re-order steps)

---

## Context — defer to the scope

The WHY lives in the **scope** — the track body (bit_scope's work), not here. Because the bars you create live *under* that track, the linkage is structural: a reader opening a bar reads the parent track (`bit task read <track> --body`) for the WHY. You don't repeat the motivation in each bar, and there's no separate Context pointer to maintain — the parent relationship is the pointer.

If **no scope track exists** — the user came straight to a plan — you still need the WHY before drafting. Either suggest writing a quick scope first (bit_scope creates the track), or capture 2–3 sentences of motivation and put them at the top of the track body yourself. A reader who knows nothing about the codebase should understand *why* this work is needed:

**Wrong:** "This plan covers fixing the prime distribution queries and cleaning up the streaming code."
**Right:** "Congressional and house district voter files show wrong prime totals because each county's ETL job overwrites the district-level aggregate — only the last county to finish is reflected. Voter targeting for non-county geographies is unreliable until this is fixed."

If the user gives you only a "what", push back once to get the "why". Ask: what breaks or fails today because this isn't done?

---

## Gathering context (new plans)

**Start from the scope track.** The user names it (by ID or title); read its body end to end with `bit task read <track> --body` — the WHY, the verses, the "touches" pointers, and the risks. The scope hands you the delivery order and the code areas each verse affects — your job is to turn its verses into TDD steps, one bar each under that track.

Default to planning every verse in one pass. Splitting into multiple planning sessions exists to route around a genuine unknown — a risk the scope flagged "de-risk before planning? Yes", or a later verse whose shape depends on what an earlier verse turns out to build — not as a default posture just because a scope has more than one verse. If the scope is clear and none of its risks block a verse, plan it end to end now; don't ask the user to pick a verse to be cautious. If an unknown does block a later verse, plan up through whatever isn't blocked, then tell the user which verse(s) you're deferring and why.

If there's no scope, ask before researching:

1. What's the problem or goal — user-facing impact, not technical description?
2. What triggered this now — bug report, wrong data, a deadline?
3. Any constraints — things we must not touch, production concerns, time pressure?

Then research the codebase (a scope's light "touches" pointers are a starting point, not a substitute — go deeper here):
- Find every file the change will touch
- Read adjacent code to understand the existing pattern before proposing changes
- Note specific function names, line numbers, and current vs. desired behavior
- Identify tests that will need updating

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

- "In `bit tui`, press `→`, then confirm `q`/`esc`/`ctrl+c` each quit and `ctrl+d` still scrolls."
- "`git status`: these files are added; the only diffs to `.claude/` are X."
- "Run `bit task list` against the real records — the 13 bars under BIT-2 stay in one column, nothing wraps."

The litmus test: *could the user pass or fail this by doing one thing and looking?* If there's
no action-and-observation — if you're asking them to **bless a choice** rather than **observe a
behavior** — it isn't a verification. See the two traps below.

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

### Trap 2: a "how does it feel" spread across every bar

Some checks are real but subjective — *does the whole slice integrate, does this capability feel
right end to end.* Legitimate, but they're about the **verse as a delivered capability**, not any
single bar. Scattering "does this feel natural" onto each bar drags a squishy check no one can
answer until the slice is whole — and manufactures busywork.

Put the verse's one integration/feel check on its **last bar** — the one that completes the verse
— phrased against the capability the verse unlocks ("Whole slice: archiving a done track moves it
and its bars out of the active view — the declutter goal actually lands"). Earlier bars carry
only their own concrete checks, or none.

**Claude never commits.** The plan includes a suggested commit message per step, but committing is always the user's action.

---

## Plan format

Each step is **one bar** under the scope's track. The bar's **title** is the step name (what it proves); its `--phase`/`--phase-label` tag the verse it serves; its **body** is the step detail below. Create each bar in delivery order — `task create --parent` prints the dotted bar ID (`BIT-7.1`, `BIT-7.2`, …), and the order you create them in is the order bit_do will execute them:

```bash
BAR=$(bit task create "Contradiction forces real fan-out" \
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

The throughline that used to live in a plan's "How this plan works" section — what the entry point is and how tests drive deeper — belongs in the **track body** (a sentence or two), not repeated per bar. If it's missing and would help, offer to add it to the track via bit_scope.

---

## Refining an existing plan

1. Read the whole plan first: the track body (`task read <track> --body`) and every bar in order (`task list --parent <track>`, then `task read <bar> --body` for each).
2. Check the verse tags: is each bar tagged to a verse (`--phase`/`--phase-label`), and do the tags follow the scope's delivery order? An untagged bar, or one that jumps ahead of the walking skeleton, is a flag.
3. Check that the bar bodies don't duplicate the WHY the track owns, or carry stray `**Status:**` lines — that's drift waiting to happen.
4. Check the throughline: can you trace *why* each bar exists? Every bar after the first should be forced by a contradiction or dependency. If a bar says "now implement X" without a test that demands it, flag it — something is missing.
5. Review each bar: does it start with a test? Is it one red-green cycle?
6. Flag any bar that bundles multiple scenarios (split it into two bars — each earns its own commit)
7. Check each bar's **User verifies** against the two traps (see Verification split): a decision-in-disguise ("reads naturally", "is acceptable", "feels right") is a missing scope Decision — flag it and hand back to bit_scope; a subjective "how does it feel" on a non-final bar belongs on its verse's last bar instead. A concrete do-X-observe-Y check is fine as is.
8. Flag YAGNI violations, over-bundled bars, or missing verification
9. Apply edits through the CLI: reword a bar with `task update <bar> -d "…"`, add a missing step with `task create --parent …`, and **reorder a misplaced bar with `task move <bar> --before|--after <sibling>`** — the ID is stable identity, so moving a bar rewrites only the track's order list and every reference to that bar (commit messages, notes) stays valid. Don't delete-and-recreate to reorder; that churns IDs and breaks those references. Propose changes with reasoning — don't silently rewrite large sections.

Confirm with the user before rewriting more than a few bars.

---

## Detail level

Plans are executed by Claude Code with minimal back-and-forth. That means:
- Name the exact files and functions being changed
- For type/interface changes, note the downstream callers that need updating
- Include the exact test/build command for each verification step
- Show the expected test failure reason so the executor knows they're on track

Don't pad. If a step is simple, the checklist can be two items. If a verification is just "tests pass", say that.
