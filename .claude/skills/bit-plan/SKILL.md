---
name: bit-plan
description: Create or refine an implementation plan for a large task — bug fix, refactor, or new feature. Use when the user says "make a plan", "let's plan this out", "let's revise the plan", references a .md plan file to write or improve, or describes a change that is too large to implement in one session. Also triggers on casual phrasing like "let's think through this" or "how should we approach X" when the scope is clearly multi-step. This skill authors and refines the plan document only — when the user wants to frame the high-level WHY and delivery order first, use bit-scope; when they want to carry out an existing plan ("implement the plan", "continue our implementation", "do the next step"), use bit-do instead. Produces a structured markdown plan that links back to its scope (bit-scope) for the WHY, with contradiction-driven steps (each one red-green cycle and one commit) tagged to the scope's phases, TDD-first checklists, and an explicit split between what Claude verifies and what the user verifies.
---

# Implementation Plan Creator

You create and refine implementation plans. Plans are markdown files that serve as the execution guide across multiple Claude Code sessions — detailed enough to work from autonomously, minimal enough not to waste tokens.

## Two modes

**Create** — start from a problem description, build the plan from scratch
**Refine** — improve an existing plan (user pastes it or points to a file with @)

---

## Context — defer to the scope

The WHY lives in the **scope** (bit-scope's `<feature>-scope.md`), not here. Duplicating motivation in two documents just lets them drift apart. So the plan's Context section is a **one-line pointer** to the scope, optionally with a single-sentence recap for a reader who opens the plan alone:

```markdown
## Context
See scope: [feature-scope.md](./feature-scope.md)
Recap: <one sentence, optional>
```

If **no scope exists** — the user came straight to a plan — you still need the WHY before drafting. Either suggest writing a quick scope first (bit-scope), or capture 2–3 sentences of motivation inline. A reader who knows nothing about the codebase should understand *why* this work is needed:

**Wrong:** "This plan covers fixing the prime distribution queries and cleaning up the streaming code."
**Right:** "Congressional and house district voter files show wrong prime totals because each county's ETL job overwrites the district-level aggregate — only the last county to finish is reflected. Voter targeting for non-county geographies is unreliable until this is fixed."

If the user gives you only a "what", push back once to get the "why". Ask: what breaks or fails today because this isn't done?

---

## Gathering context (new plans)

**If a scope doc exists, start there.** Read it end to end: the WHY, the phases, the "touches" pointers, and the risks. The scope hands you the delivery order and the code areas each phase affects — your job is to turn its phases into TDD steps. Confirm with the user which phase(s) this planning session covers; you don't have to plan the whole scope at once.

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

Each step is a committable green state. The test is what pulls real code into existence — never a plan phase boundary. You never write a phase that says "now implement the child workflow" without a failing test that makes it impossible *not to*.

**Why this matters:** Plans that say "Phase 1: top level. Phase 2: next level down" produce isolated layers that were never forced to integrate. The contradiction approach means every layer exists because a higher-level test demanded it. If no test demands it, it doesn't get built (YAGNI).

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

Call them "steps" not "phases" — a step is one red-green cycle that earns one commit. A plan typically has more steps than it would have had phases, because each step is smaller: one test + the minimum code to pass it.

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

### Tag each step with its scope phase

When a scope exists, every step names the scope phase it serves — e.g. `## Step 3 (Phase 1 — Ingest) — Contradiction forces real fan-out`. This is what lets progress roll up: bit-do checks off a scope phase once all the steps tagged to it are done. A step serves exactly one phase; if it seems to span two, it's probably two steps. Plan the phases in the scope's delivery order, so the walking skeleton lands first.

**Do not include scope-update instructions in the plan.** The plan's job ends at describing the steps — no "Scope phase rollup" sections, no reminders to check off scope phases, no guidance on updating other documents. Tracking progress and keeping the scope in sync is bit-do's responsibility; including it here just burns output tokens on information the executor already has from the skill that runs it.

### Refactor steps

TDD is red-green-**refactor**. After accumulating 3–7 examples of a pattern, consider a refactor step. This isn't test-driven (no new failing test) — it's reshaping code while keeping tests green. Include it in the plan as a step with clear criteria for when to attempt it and what to look for (repeated structures, divergent copies of the same logic). If fewer than 3 examples exist, it's too early — leave it alone.

---

## Verification split

After the TDD cycle in each step, there are two types of additional checks:

**Claude verifies** — deterministic, scriptable:
- Tests pass (`make test`, `go test ./...`, specific test file)
- Linter passes (`make lint`)
- Build succeeds (`make build`)
- Specific output assertion (e.g., count matches, format correct)

**User verifies** — judgment calls that require human eyes:
- Business logic makes sense for real data
- API design feels right
- Safe to commit or deploy
- Approach aligns with team conventions

Never put a judgment call in "Claude verifies." Never put an automatable check in "User verifies."

**Claude never commits.** The plan includes a suggested commit message per step, but committing is always the user's action.

---

## Plan format

Write plans to a `.md` file in the project root unless the user specifies otherwise. Use this structure:

```markdown
# [Plan title]

## Context
See scope: [feature-scope.md](./feature-scope.md) — the WHY and phase order live there.
[Optional one-sentence recap. If there's no scope, put 2–4 sentences of motivation here instead.]

## How this plan works
[Brief explanation of the throughline — what's the entry point, and how tests drive us deeper.]

---

## Step 1 (Phase N — [scope phase]) — [Name: what this proves]
[One sentence: what this step accomplishes and what forces it (e.g., "hardcoded return can't satisfy both tests")]

**Scope:**
- `path/to/file.go` — what changes here
- `path/to/other.go` — what changes here

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestName` (table-driven subtest if applicable)
     - **Behavior:** …
     - **Setup:** …
     - **Assertions:** …
     - **Boundary:** …
   - [ ] Confirm fails: [expected failure reason]

2. **Implement (GREEN):**
   - [ ] specific implementation task (may be a hardcoded return — that's fine for Step 1)

**Claude verifies:**
- [ ] tests pass (use the project's task runner — check CLAUDE.md or Makefile/justfile/etc.)
- [ ] linter passes

**User verifies:**
- [ ] [judgment call, if any — not every step needs one]

**Commit (user):** `feat(scope): short description`

---

## Step 2 (Phase N — [scope phase]) — [Name: contradiction forces real behavior]
[What contradicts Step 1's hardcoded return, and what real behavior this forces.]

**TDD cycle:**

1. **Write test (RED):**
   - [ ] Add subtest with contradicting inputs…
   - [ ] Confirm fails: [hardcoded return gives X, test expects Y]

2. **Implement (GREEN):**
   - [ ] Replace hardcoded return with real logic…

...
```

---

## Refining an existing plan

1. Read the full plan before commenting
2. Check the context: does it point to a scope for the WHY (or, absent a scope, say why not what)? If it duplicates a long WHY the scope already owns, flag it — that's drift waiting to happen.
3. Check the phase tags: is each step tagged to a scope phase, and do the tags follow the scope's delivery order? An untagged step, or one that jumps ahead of the walking skeleton, is a flag.
4. Check the throughline: can you trace *why* each step exists? Every step after Step 1 should be forced by a contradiction or dependency. If a step says "now implement X" without a test that demands it, flag it — something is missing.
5. Review each step: does it start with a test? Is it one red-green cycle?
6. Flag any step that bundles multiple scenarios (split them — each earns its own commit)
7. Flag YAGNI violations, over-bundled steps, or missing verification
8. Propose edits with reasoning — don't silently rewrite large sections

Confirm with the user before rewriting more than a few lines.

---

## Detail level

Plans are executed by Claude Code with minimal back-and-forth. That means:
- Name the exact files and functions being changed
- For type/interface changes, note the downstream callers that need updating
- Include the exact test/build command for each verification step
- Show the expected test failure reason so the executor knows they're on track

Don't pad. If a step is simple, the checklist can be two items. If a verification is just "tests pass", say that.
